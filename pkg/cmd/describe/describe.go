package describe

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

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	describeLong = templates.LongDesc(i18n.T(`
		Show details of a rollout.`))

	describeExample = templates.Examples(`
		# Describe the rollout named rollout-demo
		kubectl-kruise describe rollout rollout-demo`)
)

// NewCmdRollout returns a Command instance for 'rollout' sub command
func NewCmdDescibe(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "describe SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Show details of a specific resource or group of resources"),
		Long:                  describeLong,
		Example:               describeExample,
		Run:                   cmdutil.DefaultSubCommandRun(streams.Out),
	}
	// subcommands
	cmd.AddCommand(NewCmdDescribeRollout(f, streams))

	return cmd
}
