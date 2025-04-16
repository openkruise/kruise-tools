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

package get

import (
	"fmt"
	"time"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	rolloutv1alpha1 "github.com/openkruise/kruise-rollout-api/rollouts/v1alpha1"
	rolloutv1beta1 "github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	internalpolymorphichelpers "github.com/openkruise/kruise-tools/pkg/internal/polymorphichelpers"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
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
	Builder          func() *resource.Builder
	Resources        []string
	Namespace        string
	EnforceNamespace bool
	GetViewerFn      internalpolymorphichelpers.GetViewerFunc
}

func NewCmdGet(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := &GetOptions{IOStreams: streams}

	cmd := &cobra.Command{
		Use:                   "get all",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Display one or many resources"),
		Long:                  getLong,
		Example:               getExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, args))
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "If present, the namespace scope for this CLI request")

	return cmd
}

func (o *GetOptions) Complete(f cmdutil.Factory, args []string) error {
	var err error

	if len(args) == 0 {
		return fmt.Errorf("required resource not specified")
	}

	o.Resources = args
	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.GetViewerFn = internalpolymorphichelpers.GetViewerFn
	o.Builder = f.NewBuilder

	return nil
}

func (o *GetOptions) Run() error {
	if len(o.Resources) == 0 {
		return fmt.Errorf("you must specify the type of resource to get")
	}

	if o.Resources[0] == "all" {
		resourceTypes := []string{
			"clonesets.apps.kruise.io",
			"statefulsets.apps.kruise.io",
			"daemonsets.apps.kruise.io",
			"rollouts.rollouts.kruise.io",
			"broadcastjobs.apps.kruise.io",
			"containerrecreaterequests.apps.kruise.io",
			"advancedcronjobs.apps.kruise.io",
			"resourcedistributions.apps.kruise.io",
			"uniteddeployments.apps.kruise.io",
		}

		for _, resourceType := range resourceTypes {
			b := o.Builder().
				WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
				NamespaceParam(o.Namespace).DefaultNamespace().
				ResourceTypeOrNameArgs(true, resourceType).
				ContinueOnError().
				Latest().
				Flatten()

			infos, err := b.Do().Infos()
			if err != nil {
				return err
			}

			if len(infos) > 0 {
				// Print resource type header and table header only if there are resources
				fmt.Fprintf(o.Out, "\n%s:\n", resourceType)
				switch resourceType {
				case "clonesets.apps.kruise.io":
					fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-14s\t%-18s\t%-8s\t%-8s\t%-s\n",
						"NAME", "DESIRED", "UPDATED", "UPDATED_READY", "UPDATED_AVAILABLE", "READY", "TOTAL", "AGE")
				case "statefulsets.apps.kruise.io":
					fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-8s\t%-s\n",
						"NAME", "DESIRED", "CURRENT", "UPDATED", "READY", "AGE")
				case "daemonsets.apps.kruise.io":
					fmt.Fprintf(o.Out, "%-8s\t%-8s\t%-8s\t%-8s\t%-8s\t%-8s\t%-8s\t%-s\n",
						"NAME", "DESIRED", "CURRENT", "READY", "UP-TO-DATE", "AVAILABLE", "NODE SELECTOR", "AGE")
				case "rollouts.rollouts.kruise.io":
					fmt.Fprintf(o.Out, "%-20s\t%-8s\t%-12s\t%-12s\t%-40s\t%-s\n",
						"NAME", "STATUS", "CANARY_STEP", "CANARY_STATE", "MESSAGE", "AGE")
				case "broadcastjobs.apps.kruise.io":
					fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-8s\t%-s\n",
						"NAME", "DESIRED", "ACTIVE", "SUCCEEDED", "FAILED", "AGE")
				case "containerrecreaterequests.apps.kruise.io":
					fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-s\n",
						"NAME", "PHASE", "COMPLETED", "FAILED", "AGE")
				case "advancedcronjobs.apps.kruise.io":
					fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-8s\t%-s\n",
						"NAME", "SCHEDULE", "SUSPEND", "ACTIVE", "LAST SCHEDULE", "AGE")
				case "resourcedistributions.apps.kruise.io":
					fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-s\n",
						"NAME", "TARGETS", "SUCCEEDED", "FAILED", "AGE")
				case "uniteddeployments.apps.kruise.io":
					fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-8s\t%-s\n",
						"NAME", "DESIRED", "UPDATED", "READY", "AVAILABLE", "AGE")
				}

				for _, info := range infos {
					if err := o.printResourceInfo(info.Object, resourceType); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	// Handle single resource type
	b := o.Builder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(o.Namespace).DefaultNamespace().
		ResourceTypeOrNameArgs(true, o.Resources...).
		ContinueOnError().
		Latest().
		Flatten()

	infos, err := b.Do().Infos()
	if err != nil {
		return err
	}

	if len(infos) > 0 {
		switch o.Resources[0] {
		case "clonesets.apps.kruise.io":
			fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-14s\t%-18s\t%-8s\t%-8s\t%-s\n",
				"NAME", "DESIRED", "UPDATED", "UPDATED_READY", "UPDATED_AVAILABLE", "READY", "TOTAL", "AGE")
		case "statefulsets.apps.kruise.io":
			fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-8s\t%-s\n",
				"NAME", "DESIRED", "CURRENT", "UPDATED", "READY", "AGE")
		case "daemonsets.apps.kruise.io":
			fmt.Fprintf(o.Out, "%-8s\t%-8s\t%-8s\t%-8s\t%-8s\t%-8s\t%-8s\t%-s\n",
				"NAME", "DESIRED", "CURRENT", "READY", "UP-TO-DATE", "AVAILABLE", "NODE SELECTOR", "AGE")
		case "rollouts.rollouts.kruise.io":
			fmt.Fprintf(o.Out, "%-20s\t%-8s\t%-12s\t%-12s\t%-40s\t%-s\n",
				"NAME", "STATUS", "CANARY_STEP", "CANARY_STATE", "MESSAGE", "AGE")
		case "broadcastjobs.apps.kruise.io":
			fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-8s\t%-s\n",
				"NAME", "DESIRED", "ACTIVE", "SUCCEEDED", "FAILED", "AGE")
		case "containerrecreaterequests.apps.kruise.io":
			fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-s\n",
				"NAME", "PHASE", "COMPLETED", "FAILED", "AGE")
		case "advancedcronjobs.apps.kruise.io":
			fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-8s\t%-s\n",
				"NAME", "SCHEDULE", "SUSPEND", "ACTIVE", "LAST SCHEDULE", "AGE")
		case "resourcedistributions.apps.kruise.io":
			fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-s\n",
				"NAME", "TARGETS", "SUCCEEDED", "FAILED", "AGE")
		case "uniteddeployments.apps.kruise.io":
			fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-8s\t%-s\n",
				"NAME", "DESIRED", "UPDATED", "READY", "AVAILABLE", "AGE")
		}

		for _, info := range infos {
			if err := o.printResourceInfo(info.Object, o.Resources[0]); err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *GetOptions) printResourceInfo(obj runtime.Object, resourceType string) error {
	metaObj, ok := obj.(v1.Object)
	if !ok {
		return fmt.Errorf("object does not implement v1.Object interface")
	}

	name := metaObj.GetName()
	age := time.Since(metaObj.GetCreationTimestamp().Time).Round(time.Second)

	// Print based on resource type
	switch resourceType {
	case "clonesets.apps.kruise.io":
		cloneset, ok := obj.(*kruiseappsv1alpha1.CloneSet)
		if !ok {
			return fmt.Errorf("object is not a CloneSet")
		}
		desired := cloneset.Spec.Replicas
		updated := cloneset.Status.UpdatedReplicas
		updatedReady := cloneset.Status.UpdatedReadyReplicas
		updatedAvailable := cloneset.Status.UpdatedAvailableReplicas
		ready := cloneset.Status.ReadyReplicas
		total := cloneset.Status.Replicas
		fmt.Fprintf(o.Out, "%-12s\t%-8d\t%-8d\t%-14d\t%-18d\t%-8d\t%-8d\t%-s\n",
			name, *desired, updated, updatedReady, updatedAvailable, ready, total, age)
	case "statefulsets.apps.kruise.io":
		statefulset, ok := obj.(*kruiseappsv1beta1.StatefulSet)
		if !ok {
			return fmt.Errorf("object is not a StatefulSet")
		}
		desired := statefulset.Spec.Replicas
		current := statefulset.Status.CurrentReplicas
		updated := statefulset.Status.UpdatedReplicas
		ready := statefulset.Status.ReadyReplicas
		fmt.Fprintf(o.Out, "%-12s\t%-8d\t%-8d\t%-8d\t%-8d\t%-s\n",
			name, *desired, current, updated, ready, age)
	case "daemonsets.apps.kruise.io":
		daemonset, ok := obj.(*kruiseappsv1alpha1.DaemonSet)
		if !ok {
			return fmt.Errorf("object is not a DaemonSet")
		}
		desired := daemonset.Status.DesiredNumberScheduled
		current := daemonset.Status.CurrentNumberScheduled
		ready := daemonset.Status.NumberReady
		updated := daemonset.Status.UpdatedNumberScheduled
		available := daemonset.Status.NumberAvailable
		nodeSelector := daemonset.Spec.Template.Spec.NodeSelector
		fmt.Fprintf(o.Out, "%-8s\t%-8d\t%-8d\t%-8d\t%-8d\t%-8d\t%-8s\t%-s\n",
			name, desired, current, ready, updated, available, fmt.Sprintf("%v", nodeSelector), age)
	case "rollouts.rollouts.kruise.io":
		rollout, ok := obj.(*rolloutv1beta1.Rollout)
		if !ok {
			rolloutV1alpha1, ok := obj.(*rolloutv1alpha1.Rollout)
			if !ok {
				return fmt.Errorf("object is not a Rollout")
			}
			status := rolloutV1alpha1.Status.Phase
			canaryStep := int32(0)
			canaryState := string(rolloutv1alpha1.CanaryStepStateCompleted)
			message := rolloutV1alpha1.Status.Message
			if rolloutV1alpha1.Status.CanaryStatus != nil {
				canaryStep = rolloutV1alpha1.Status.CanaryStatus.CurrentStepIndex
				canaryState = string(rolloutV1alpha1.Status.CanaryStatus.CurrentStepState)
			}
			fmt.Fprintf(o.Out, "%-20s\t%-8s\t%-12d\t%-12s\t%-40s\t%-s\n",
				name, status, canaryStep, canaryState, message, age)
			return nil
		}
		status := rollout.Status.Phase
		canaryStep := int32(0)
		canaryState := string(rolloutv1beta1.CanaryStepStateCompleted)
		message := rollout.Status.Message
		if rollout.Status.CanaryStatus != nil {
			canaryStep = rollout.Status.CanaryStatus.CurrentStepIndex
			canaryState = string(rollout.Status.CanaryStatus.CurrentStepState)
		}
		fmt.Fprintf(o.Out, "%-20s\t%-8s\t%-12d\t%-12s\t%-40s\t%-s\n",
			name, status, canaryStep, canaryState, message, age)
	case "broadcastjobs.apps.kruise.io":
		broadcastjob, ok := obj.(*kruiseappsv1alpha1.BroadcastJob)
		if !ok {
			return fmt.Errorf("object is not a BroadcastJob")
		}
		desired := broadcastjob.Spec.Parallelism
		active := broadcastjob.Status.Active
		successful := broadcastjob.Status.Succeeded
		failed := broadcastjob.Status.Failed
		fmt.Fprintf(o.Out, "%-12s\t%-8d\t%-8d\t%-8d\t%-8d\t%-s\n",
			name, *desired, active, successful, failed, age)
	case "containerrecreaterequests.apps.kruise.io":
		containerrecreaterequest, ok := obj.(*kruiseappsv1alpha1.ContainerRecreateRequest)
		if !ok {
			return fmt.Errorf("object is not a ContainerRecreateRequest")
		}
		phase := containerrecreaterequest.Status.Phase
		fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-s\n",
			name, phase, "-", "-", age)
	case "advancedcronjobs.apps.kruise.io":
		advancedcronjob, ok := obj.(*kruiseappsv1alpha1.AdvancedCronJob)
		if !ok {
			return fmt.Errorf("object is not a AdvancedCronJob")
		}
		schedule := advancedcronjob.Spec.Schedule
		suspend := advancedcronjob.Spec.Paused
		active := len(advancedcronjob.Status.Active)
		lastSchedule := advancedcronjob.Status.LastScheduleTime
		fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8v\t%-8d\t%-8s\t%-s\n",
			name, schedule, suspend, active, lastSchedule, age)
	case "resourcedistributions.apps.kruise.io":
		_, ok := obj.(*kruiseappsv1alpha1.ResourceDistribution)
		if !ok {
			return fmt.Errorf("object is not a ResourceDistribution")
		}
		fmt.Fprintf(o.Out, "%-12s\t%-8s\t%-8s\t%-8s\t%-s\n",
			name, "-", "-", "-", age)
	case "uniteddeployments.apps.kruise.io":
		uniteddeployment, ok := obj.(*kruiseappsv1alpha1.UnitedDeployment)
		if !ok {
			return fmt.Errorf("object is not a UnitedDeployment")
		}
		desired := uniteddeployment.Spec.Replicas
		updated := uniteddeployment.Status.UpdatedReplicas
		ready := uniteddeployment.Status.ReadyReplicas
		total := uniteddeployment.Status.Replicas
		fmt.Fprintf(o.Out, "%-12s\t%-8d\t%-8d\t%-8d\t%-8d\t%-s\n",
			name, *desired, updated, ready, total, age)
	}

	return nil
}
