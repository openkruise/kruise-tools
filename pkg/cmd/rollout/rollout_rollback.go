/*
Copyright 2025 The Kruise Authors.

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

	internalapi "github.com/openkruise/kruise-tools/pkg/api"
	"github.com/openkruise/kruise-tools/pkg/cmd/util"
	internalpolymorphichelpers "github.com/openkruise/kruise-tools/pkg/internal/polymorphichelpers"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/set"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// RollbackOptions is the start of the data required to perform the operation.  As new fields are added, add them here instead of
// referencing the cmd.Flags()
type RollbackOptions struct {
	PrintFlags *genericclioptions.PrintFlags
	ToPrinter  func(string) (printers.ResourcePrinter, error)

	Resources []string

	Builder          func() *resource.Builder
	RollbackFunc     internalpolymorphichelpers.RolloutRollbackFuncGetter
	Namespace        string
	EnforceNamespace bool

	resource.FilenameOptions
	genericclioptions.IOStreams
}

var (
	rollbackLong = templates.LongDesc(`
		Rollback a rollout to a previous step during release progress.
		
		**CAUTION** This command only works on rollouts version >= 0.6.0

		You can rollback to a specific previous step using the --step flag. 
		If not specified, kubectl-kruise will look for a previous step that has no traffic and the most replicas to perform the rollback.`)

	rollbackExample = templates.Examples(`
		# Rollback a kruise rollout named "rollout-demo" in "ns-demo" namespace
		
		kubectl-kruise rollout rollback rollout/rollout-demo -n ns-demo

		# Rollback a kruise rollout resource to a specific step
		
		kubectl-kruise rollout rollback rollout/rollout-demo --step 1`)

	targetStep int32
)

// NewRollbackOptions returns an initialized RollbackOptions instance
func NewRollbackOptions(streams genericclioptions.IOStreams) *RollbackOptions {
	return &RollbackOptions{
		PrintFlags: genericclioptions.NewPrintFlags("rolled back").WithTypeSetter(internalapi.GetScheme()),
		IOStreams:  streams,
	}
}

// NewCmdRolloutRollback returns a Command instance for 'rollout rollback' sub command
func NewCmdRolloutRollback(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewRollbackOptions(streams)

	cmd := &cobra.Command{
		Use:                   "rollback RESOURCE",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Rollback a resource"),
		Long:                  rollbackLong,
		Example:               rollbackExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.RunRollback())
		},
	}

	usage := "identifying the resource to get from a server."
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
	o.PrintFlags.AddFlags(cmd)
	cmd.Flags().Int32Var(&targetStep, "step", -1, "Rollback to a specific previous step")
	return cmd
}

// Complete completes all the required options
func (o *RollbackOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.Resources = args

	o.RollbackFunc = internalpolymorphichelpers.RolloutRollbackGetter

	var err error
	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.ToPrinter = func(operation string) (printers.ResourcePrinter, error) {
		o.PrintFlags.NamePrintFlags.Operation = operation
		return o.PrintFlags.ToPrinter()
	}

	o.Builder = f.NewBuilder

	return nil
}

func (o *RollbackOptions) Validate() error {
	if len(o.Resources) == 0 && cmdutil.IsFilenameSliceEmpty(o.Filenames, o.Kustomize) {
		return fmt.Errorf("required resource not specified")
	}
	return nil
}

// RunRollback performs the execution of 'rollout rollback' sub command
func (o *RollbackOptions) RunRollback() error {
	r := o.Builder().
		WithScheme(internalapi.GetScheme(), scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(o.Namespace).DefaultNamespace().
		FilenameParam(o.EnforceNamespace, &o.FilenameOptions).
		ResourceTypeOrNameArgs(true, o.Resources...).
		ContinueOnError().
		Latest().
		Flatten().
		Do()
	if err := r.Err(); err != nil {
		return err
	}

	allErrs := []error{}
	infos, err := r.Infos()
	if err != nil {
		// restore previous command behavior where
		// an error caused by retrieving infos due to
		// at least a single broken object did not result
		// in an immediate return, but rather an overall
		// aggregation of errors.
		allErrs = append(allErrs, err)
	}

	for _, patch := range set.CalculatePatches(infos, scheme.DefaultJSONEncoder(), o.RollbackFunc(targetStep)) {
		info := patch.Info

		if patch.Err != nil {
			resourceString := info.Mapping.Resource.Resource
			if len(info.Mapping.Resource.Group) > 0 {
				resourceString = resourceString + "." + info.Mapping.Resource.Group
			}
			allErrs = append(allErrs, fmt.Errorf("error: %s %q %v", resourceString, info.Name, patch.Err))
			continue
		}

		if string(patch.Patch) == "{}" || len(patch.Patch) == 0 {
			printer, err := o.ToPrinter("already rolled back")
			if err != nil {
				allErrs = append(allErrs, err)
				continue
			}
			if err = printer.PrintObj(info.Object, o.Out); err != nil {
				allErrs = append(allErrs, err)
			}
			continue
		}

		obj, err := util.PatchSubResource(info.Client, info.Mapping.Resource.Resource, "status", info.Namespace, info.Name, info.Namespaced(), types.MergePatchType, patch.Patch, nil)
		if err != nil {
			allErrs = append(allErrs, fmt.Errorf("failed to patch: %v", err))
			continue
		}

		info.Refresh(obj, true)
		printer, err := o.ToPrinter("rolled back" +
			"")
		if err != nil {
			allErrs = append(allErrs, err)
			continue
		}
		if err = printer.PrintObj(info.Object, o.Out); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	return utilerrors.NewAggregate(allErrs)
}
