/*
Copyright 2021 The Kruise Authors.
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rollout

import (
	"fmt"

	rolloutsapi "github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	internalapi "github.com/openkruise/kruise-tools/pkg/api"
	internalpolymorphichelpers "github.com/openkruise/kruise-tools/pkg/internal/polymorphichelpers"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// UndoOptions is the start of the data required to perform the operation.  As new fields are added, add them here instead of
// referencing the cmd.Flags()
type UndoOptions struct {
	PrintFlags *genericclioptions.PrintFlags
	ToPrinter  func(string) (printers.ResourcePrinter, error)

	Builder          func() *resource.Builder
	ToRevision       int64
	DryRunStrategy   cmdutil.DryRunStrategy
	Resources        []string
	Namespace        string
	EnforceNamespace bool
	RESTClientGetter genericclioptions.RESTClientGetter

	resource.FilenameOptions
	genericclioptions.IOStreams
}

var (
	undoLong = templates.LongDesc(`
		Rollback to a previous rollout.`)

	undoExample = templates.Examples(`
		# Rollback to the previous cloneset
		kubectl-kruise rollout undo cloneset/abc

		# Rollback to the previous Advanced StatefulSet
		kubectl-kruise rollout undo asts/abc

		# Rollback to daemonset revision 3
		kubectl-kruise rollout undo daemonset/abc --to-revision=3

		# Rollback to the previous deployment with dry-run
		kubectl-kruise rollout undo --dry-run=server deployment/abc
		
		# Rollback to workload via rollout api object
		kubectl-kruise rollout undo rollout/abc`)
)

// NewRolloutUndoOptions returns an initialized UndoOptions instance
func NewRolloutUndoOptions(streams genericclioptions.IOStreams) *UndoOptions {
	return &UndoOptions{
		PrintFlags: genericclioptions.NewPrintFlags("rolled back").WithTypeSetter(internalapi.GetScheme()),
		IOStreams:  streams,
		ToRevision: int64(0),
	}
}

// NewCmdRolloutUndo returns a Command instance for the 'rollout undo' sub command
func NewCmdRolloutUndo(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewRolloutUndoOptions(streams)

	validArgs := []string{"deployment", "daemonset", "statefulset", "cloneset", "advanced statefulset", "rollout"}

	cmd := &cobra.Command{
		Use:                   "undo (TYPE NAME | TYPE/NAME) [flags]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Undo a previous rollout"),
		Long:                  undoLong,
		Example:               undoExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.RunUndo())
		},
		ValidArgs: validArgs,
	}

	cmd.Flags().Int64Var(&o.ToRevision, "to-revision", o.ToRevision, "The revision to rollback to. Default to 0 (last revision).")
	usage := "identifying the resource to get from a server."
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
	cmdutil.AddDryRunFlag(cmd)
	o.PrintFlags.AddFlags(cmd)
	return cmd
}

// Complete completes all the required options
func (o *UndoOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.Resources = args
	var err error
	o.DryRunStrategy, err = cmdutil.GetDryRunStrategy(cmd)
	if err != nil {
		return err
	}

	if o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace(); err != nil {
		return err
	}

	o.ToPrinter = func(operation string) (printers.ResourcePrinter, error) {
		o.PrintFlags.NamePrintFlags.Operation = operation
		cmdutil.PrintFlagsWithDryRunStrategy(o.PrintFlags, o.DryRunStrategy)
		return o.PrintFlags.ToPrinter()
	}

	o.RESTClientGetter = f
	o.Builder = f.NewBuilder

	return err
}

func (o *UndoOptions) Validate() error {
	if len(o.Resources) == 0 && cmdutil.IsFilenameSliceEmpty(o.Filenames, o.Kustomize) {
		return fmt.Errorf("required resource not specified")
	}
	return nil
}

// RunUndo performs the execution of 'rollout undo' sub command
func (o *UndoOptions) RunUndo() error {
	r := o.Builder().
		WithScheme(internalapi.GetScheme(), scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(o.Namespace).DefaultNamespace().
		FilenameParam(o.EnforceNamespace, &o.FilenameOptions).
		ResourceTypeOrNameArgs(true, o.Resources...).
		ContinueOnError().
		Latest().
		Flatten().Do()
	if err := r.Err(); err != nil {
		return err
	}

	// perform undo logic here
	undoFunc := func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}
		rollbacker, err := internalpolymorphichelpers.RollbackerFn(o.RESTClientGetter, info.ResourceMapping())
		if err != nil {
			return err
		}

		result, err := rollbacker.Rollback(info.Object, nil, o.ToRevision, o.DryRunStrategy)
		if err != nil {
			return err
		}

		printer, err := o.ToPrinter(result)
		if err != nil {
			return err
		}

		return printer.PrintObj(info.Object, o.Out)
	}

	var refResources []string
	// deduplication: If a rollout arg references a workload which is also specified as an arg in the same command,
	// performing multiple undo operations on the workload within a single command is not smart. Such an action could
	// lead to confusion and yield unintended consequences. Therefore, undo operations in this context are disallowed.
	// Should such a scenario occur, only the first argument that points to the workload will be executed.
	deDuplica := make(map[string]struct{})

	err := r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		if info.Mapping.GroupVersionKind.Group == "rollouts.kruise.io" && info.Mapping.GroupVersionKind.Kind == "Rollout" {
			obj := info.Object
			if obj == nil {
				return fmt.Errorf("Rollout object not found")
			}
			ro, ok := obj.(*rolloutsapi.Rollout)
			if !ok {
				return fmt.Errorf("unsupported version of Rollout")
			}
			workloadRef := ro.Spec.WorkloadRef
			gv, err := schema.ParseGroupVersion(workloadRef.APIVersion)
			if err != nil {
				return err
			}
			deDuplicaKey := workloadRef.Kind + "." + gv.Version + "." + gv.Group + "/" + workloadRef.Name
			if _, ok := deDuplica[deDuplicaKey]; ok {
				return nil
			}
			deDuplica[deDuplicaKey] = struct{}{}
			refResources = append(refResources, deDuplicaKey)
			return nil
		}
		gvk := info.Mapping.GroupVersionKind
		deDuplicaKey := gvk.Kind + "." + gvk.Version + "." + gvk.Group + "/" + info.Name
		if _, ok := deDuplica[deDuplicaKey]; ok {
			return nil
		}
		deDuplica[deDuplicaKey] = struct{}{}
		return undoFunc(info, nil)
	})

	if len(refResources) < 1 {
		return err
	}

	var aggErrs []error
	aggErrs = append(aggErrs, err)
	r2 := o.Builder().
		WithScheme(internalapi.GetScheme(), scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(o.Namespace).DefaultNamespace().
		FilenameParam(o.EnforceNamespace, &o.FilenameOptions).
		ResourceTypeOrNameArgs(true, refResources...).
		ContinueOnError().
		Latest().
		Flatten().Do()

	if err = r2.Err(); err != nil {
		aggErrs = append(aggErrs, err)
		return errors.NewAggregate(aggErrs)
	}
	err = r2.Visit(undoFunc)
	aggErrs = append(aggErrs, err)
	return errors.NewAggregate(aggErrs)
}
