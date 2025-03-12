/*
Copyright 2024 The Kruise Authors.

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

package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (	
	getLong = templates.Examples(i18n.T(`
		Display one or many resources related to kruise.`))
	
	getExample = templates.Examples(i18n.T(`
		# List all resources in the default namespace
		kubectl-kruise get all 

		# List all resources in the specific namespace
		kubectl-kruise get all -n namespace

		# Watch all resources in the default namespace
		kubectl-kruise get all -w`))


)


type GetOptions struct {
	genericclioptions.IOStreams
	Builder func() *resource.Builder
	Resources []string
	Namespace string
	EnforceNamespace bool
	Watch bool
}


func NewCmdGet(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := &GetOptions{IOStreams: streams}

	cmd := &cobra.Command{
		Use:                   "get SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Display one or many resources"),
		Run: func(cmd *cobra.Command, args []string) {
			// cmdutil.CheckErr(o.Complete(f, args))
			// cmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().BoolVarP(&o.Watch, "watch", "w", o.Watch, "Watch all details related to kruise resources")


	return cmd
}

func (o *GetOptions) Complete(f cmdutil.Factory, args []string) error {
	var err error

	if len(args) == 0 {
		return fmt.Errorf("required resource not specified")
	}

	o.Builder = f.NewBuilder

	return nil
}

func (o *GetOptions) Run() error {
	
}
