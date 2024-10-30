/*
Copyright 2022 The Kruise Authors.

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

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseclientsets "github.com/openkruise/kruise-api/client/clientset/versioned"
	internalapi "github.com/openkruise/kruise-tools/pkg/api"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	utilpointer "k8s.io/utils/pointer"
)

var (
	crrLong = templates.LongDesc(i18n.T(`
		Create a crr with the specified name.`))

	crrExample = templates.Examples(i18n.T(`
		NOTE: The default value of CRR is:
		strategy:
    		failurePolicy: Fail                
    		orderedRecreate: false             
    		unreadyGracePeriodSeconds: 3       
  		activeDeadlineSeconds: 300        
  		ttlSecondsAfterFinished: 1800     
		# Create a crr with default value to restart all containers in pod-1
		kubectl kruise create ContainerRecreateRequest my-crr --namespace=ns --pod=pod-1

		# Create a crr with default value to restart  container-1 in pod-1
		kubectl kruise create ContainerRecreateRequest my-crr --namespace=ns --pod=pod-1 --containers=container-1

		# Create a crr with unreadyGracePeriodSeconds 5 and terminationGracePeriodSeconds 30 to restart  container-1 in pod-1
		kubectl kruise create ContainerRecreateRequest my-crr --namespace=ns --pod=pod-1 --containers=container-1 --unreadyGracePeriodSeconds=5 --terminationGracePeriodSeconds=30`))

	defaultCRRStrategy = kruiseappsv1alpha1.ContainerRecreateRequestStrategy{
		FailurePolicy:             kruiseappsv1alpha1.ContainerRecreateRequestFailurePolicyFail,
		OrderedRecreate:           false,
		UnreadyGracePeriodSeconds: utilpointer.Int64Ptr(3),
		MinStartedSeconds:         3,
	}
	defaultActiveDeadlineSeconds   int64 = 300
	defaultTtlSecondsAfterFinished int32 = 1800
)

// CreateCRROptions is the command line options for 'create crr'
type CreateCRROptions struct {
	PrintFlags *genericclioptions.PrintFlags

	PrintObj func(obj runtime.Object) error

	PodName                           string
	UnreadyGracePeriodSeconds         int64
	MinStartedSeconds                 int32
	Containers                        []string
	ContainerRecreateRequestContainer []kruiseappsv1alpha1.ContainerRecreateRequestContainer

	Namespace            string
	Name                 string
	From                 string
	kruisev1alpha1Client kruiseclientsets.Interface
	ClientSet            kubernetes.Interface
	EnforceNamespace     bool
	DryRunStrategy       cmdutil.DryRunStrategy
	Builder              *resource.Builder
	FieldManager         string
	CreateAnnotation     bool

	genericclioptions.IOStreams
}

// NewCreateCRROptions initializes and returns new CreateCRROptions instance
func NewCreateCRROptions(ioStreams genericclioptions.IOStreams) *CreateCRROptions {
	return &CreateCRROptions{
		PrintFlags: genericclioptions.NewPrintFlags("created").WithTypeSetter(internalapi.GetScheme()),
		IOStreams:  ioStreams,
	}
}

// NewCmdCreateCRR is a command to ease creating ContainerRecreateRequest.
func NewCmdCreateCRR(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewCreateCRROptions(ioStreams)
	cmd := &cobra.Command{
		Use:                   "ContainerRecreateRequest NAME --pod=podName [--containers=container]",
		DisableFlagsInUseLine: true,
		Short:                 crrLong,
		Long:                  crrLong,
		Example:               crrExample,
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
	cmd.Flags().StringVarP(&o.PodName, "pod", "p", o.PodName, "The name of the pod).")
	cmd.Flags().StringSlice("containers", o.Containers, "The containers those need to restarted.")
	cmd.Flags().Int64VarP(&o.UnreadyGracePeriodSeconds, "unreadyGracePeriodSeconds", "u", o.UnreadyGracePeriodSeconds, "UnreadyGracePeriodSeconds is the optional duration in seconds to mark Pod as not ready over this duration before executing preStop hook and stopping the container")
	cmd.Flags().Int32VarP(&o.MinStartedSeconds, "minStartedSeconds", "m", o.MinStartedSeconds, "Minimum number of seconds for which a newly created container should be started and ready without any of its container crashing, for it to be considered Succeeded.Defaults to 0 (container will be considered Succeeded as soon as it is started and ready)")
	cmdutil.AddFieldManagerFlagVar(cmd, &o.FieldManager, "kubectl kruise-create")
	return cmd
}

// Complete completes all the required options
func (o *CreateCRROptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}
	o.Name = name
	clientConfig, err := f.ToRESTConfig()
	if err != nil {
		return err
	}
	o.ClientSet, err = kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	o.kruisev1alpha1Client, err = kruiseclientsets.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	if len(o.Namespace) == 0 {
		o.Namespace = "default"
	}
	o.CreateAnnotation = cmdutil.GetFlagBool(cmd, cmdutil.ApplyAnnotationsFlag)

	o.Builder = f.NewBuilder()

	if len(o.Containers) == 0 {
		// restart all containers in pod
		o.ContainerRecreateRequestContainer, err = o.getAllCRRContainersInPod()
		if err != nil {
			return err
		}
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

// Validate makes sure provided values and valid crr options
func (o *CreateCRROptions) Validate() error {
	// validate whether containers in pod
	if !o.isPodExist() {
		return fmt.Errorf("pod %s in namespace %s not exist", o.Name, o.Namespace)
	}
	var containers []string
	crrContainers, err := o.getAllCRRContainersInPod()
	if err != nil {
		return fmt.Errorf("found containers failed")
	}
	for i := range crrContainers {
		containers = append(containers, crrContainers[i].Name)
	}
	allContainerSet := sets.NewString(containers...)
	if !allContainerSet.HasAll(o.Containers...) {
		return fmt.Errorf("has container not exist")
	}

	return nil
}

// Run performs the execution of 'create crr' sub command
func (o *CreateCRROptions) Run() error {
	crr := o.createCRR()

	if o.DryRunStrategy != cmdutil.DryRunClient {
		createOptions := metav1.CreateOptions{}
		if o.FieldManager != "" {
			createOptions.FieldManager = o.FieldManager
		}
		if o.DryRunStrategy == cmdutil.DryRunServer {
			createOptions.DryRun = []string{metav1.DryRunAll}
		}
		var err error
		crr, err = o.kruisev1alpha1Client.AppsV1alpha1().ContainerRecreateRequests(o.Namespace).Create(context.TODO(), crr, createOptions)
		if err != nil {
			return fmt.Errorf("failed to create crr: %v", err)
		}
	}

	return o.PrintObj(crr)
}

func (o *CreateCRROptions) getAllCRRContainersInPod() ([]kruiseappsv1alpha1.ContainerRecreateRequestContainer, error) {
	var crrContainers []kruiseappsv1alpha1.ContainerRecreateRequestContainer
	var selectedCrrContainers []kruiseappsv1alpha1.ContainerRecreateRequestContainer
	selectedContainers := sets.NewString(o.Containers...)
	pod, err := o.ClientSet.CoreV1().Pods(o.Namespace).Get(context.TODO(), o.PodName, metav1.GetOptions{})
	if err != nil {
		return crrContainers, err
	}

	for _, container := range pod.Spec.Containers {
		crrContainer := kruiseappsv1alpha1.ContainerRecreateRequestContainer{}
		if container.Lifecycle != nil {
			crrContainer.PreStop = &kruiseappsv1alpha1.ProbeHandler{
				Exec:      container.Lifecycle.PreStop.Exec,
				HTTPGet:   container.Lifecycle.PreStop.HTTPGet,
				TCPSocket: container.Lifecycle.PreStop.TCPSocket,
			}
		}
		crrContainer.Name = container.Name
		crrContainer.Ports = container.Ports
		crrContainers = append(crrContainers, crrContainer)
	}
	if len(o.Containers) == 0 {
		return crrContainers, nil
	} else {
		for _, container := range crrContainers {
			if selectedContainers.Has(container.Name) {
				selectedCrrContainers = append(selectedCrrContainers, container)
			}
		}

		return selectedCrrContainers, nil
	}
}

func (o *CreateCRROptions) isPodExist() bool {
	_, err := o.ClientSet.CoreV1().Pods(o.Namespace).Get(context.TODO(), o.PodName, metav1.GetOptions{})
	if err != nil {
		return false
	}
	return true
}

func (o *CreateCRROptions) createCRR() *kruiseappsv1alpha1.ContainerRecreateRequest {
	crr := &kruiseappsv1alpha1.ContainerRecreateRequest{
		TypeMeta: metav1.TypeMeta{APIVersion: kruiseappsv1alpha1.SchemeGroupVersion.String(), Kind: "ContainerRecreateRequest"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.Name,
			Namespace: o.Namespace,
		},
		Spec: kruiseappsv1alpha1.ContainerRecreateRequestSpec{
			PodName:                 o.PodName,
			Containers:              o.ContainerRecreateRequestContainer,
			Strategy:                &defaultCRRStrategy,
			ActiveDeadlineSeconds:   utilpointer.Int64Ptr(defaultActiveDeadlineSeconds),
			TTLSecondsAfterFinished: utilpointer.Int32Ptr(defaultTtlSecondsAfterFinished),
		},
	}
	if o.EnforceNamespace {
		crr.Namespace = o.Namespace
	}

	return crr
}
