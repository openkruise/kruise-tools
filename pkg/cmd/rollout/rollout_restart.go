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

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	internalapi "github.com/openkruise/kruise-tools/pkg/api"
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

// RestartOptions is the start of the data required to perform the operation.  As new fields are added, add them here instead of
// referencing the cmd.Flags()
type RestartOptions struct {
	PrintFlags *genericclioptions.PrintFlags
	ToPrinter  func(string) (printers.ResourcePrinter, error)

	Resources []string

	Builder          func() *resource.Builder
	Restarter        internalpolymorphichelpers.ObjectRestarterFunc
	Namespace        string
	EnforceNamespace bool

	Filenames []string
	Kustomize string
	Out       io.Writer

	resource.FilenameOptions
	genericclioptions.IOStreams
}

var (
	restartLong = templates.LongDesc(`
		Restart a resource.

	        Resource will be rollout restarted.`)

	restartExample = templates.Examples(`
		# Restart a deployment
		kubectl-kruise rollout restart deployment/nginx
		kubectl-kruise rollout restart cloneset/abc

		# Restart a daemonset
		kubectl-kruise rollout restart daemonset/abc

		# Restart a UnitedDeployment
		kubectl-kruise rollout restart uniteddeployment/my-app`)
)

// NewRolloutRestartOptions returns an initialized RestartOptions instance
func NewRolloutRestartOptions(streams genericclioptions.IOStreams) *RestartOptions {
	return &RestartOptions{
		PrintFlags: genericclioptions.NewPrintFlags("restarted").WithTypeSetter(internalapi.GetScheme()),
		IOStreams:  streams,
	}
}

// NewCmdRolloutRestart returns a Command instance for 'rollout restart' sub command
func NewCmdRolloutRestart(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewRolloutRestartOptions(streams)

	validArgs := []string{"deployment", "daemonset", "statefulset", "cloneset", "uniteddeployment"}

	cmd := &cobra.Command{
		Use:                   "restart RESOURCE",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Restart a resource"),
		Long:                  restartLong,
		Example:               restartExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.RunRestart())
		},
		ValidArgs: validArgs,
	}

	usage := "identifying the resource to get from a server."
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
	o.PrintFlags.AddFlags(cmd)
	return cmd
}

// Complete completes all the required options
func (o *RestartOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	o.Resources = args

	o.Restarter = internalpolymorphichelpers.ObjectRestarterFn

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

func (o *RestartOptions) Validate() error {
	if len(o.Resources) == 0 && cmdutil.IsFilenameSliceEmpty(o.Filenames, o.Kustomize) {
		return fmt.Errorf("required resource not specified")
	}
	return nil
}

// RunRestart performs the execution of 'rollout restart' sub command
func (o RestartOptions) RunRestart() error {
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

	switch infos[0].Object.(type) {
	case *kruiseappsv1alpha1.CloneSet, *kruiseappsv1beta1.StatefulSet, *kruiseappsv1alpha1.DaemonSet:

		obj, err := resource.
			NewHelper(infos[0].Client, infos[0].Mapping).
			Get(infos[0].Namespace, infos[0].Name)
		if err != nil {
			return err
		}
		internalpolymorphichelpers.UpdateResourceEnv(obj)

		_, err = resource.
			NewHelper(infos[0].Client, infos[0].Mapping).
			Replace(infos[0].Namespace, infos[0].Name, true, obj)
		if err != nil {
			return err
		}
		printer, err := o.ToPrinter("restarted")
		if err != nil {
			allErrs = append(allErrs, err)
		}
		if err = printer.PrintObj(infos[0].Object, o.Out); err != nil {
			allErrs = append(allErrs, err)
		}
		return utilerrors.NewAggregate(allErrs)

	default:
		for _, patch := range set.CalculatePatches(infos, scheme.DefaultJSONEncoder(), set.PatchFn(o.Restarter)) {
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
				allErrs = append(allErrs, fmt.Errorf("failed to create patch for %v: empty patch", info.Name))
			}

			obj, err := resource.NewHelper(info.Client, info.Mapping).Patch(info.Namespace, info.Name, types.MergePatchType, patch.Patch, nil)
			if err != nil {
				allErrs = append(allErrs, fmt.Errorf("failed to patch: %v", err))
				continue
			}

			info.Refresh(obj, true)
			printer, err := o.ToPrinter("restarted")
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

}
