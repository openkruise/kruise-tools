/*
Copyright 2022 The Kruise Authors.
Copyright 2018 The Kubernetes Authors.

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
	"github.com/spf13/cobra"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	broadcastJobLong = templates.LongDesc(i18n.T(`
		Create a broadcastJob with the specified name.`))

	broadcastJobExample = templates.Examples(i18n.T(`
		# Create a broadcastJob
		kubectl kruise create broadcastJob my-bcj --image=busybox

		# Create a broadcastJob with command
		kubectl kruise create broadcastJob my-bcj --image=busybox -- date

		# Create a broadcastJob from a AdvancedCronJob named "a-advancedCronjob"
		kubectl kruise create broadcastJob test-bcj --from=acj/a-advancedCronjob`))
)

// CreateBroadcastJobOptions is the command line options for 'create broadcastJob'
type CreateBroadcastJobOptions struct {
	PrintFlags *genericclioptions.PrintFlags

	PrintObj func(obj runtime.Object) error

	Name    string
	Image   string
	From    string
	Command []string

	Namespace            string
	EnforceNamespace     bool
	kruisev1alpha1Client kruiseclientsets.Interface
	DryRunStrategy       cmdutil.DryRunStrategy
	Builder              *resource.Builder
	FieldManager         string
	CreateAnnotation     bool

	genericclioptions.IOStreams
}

// NewCreateBroadcastJobOptions initializes and returns new CreateBroadcastJobOptions instance
func NewCreateBroadcastJobOptions(ioStreams genericclioptions.IOStreams) *CreateBroadcastJobOptions {
	return &CreateBroadcastJobOptions{
		PrintFlags: genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
		IOStreams:  ioStreams,
	}
}

// NewCmdCreateBroadcastJob is a command to ease creating BroadcastJobs from AdvancedCronJobs.
func NewCmdCreateBroadcastJob(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewCreateBroadcastJobOptions(ioStreams)
	cmd := &cobra.Command{
		Use:                   "broadcastJob NAME --image=image [--from=cronjob/name] -- [COMMAND] [args...]",
		DisableFlagsInUseLine: true,
		Short:                 broadcastJobLong,
		Long:                  broadcastJobLong,
		Example:               broadcastJobExample,
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
	cmd.Flags().StringVar(&o.Image, "image", o.Image, "Image name to run.")
	cmd.Flags().StringVar(&o.From, "from", o.From, "The name of the resource to create a BroadcastJob from (only advancedCronjob is supported).")
	cmdutil.AddFieldManagerFlagVar(cmd, &o.FieldManager, "kubectl kruise-create")
	return cmd
}

// Complete completes all the required options
func (o *CreateBroadcastJobOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}
	o.Name = name
	if len(args) > 1 {
		o.Command = args[1:]
	}

	clientConfig, err := f.ToRESTConfig()
	if err != nil {
		return err
	}
	o.kruisev1alpha1Client, err = kruiseclientsets.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	o.CreateAnnotation = cmdutil.GetFlagBool(cmd, cmdutil.ApplyAnnotationsFlag)

	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	o.Builder = f.NewBuilder()

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

// Validate makes sure provided values and valid BroadcastJob options
func (o *CreateBroadcastJobOptions) Validate() error {
	if (len(o.Image) == 0 && len(o.From) == 0) || (len(o.Image) != 0 && len(o.From) != 0) {
		return fmt.Errorf("either --image or --from must be specified")
	}
	if o.Command != nil && len(o.Command) != 0 && len(o.From) != 0 {
		return fmt.Errorf("cannot specify --from and command")
	}
	return nil
}

// Run performs the execution of 'create broadcastJob' sub command
func (o *CreateBroadcastJobOptions) Run() error {
	var job *kruiseappsv1alpha1.BroadcastJob
	if len(o.Image) > 0 {
		job = o.createBroadcastJob()
	} else {
		infos, err := o.Builder.
			WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
			NamespaceParam(o.Namespace).DefaultNamespace().
			ResourceTypeOrNameArgs(false, o.From).
			Flatten().
			Latest().
			Do().
			Infos()
		if err != nil {
			return err
		}
		if len(infos) != 1 {
			return fmt.Errorf("from must be an existing advancedCronJob")
		}

		switch obj := infos[0].Object.(type) {
		case *kruiseappsv1alpha1.AdvancedCronJob:
			job, err = o.createBroadcastJobFromAdvancedCronJob(obj)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown object type %T", obj)
		}
	}

	if err := util.CreateOrUpdateAnnotation(o.CreateAnnotation, job, scheme.DefaultJSONEncoder()); err != nil {
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
		job, err = o.kruisev1alpha1Client.AppsV1alpha1().BroadcastJobs(o.Namespace).Create(context.TODO(), job, createOptions)
		if err != nil {
			return fmt.Errorf("failed to create job: %v", err)
		}
	}

	return o.PrintObj(job)
}

func (o *CreateBroadcastJobOptions) createBroadcastJob() *kruiseappsv1alpha1.BroadcastJob {
	job := &kruiseappsv1alpha1.BroadcastJob{
		// this is ok because we know exactly how we want to be serialized
		TypeMeta: metav1.TypeMeta{APIVersion: batchv1.SchemeGroupVersion.String(), Kind: kruiseappsv1alpha1.AdvancedCronJobKind},
		ObjectMeta: metav1.ObjectMeta{
			Name: o.Name,
		},
		Spec: kruiseappsv1alpha1.BroadcastJobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    o.Name,
							Image:   o.Image,
							Command: o.Command,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
	if o.EnforceNamespace {
		job.Namespace = o.Namespace
	}
	return job
}

func (o *CreateBroadcastJobOptions) createBroadcastJobFromAdvancedCronJob(cronJob *kruiseappsv1alpha1.AdvancedCronJob) (*kruiseappsv1alpha1.BroadcastJob, error) {
	if cronJob.Status.Type != kruiseappsv1alpha1.BroadcastJobTemplate {
		return nil, fmt.Errorf("from must be a broadcastJob template, but got %v", cronJob.Status.Type)
	}
	job := o.createBroadcastJobFromBroadcastJobTemplate(cronJob.GetName(), cronJob.GetUID(), cronJob.Spec.Template.BroadcastJobTemplate)
	return job, nil
}

func (o *CreateBroadcastJobOptions) createBroadcastJobFromBroadcastJobTemplate(ownerReferenceName string, ownerReferenceUID types.UID,
	broadcastJobTemplate *kruiseappsv1alpha1.BroadcastJobTemplateSpec) *kruiseappsv1alpha1.BroadcastJob {

	annotations := make(map[string]string)
	broadcastJob := &kruiseappsv1alpha1.BroadcastJob{}
	annotations["cronjob.kubernetes.io/instantiate"] = "manual"
	for k, v := range broadcastJobTemplate.Annotations {
		annotations[k] = v
	}
	broadcastJob = &kruiseappsv1alpha1.BroadcastJob{
		// this is ok because we know exactly how we want to be serialized
		TypeMeta: metav1.TypeMeta{APIVersion: batchv1.SchemeGroupVersion.String(), Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        o.Name,
			Annotations: annotations,
			Labels:      broadcastJobTemplate.Labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: kruiseappsv1alpha1.SchemeGroupVersion.String(),
					Kind:       kruiseappsv1alpha1.AdvancedCronJobKind,
					Name:       ownerReferenceName,
					UID:        ownerReferenceUID,
				},
			},
		},
		Spec: broadcastJobTemplate.Spec,
	}
	if o.EnforceNamespace {
		broadcastJob.Namespace = o.Namespace
	}

	return broadcastJob
}
