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
	"fmt"

	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	"github.com/openkruise/kruise-tools/pkg/cmd/util"
	"github.com/openkruise/kruise-tools/pkg/top"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

type TopAdvancedStatefulSetOptions struct {
	genericclioptions.IOStreams
	ResourceName string
	Namespace    string
	Factory      util.Factory
}

func NewCmdTopAdvancedStatefulSet(f util.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := &TopAdvancedStatefulSetOptions{
		IOStreams: ioStreams,
		Factory:   f,
	}
	cmd := &cobra.Command{
		Use:                   "advancedstatefulset NAME",
		Short:                 i18n.T("Display resource (CPU/Memory) usage of an AdvancedStatefulSet."),
		Aliases:               []string{"asts"},
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "If present, the namespace scope for this CLI request")
	return cmd
}

func (o *TopAdvancedStatefulSetOptions) Complete(cmd *cobra.Command, args []string) error {
	o.ResourceName = args[0]
	var err error
	o.Namespace, _, err = o.Factory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("namespace") {
		o.Namespace, _ = cmd.Flags().GetString("namespace")
	}
	return nil
}

func (o *TopAdvancedStatefulSetOptions) Run() error {
	builder := o.Factory.NewBuilder()
	result := builder.NamespaceParam(o.Namespace).DefaultNamespace().ResourceNames("advancedstatefulset", o.ResourceName).Do()
	infos, err := result.Infos()
	if err != nil {
		return err
	}
	if len(infos) != 1 {
		return fmt.Errorf("expected one AdvancedStatefulSet resource, but found %d", len(infos))
	}
	resourceObj, ok := infos[0].Object.(*kruiseappsv1beta1.StatefulSet)
	if !ok {
		return fmt.Errorf("unexpected object type: %T, expected *kruiseappsv1beta1.StatefulSet", infos[0].Object)
	}

	selector, err := metav1.LabelSelectorAsSelector(resourceObj.Spec.Selector)
	if err != nil {
		return fmt.Errorf("could not convert label selector for AdvancedStatefulSet %s: %v", resourceObj.Name, err)
	}

	kubeClient, err := o.Factory.KubernetesClientSet()
	if err != nil {
		return err
	}
	metricsClient, err := o.Factory.MetricsClient()
	if err != nil {
		return fmt.Errorf("error getting metrics client: %v. Is the metrics-server installed?", err)
	}

	totalCPU, totalMemory, err := top.SumUsageForSelector(kubeClient, metricsClient, resourceObj.Namespace, selector)
	if err != nil {
		return err
	}

	cpuString, memoryString := top.FormatResourceUsage(totalCPU, totalMemory)
	fmt.Fprintf(o.Out, "%-20s\t%-12s\t%-15s\n", "NAME", "CPU(cores)", "MEMORY(bytes)")
	fmt.Fprintf(o.Out, "%-20s\t%-12s\t%-15s\n", resourceObj.Name, cpuString, memoryString)
	return nil
}
