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

package top

import (
	"github.com/openkruise/kruise-tools/pkg/cmd/util"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	topLong = templates.LongDesc(i18n.T(`
		Display resource (CPU/Memory) usage.

		The top command allows you to see the resource consumption for OpenKruise workloads.`))
)

func NewCmdTop(f util.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "top",
		Short: i18n.T("Display resource (CPU/Memory) usage"),
		Long:  topLong,
		Run:   cmdutil.DefaultSubCommandRun(ioStreams.ErrOut),
	}

	cmd.AddCommand(NewCmdTopCloneSet(f, ioStreams))
	cmd.AddCommand(NewCmdTopAdvancedStatefulSet(f, ioStreams))
	cmd.AddCommand(NewCmdTopAdvancedDaemonSet(f, ioStreams))
	cmd.AddCommand(NewCmdTopUnitedDeployment(f, ioStreams))
	cmd.AddCommand(NewCmdTopBroadcastJob(f, ioStreams))

	return cmd
}
