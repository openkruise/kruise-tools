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
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	rolloutLong = templates.LongDesc(i18n.T(`
		Get details about and visual representation of a rollout.`))

	rolloutExample = templates.Examples(`
		# Describe the rollout of a resource
		kubectl-kruise describe rollout cloneset/abc`)
)


type DescribeRolloutOptions struct {
	Builder          func() *resource.Builder
	Namespace        string
	EnforceNamespace bool

	PrintFlags *genericclioptions.PrintFlags
	ClientSet      kubernetes.Interface
	Watch bool
	NoColor bool
	TimeoutSeconds int 
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
		Aliases: []string{"rollouts", "ro"},
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(cmd.Run())
		},
	}


	cmd.Flags().BoolVarP(&o.Watch, "watch", "w", false, "Watch for changes to the rollout")
	cmd.Flags().BoolVar(&o.NoColor, "no-color", false, "If true, print output without color")
	cmd.Flags().IntVar(&o.TimeoutSeconds, "timeout", 0, "Timeout after specified seconds")

	return cmd
}

