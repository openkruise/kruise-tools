/*
Copyright 2021 The Kruise Authors.

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
	"io"

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

	internalapi "github.com/openkruise/kruise-tools/pkg/api"
	"github.com/openkruise/kruise-tools/pkg/cmd/util"
	internalpolymorphichelpers "github.com/openkruise/kruise-tools/pkg/internal/polymorphichelpers"
)

// ApproveOptions is the start of the data required to perform the operation.  As new fields are added, add them here instead of
// referencing the cmd.Flags()
type ApproveOptions struct {
	PrintFlags *genericclioptions.PrintFlags
	ToPrinter  func(string) (printers.ResourcePrinter, error)

	Resources []string

	Builder          func() *resource.Builder
	Approver         internalpolymorphichelpers.ObjectApproverFunc
	Namespace        string
	EnforceNamespace bool

	Filenames []string
	Kustomize string
	Out       io.Writer

	resource.FilenameOptions
	genericclioptions.IOStreams
}

var (
	ApproveLong = templates.LongDesc(`
		Approve a resource which can be continued.

		Paused resources will not be reconciled by a controller. By approving a
		resource, we allow it to be continue to rollout.
		Currently only kruise-rollouts support being approved.`)

	ApproveExample = templates.Examples(`
		# approve a kruise rollout resource named "rollout-demo" in "ns-demo" namespace
		
		kubectl-kruise rollout approve rollout/rollout-demo -n ns-demo`)
)

// NewRolloutApproveOptions returns an initialized ApproveOptions instance
func NewRolloutApproveOptions(streams genericclioptions.IOStreams) *ApproveOptions {
	return &ApproveOptions{
		PrintFlags: genericclioptions.NewPrintFlags("approved").WithTypeSetter(internalapi.GetScheme()),
		IOStreams:  streams,
	}
}

// NewCmdRolloutApprove returns a Command instance for 'rollout approve' sub command
func NewCmdRolloutApprove(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewRolloutApproveOptions(streams)

	cmd := &cobra.Command{
		Use:                   "approve RESOURCE",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Approve a resource"),
		Long:                  ApproveLong,
		Example:               ApproveExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.RunApprove())
		},
	}

	usage := "identifying the resource to get from a server."
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
	o.PrintFlags.AddFlags(cmd)
	return cmd
}

// Complete completes all the required options
func (o *ApproveOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.Resources = args

	o.Approver = internalpolymorphichelpers.ObjectApproverFn

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

func (o *ApproveOptions) Validate() error {
	if len(o.Resources) == 0 && cmdutil.IsFilenameSliceEmpty(o.Filenames, o.Kustomize) {
		return fmt.Errorf("required resource not specified")
	}
	return nil
}

// RunApprove performs the execution of 'rollout approve' sub command
func (o *ApproveOptions) RunApprove() error {
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

	for _, patch := range set.CalculatePatches(infos, scheme.DefaultJSONEncoder(), set.PatchFn(o.Approver)) {
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
			printer, err := o.ToPrinter("already approved")
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
		printer, err := o.ToPrinter("approved")
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
