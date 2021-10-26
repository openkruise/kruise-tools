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

package scaledown

import (
	"fmt"
	"k8s.io/cli-runtime/pkg/printers"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	internalapi "github.com/openkruise/kruise-tools/pkg/api"
	"github.com/openkruise/kruise-tools/pkg/cmd/util"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
)

type ScaleDownOptions struct {
	Resources        []string
	Namespace        string
	EnforceNamespace bool
	Pods             string
	Builder          func() *resource.Builder

	PrintFlags *genericclioptions.PrintFlags
	PrintObj   printers.ResourcePrinterFunc
	resource.FilenameOptions
	genericclioptions.IOStreams
}

func newScaleDownOptions(ioStreams genericclioptions.IOStreams) *ScaleDownOptions {
	return &ScaleDownOptions{
		PrintFlags: genericclioptions.NewPrintFlags("selective pod deletion").WithTypeSetter(scheme.Scheme),
		IOStreams:  ioStreams,
	}
}

func NewCmdScaleDown(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := newScaleDownOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "scaledown [CLONESET] --pods [POD1,POD2] -n [NAMESPACE]",
		DisableFlagsInUseLine: true,
		Short:                 "Scaledown a cloneset with selective Pods",
		Long:                  "Scaledown a cloneset with selective Pods",
		Example: `
		# Scale down 2 with  selective pods
		kubectl-kruise scaledown CloneSet cloneset-demo --pods pod-1, pod-2 -n default
`,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Run(f, cmd))
		},
	}

	cmd.Flags().StringVar(&o.Pods, "pods", "", "Name of the pods to delete")

	return cmd
}

func (o *ScaleDownOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	namespace, explicitNamespace, err := f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	o.Namespace = namespace
	o.EnforceNamespace = explicitNamespace
	o.Resources = args
	o.Builder = f.NewBuilder

	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.PrintObj = printer.PrintObj
	if len(o.Pods) == 0 {
		return fmt.Errorf("must specify one pod name")
	}

	return nil
}

func (o *ScaleDownOptions) Run(f cmdutil.Factory, cmd *cobra.Command) error {
	b := o.Builder().
		WithScheme(internalapi.GetScheme(), scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(o.Namespace).DefaultNamespace().
		FilenameParam(o.EnforceNamespace, &o.FilenameOptions).
		ResourceTypeOrNameArgs(true, o.Resources...).
		ContinueOnError().
		Latest().
		Flatten()

	infos, err := b.Do().Infos()
	if err != nil {
		return err
	}
	if len(infos) == 0 {
		return nil
	}

	cl := util.BaseClient()
	switch infos[0].Object.(type) {
	case *kruiseappsv1alpha1.CloneSet:
		err = o.ScaleDownCloneSet(f, infos[0].Name, cl.Reader, cl.Client)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("currently only supported CloneSet selective pods deletion")
	}

	return nil
}
