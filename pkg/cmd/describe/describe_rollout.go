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

package describe

import (
	"fmt"

	internalapi "github.com/openkruise/kruise-tools/pkg/api"
	internalpolymorphichelpers "github.com/openkruise/kruise-tools/pkg/internal/polymorphichelpers"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	rolloutLong = templates.LongDesc(i18n.T(`
		Get details about and visual representation of a rollout.`))

	rolloutExample = templates.Examples(`
		# Describe the rollout of a resource
		kubectl-kruise describe rollout default/cloneset`)
)

type DescribeRolloutOptions struct {
	Builder          func() *resource.Builder
	Namespace        string
	EnforceNamespace bool
	Resources []string
	RolloutViewerFn  func(rollout *meta.RESTMapping) (internalpolymorphichelpers.RolloutViewer, error)

	PrintFlags      *genericclioptions.PrintFlags
	ToPrinter  func(string) (printers.ResourcePrinter, error)
	Watch           bool
	NoColor        bool
	TimeoutSeconds int

	resource.FilenameOptions
	genericclioptions.IOStreams
}

func NewDescribeRolloutOptions(streams genericclioptions.IOStreams) *DescribeRolloutOptions {
	return &DescribeRolloutOptions{
		PrintFlags: genericclioptions.NewPrintFlags("").WithDefaultOutput("yaml"),
		IOStreams:  streams,
	}
}

func NewCmdDescribeRollout(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewDescribeRolloutOptions(streams)
	cmd := &cobra.Command{
		Use:                   "rollout SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Get details about a rollout"),
		Long:                  rolloutLong,
		Example:               rolloutExample,
		Aliases:               []string{"rollouts", "ro"},
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().BoolVarP(&o.Watch, "watch", "w", false, "Watch for changes to the rollout")
	cmd.Flags().BoolVar(&o.NoColor, "no-color", false, "If true, print output without color")
	cmd.Flags().IntVar(&o.TimeoutSeconds, "timeout", 0, "Timeout after specified seconds")

	return cmd
}

func (o *DescribeRolloutOptions) Complete(f cmdutil.Factory, args []string) error {
	o.Resources = args 

	var err error
	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.RolloutViewerFn = internalpolymorphichelpers.RolloutViewerFn


	o.Builder = f.NewBuilder
	return nil
}

func (o *DescribeRolloutOptions) Validate() error {
	if len(o.Resources) == 0 {
		return fmt.Errorf("required resource not specified")
	}
	return nil
}
func (o *DescribeRolloutOptions) Run() error {
	r := o.Builder().
		WithScheme(internalapi.GetScheme(), scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(o.Namespace).
		DefaultNamespace().
		FilenameParam(o.EnforceNamespace, &o.FilenameOptions).
		ResourceTypeOrNameArgs(true, o.Resources...).
		ContinueOnError().
		Latest().
		Flatten().
		Do()

	if err := r.Err(); err != nil {
		return err
	}

	return nil
}
