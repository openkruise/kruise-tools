/*
Copyright 2020 The Kruise Authors.

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

package migrate

import (
	"fmt"

	"github.com/openkruise/kruise-tools/pkg/api"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type migrateOptions struct {
	Namespace string

	From    string
	To      string
	SrcName string
	DstName string
	SrcRef  api.ResourceRef
	DstRef  api.ResourceRef

	IsCreate       bool
	IsCopy         bool
	Replicas       int32
	MaxSurge       int32
	TimeoutSeconds int32

	genericclioptions.IOStreams
}

func newMigrateOptions(ioStreams genericclioptions.IOStreams) *migrateOptions {
	return &migrateOptions{IOStreams: ioStreams}
}

func NewCmdMigrate(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := newMigrateOptions(ioStreams)

	cmd := &cobra.Command{
		Use:                   "migrate [DST_KIND] --from [SRC_KIND] [flags]",
		DisableFlagsInUseLine: true,
		Short:                 "Migrate from K8s original workloads to Kruise workloads",
		Long:                  "Migrate from K8s original workloads to Kruise workloads",
		Example: `
	# Create an empty CloneSet from an existing Deployment.
	kruise migrate CloneSet --from Deployment -n default --dst-name deployment-name --create

	# Create a same replicas CloneSet from an existing Deployment.
	kruise migrate CloneSet --from Deployment -n default --dst-name deployment-name --create --copy

	# Migrate replicas from an existing Deployment to an existing CloneSet.
	kruise migrate CloneSet --from Deployment -n default --src-name cloneset-name --dst-name deployment-name --replicas 10 --max-surge=2
`,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Run(f, cmd))
		},
	}

	cmd.Flags().StringVar(&o.From, "from", "", "Type of the source workload (e.g. Deployment).")
	cmd.Flags().StringVar(&o.SrcName, "src-name", "", "Name of the source workload.")
	cmd.Flags().StringVar(&o.DstName, "dst-name", "", "Name of the destination workload.")

	cmd.Flags().BoolVar(&o.IsCreate, "create", false, "Create dst workload with replicas=0 from src workload.")
	cmd.Flags().BoolVar(&o.IsCopy, "copy", false, "Copy replicas from src workload when create.")
	cmd.Flags().Int32Var(&o.Replicas, "replicas", -1, "The replicas needs to migrate, -1 indicates all replicas in src workload.")
	cmd.Flags().Int32Var(&o.MaxSurge, "max-surge", 1, "Max surge during migration.")
	cmd.Flags().Int32Var(&o.TimeoutSeconds, "timeout-seconds", -1, "Timeout seconds for migration, -1 indicates no limited.")

	return cmd
}

func (o *migrateOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	namespace, explicitNamespace, err := f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	} else if !explicitNamespace {
		return fmt.Errorf("must specify namespace by -n or --namespace")
	}
	o.Namespace = namespace

	if len(args) == 0 {
		return fmt.Errorf("must specify workload type like CloneSet")
	} else if len(args) > 1 {
		return fmt.Errorf("more than one given args")
	}

	if len(o.From) == 0 {
		return fmt.Errorf("must specify --from")
	}
	if len(o.SrcName) == 0 {
		return fmt.Errorf("must specify --src-name")
	}
	if len(o.DstName) == 0 && !o.IsCreate {
		return fmt.Errorf("must specify --dst-name")
	}

	switch args[0] {
	case "CloneSet", "cloneset":
		o.To = "CloneSet"
		o.DstRef = api.NewCloneSetRef(namespace, o.DstName)
	default:
		return fmt.Errorf("currently only supported CloneSet as dst type")
	}

	switch o.From {
	case "Deployment", "deployment":
		o.From = "Deployment"
		o.SrcRef = api.NewDeploymentRef(namespace, o.SrcName)
	default:
		return fmt.Errorf("currently only supported Deployment as src type")
	}

	return nil
}

func (o *migrateOptions) Run(f cmdutil.Factory, cmd *cobra.Command) error {
	switch o.To {
	case "CloneSet":
		return o.migrateCloneSet(f, cmd)
	}
	return nil
}
