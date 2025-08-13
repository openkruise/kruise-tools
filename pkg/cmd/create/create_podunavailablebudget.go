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

package create

import (
	"context"
	"fmt"
	"strings"

	kruisepolicyv1alpha1 "github.com/openkruise/kruise-api/policy/v1alpha1"
	"github.com/spf13/cobra"

	policyclientv1alpha1 "github.com/openkruise/kruise-api/client/clientset/versioned/typed/policy/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	podUnavailableBudgetLong = templates.LongDesc(i18n.T(`
		Create a pod unavailable budget with the specified name, selector, and desired minimum available pods.`))

	podUnavailableBudgetExample = templates.Examples(i18n.T(`
		# Create a pod unavailable budget named my-pub that will select all pods with the app=rails label
		# and require at least one of them being available at any point in time
		kubectl kruise create podunavailablebudget my-pub --selector=app=rails --min-available=1

		# Create a pod unavailable budget named my-pub that will select all pods with the app=nginx label
		# and require at least half of the pods selected to be available at any point in time
		kubectl kruise create podunavailablebudget my-pub --selector=app=nginx --min-available=50%

		# Create a pod unavailable budget named my-pub that will select all pods with the app=nginx label
		# and allow a maximum of 30% of pods to be unavailable at any point in time
		kubectl kruise create podunavailablebudget my-pub --selector=app=nginx --max-unavailable=30%`))
)

// CreatePodUnavailableBudgetOptions is the command line options for 'create podunavailablebudget'
type CreatePodUnavailableBudgetOptions struct {
	PrintFlags *genericclioptions.PrintFlags

	PrintObj func(obj runtime.Object) error

	Name           string
	Selector       string
	MinAvailable   string
	MaxUnavailable string

	Namespace        string
	EnforceNamespace bool
	Client           policyclientv1alpha1.PolicyV1alpha1Interface
	DryRunStrategy   cmdutil.DryRunStrategy
	FieldManager     string
	CreateAnnotation bool

	genericclioptions.IOStreams
}

// NewCreatePodUnavailableBudgetOptions initializes and returns new CreatePodUnavailableBudgetOptions instance
func NewCreatePodUnavailableBudgetOptions(ioStreams genericclioptions.IOStreams) *CreatePodUnavailableBudgetOptions {
	return &CreatePodUnavailableBudgetOptions{
		PrintFlags: genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
		IOStreams:  ioStreams,
	}
}

// NewCmdCreatePodUnavailableBudget is a command to ease creating PodUnavailableBudgets.
func NewCmdCreatePodUnavailableBudget(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewCreatePodUnavailableBudgetOptions(ioStreams)
	cmd := &cobra.Command{
		Use:                   "podunavailablebudget NAME --selector=SELECTOR (--min-available=N | --max-unavailable=N)",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"pub"},
		Short:                 podUnavailableBudgetLong,
		Long:                  podUnavailableBudgetLong,
		Example:               podUnavailableBudgetExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	o.PrintFlags.AddFlags(cmd)

	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddDryRunFlag(cmd)
	cmd.Flags().StringVar(&o.Selector, "selector", o.Selector, "A label selector to use for this budget. Only equality-based selector requirements are supported.")
	cmd.Flags().StringVar(&o.MinAvailable, "min-available", o.MinAvailable, "The minimum number or percentage of available pods this budget requires.")
	cmd.Flags().StringVar(&o.MaxUnavailable, "max-unavailable", o.MaxUnavailable, "The maximum number or percentage of unavailable pods this budget requires.")
	cmdutil.AddFieldManagerFlagVar(cmd, &o.FieldManager, "kubectl-kruise-create")
	return cmd
}

// Complete completes all the required options
func (o *CreatePodUnavailableBudgetOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}
	o.Name = name

	clientConfig, err := f.ToRESTConfig()
	if err != nil {
		return err
	}

	kruiseClient, err := policyclientv1alpha1.NewForConfig(clientConfig)
	if err != nil {
		return err
	}
	o.Client = kruiseClient

	o.CreateAnnotation = cmdutil.GetFlagBool(cmd, cmdutil.ApplyAnnotationsFlag)

	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.DryRunStrategy, err = cmdutil.GetDryRunStrategy(cmd)
	if err != nil {
		return err
	}
	cmdutil.PrintFlagsWithDryRunStrategy(o.PrintFlags, o.DryRunStrategy)
	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.PrintObj = func(obj runtime.Object) error {
		return printer.PrintObj(obj, o.Out)
	}

	return nil
}

// Validate makes sure provided values and valid PodUnavailableBudget options
func (o *CreatePodUnavailableBudgetOptions) Validate() error {
	if len(o.Selector) == 0 {
		return fmt.Errorf("--selector must be specified")
	}
	if len(o.MinAvailable) == 0 && len(o.MaxUnavailable) == 0 {
		return fmt.Errorf("either --min-available or --max-unavailable must be specified")
	}
	if len(o.MinAvailable) > 0 && len(o.MaxUnavailable) > 0 {
		return fmt.Errorf("--min-available and --max-unavailable cannot both be specified")
	}
	return nil
}

// Run performs the execution of 'create podunavailablebudget' sub command
func (o *CreatePodUnavailableBudgetOptions) Run() error {
	pub := o.createPodUnavailableBudget()
	if err := util.CreateOrUpdateAnnotation(o.CreateAnnotation, pub, scheme.DefaultJSONEncoder()); err != nil {
		return err
	}

	if o.DryRunStrategy != cmdutil.DryRunClient {
		createOptions := metav1.CreateOptions{}
		if o.FieldManager != "" {
			createOptions.FieldManager = o.FieldManager
		}
		if o.DryRunStrategy == cmdutil.DryRunServer {
			createOptions.DryRun = []string{metav1.DryRunAll}
		}
		var err error
		pub, err = o.Client.PodUnavailableBudgets(o.Namespace).Create(context.TODO(), pub, createOptions)
		if err != nil {
			return err
		}
	}

	return o.PrintObj(pub)
}

func (o *CreatePodUnavailableBudgetOptions) createPodUnavailableBudget() *kruisepolicyv1alpha1.PodUnavailableBudget {
	pub := &kruisepolicyv1alpha1.PodUnavailableBudget{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kruisepolicyv1alpha1.SchemeGroupVersion.String(),
			Kind:       "PodUnavailableBudget",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: o.Name,
		},
		Spec: kruisepolicyv1alpha1.PodUnavailableBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: parseLabels(o.Selector),
			},
		},
	}

	if len(o.MinAvailable) > 0 {
		minAvailable := intstr.Parse(o.MinAvailable)
		pub.Spec.MinAvailable = &minAvailable
	} else if len(o.MaxUnavailable) > 0 {
		maxUnavailable := intstr.Parse(o.MaxUnavailable)
		pub.Spec.MaxUnavailable = &maxUnavailable
	}

	return pub
}

// parseLabels parses a selector string in the format "key1=value1,key2=value2" into a map
func parseLabels(selector string) map[string]string {
	labels := make(map[string]string)
	if selector == "" {
		return labels
	}

	pairs := strings.Split(selector, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			labels[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return labels
}
