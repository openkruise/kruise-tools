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

package describe

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	rolloutsv1beta1 "github.com/openkruise/kruise-rollout-api/client/clientset/versioned"
	rolloutsv1beta1types "github.com/openkruise/kruise-rollout-api/client/clientset/versioned/typed/rollouts/v1beta1"
	rolloutsapi "github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	internalapi "github.com/openkruise/kruise-tools/pkg/api"
	internalpolymorphichelpers "github.com/openkruise/kruise-tools/pkg/internal/polymorphichelpers"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

const (
	tableFormat = "%-19s%v\n"
)

var (
	rolloutLong = templates.LongDesc(i18n.T(`
		Get details about and visual representation of a rollout.`))

	rolloutExample = templates.Examples(`
		# Describe the rollout named rollout-demo within namespace default
		kubectl-kruise describe rollout rollout-demo/default

		# Watch for changes to the rollout named rollout-demo
		kubectl-kruise describe rollout rollout-demo/default -w`)
)

type DescribeRolloutOptions struct {
	genericclioptions.IOStreams
	Builder          func() *resource.Builder
	Namespace        string
	EnforceNamespace bool
	Resources        []string
	RolloutViewerFn  func(runtime.Object) (*rolloutsapi.Rollout, error)
	Watch            bool
	NoColor          bool
	TimeoutSeconds   int
	RolloutsClient   rolloutsv1beta1types.RolloutInterface
}

type WorkloadInfo struct {
	Name     string
	Kind     string
	Images   []string
	Replicas struct {
		Desired   int32
		Current   int32
		Updated   int32
		Ready     int32
		Available int32
	}
	Pod []struct {
		Name     string
		BatchID  string
		Status   string
		Ready    string
		Age      string
		Restarts string
		Revision string
	}
	CurrentRevision string
	UpdateRevision  string
}

func NewCmdDescribeRollout(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := &DescribeRolloutOptions{IOStreams: streams}
	cmd := &cobra.Command{
		Use:                   "rollout SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Get details about a rollout"),
		Long:                  rolloutLong,
		Example:               rolloutExample,
		Aliases:               []string{"rollouts", "ro"},
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, args))
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().BoolVarP(&o.Watch, "watch", "w", false, "Watch for changes to the rollout")
	cmd.Flags().BoolVar(&o.NoColor, "no-color", false, "If true, print output without color")
	cmd.Flags().IntVar(&o.TimeoutSeconds, "timeout", 0, "Timeout after specified seconds")

	return cmd
}

func (o *DescribeRolloutOptions) Complete(f cmdutil.Factory, args []string) error {
	var err error

	if len(args) == 0 {
		return fmt.Errorf("required rollout name not specified")
	}

	o.Resources = args
	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()

	if err != nil {
		return err
	}

	parts := strings.Split(args[0], "/")
	if len(parts) == 2 {
		o.Resources = []string{parts[0]}
		o.Namespace = parts[1]
	}

	o.RolloutViewerFn = internalpolymorphichelpers.RolloutViewerFn
	o.Builder = f.NewBuilder

	config, err := f.ToRESTConfig()
	if err != nil {
		return err
	}

	rolloutsClientset, err := rolloutsv1beta1.NewForConfig(config)
	if err != nil {
		return err
	}
	o.RolloutsClient = rolloutsClientset.RolloutsV1beta1().Rollouts(o.Namespace)

	return nil
}

func (o *DescribeRolloutOptions) Run() error {
	rolloutName := o.Resources[0]

	r := o.Builder().
		WithScheme(internalapi.GetScheme(), scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(o.Namespace).DefaultNamespace().
		ResourceNames("rollouts.rollouts.kruise.io", rolloutName).
		ContinueOnError().
		Latest().
		Flatten().
		Do()

	if err := r.Err(); err != nil {
		return err
	}

	if !o.Watch {
		return r.Visit(o.describeRollout)
	}

	return o.watchRollout(r)
}

func (o *DescribeRolloutOptions) describeRollout(info *resource.Info, err error) error {
	if err != nil {
		return err
	}

	rollout, err := o.RolloutViewerFn(info.Object)
	if err != nil {
		return err
	}

	o.printRolloutInfo(rollout)
	return nil
}

func (o *DescribeRolloutOptions) watchRollout(r *resource.Result) error {
	infos, err := r.Infos()
	if err != nil {
		return err
	}
	if len(infos) != 1 {
		return fmt.Errorf("watch is only supported on a single rollout")
	}
	info := infos[0]

	watcher, err := o.RolloutsClient.Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: "metadata.name=" + info.Name,
	})

	if err != nil {
		return err
	}
	defer watcher.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if o.TimeoutSeconds > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(o.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	return o.watchRolloutUpdates(ctx, watcher)
}

func (o *DescribeRolloutOptions) watchRolloutUpdates(ctx context.Context, watcher watch.Interface) error {
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil
			}
			if event.Type == watch.Added || event.Type == watch.Modified {
				if rollout, ok := event.Object.(*rolloutsapi.Rollout); ok {
					o.clearScreen()
					o.printRolloutInfo(rollout)
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (o *DescribeRolloutOptions) clearScreen() {
	fmt.Fprint(o.Out, "\033[2J\033[H")
}

func (o *DescribeRolloutOptions) GetResources(rollout *rolloutsapi.Rollout) (*WorkloadInfo, error) {
	resources := []string{rollout.Spec.WorkloadRef.Kind + "/" + rollout.Spec.WorkloadRef.Name}
	r := o.Builder().
		WithScheme(internalapi.GetScheme(), scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(o.Namespace).DefaultNamespace().
		ResourceTypeOrNameArgs(true, resources...).
		ContinueOnError().
		Latest().
		Flatten().
		Do()

	if err := r.Err(); err != nil {
		return nil, err
	}

	obj, err := r.Object()
	if err != nil {
		return nil, err
	}

	workloadInfo := &WorkloadInfo{}
	objValue := reflect.ValueOf(obj).Elem()
	workloadInfo.Name = objValue.FieldByName("Name").String()
	workloadInfo.Kind = objValue.Type().Name()

	podTemplateSpec := objValue.FieldByName("Spec").FieldByName("Template").FieldByName("Spec")
	containers := podTemplateSpec.FieldByName("Containers")
	for i := 0; i < containers.Len(); i++ {
		container := containers.Index(i)
		workloadInfo.Images = append(workloadInfo.Images, container.FieldByName("Image").String())
	}

	// Deployment,StatefulSet,CloneSet,Advanced StatefulSet,Advanced DaemonSet

	switch o := obj.(type) {
	case *appsv1.Deployment:
		workloadInfo.Replicas.Desired = *o.Spec.Replicas
		workloadInfo.Replicas.Current = o.Status.Replicas
		workloadInfo.Replicas.Updated = o.Status.UpdatedReplicas
		workloadInfo.Replicas.Ready = o.Status.ReadyReplicas
		workloadInfo.Replicas.Available = o.Status.AvailableReplicas
	case *appsv1.StatefulSet:
		workloadInfo.Replicas.Desired = *o.Spec.Replicas
		workloadInfo.Replicas.Current = o.Status.Replicas
		workloadInfo.Replicas.Updated = o.Status.UpdatedReplicas
		workloadInfo.Replicas.Ready = o.Status.ReadyReplicas
		workloadInfo.Replicas.Available = o.Status.AvailableReplicas
	case *kruiseappsv1alpha1.DaemonSet:
		workloadInfo.Replicas.Desired = o.Spec.BurstReplicas.IntVal
		workloadInfo.Replicas.Current = o.Status.CurrentNumberScheduled
		workloadInfo.Replicas.Updated = o.Status.UpdatedNumberScheduled
		workloadInfo.Replicas.Ready = o.Status.NumberReady
		workloadInfo.Replicas.Available = o.Status.NumberAvailable
	case *kruiseappsv1beta1.StatefulSet:
		workloadInfo.Replicas.Desired = *o.Spec.Replicas
		workloadInfo.Replicas.Current = o.Status.Replicas
		workloadInfo.Replicas.Updated = o.Status.UpdatedReplicas
		workloadInfo.Replicas.Ready = o.Status.ReadyReplicas
		workloadInfo.Replicas.Available = o.Status.AvailableReplicas
	case *kruiseappsv1alpha1.CloneSet:
		workloadInfo.Replicas.Desired = *o.Spec.Replicas
		workloadInfo.Replicas.Current = o.Status.Replicas
		workloadInfo.Replicas.Updated = o.Status.UpdatedReplicas
		workloadInfo.Replicas.Ready = o.Status.ReadyReplicas
		workloadInfo.Replicas.Available = o.Status.AvailableReplicas
	default:
		return nil, fmt.Errorf("unsupported workload kind %T", obj)
	}

	workloadInfo.CurrentRevision = rollout.Status.CanaryStatus.StableRevision
	workloadInfo.UpdateRevision = rollout.Status.CanaryStatus.CanaryRevision

	var labelSelectorParam string
	switch obj.(type) {
	case *appsv1.Deployment:
		labelSelectorParam = "pod-template-hash"
	default:
		labelSelectorParam = "controller-revision-hash"
	}

	SelectorParam := rollout.Status.CanaryStatus.PodTemplateHash
	if SelectorParam == "" {
		SelectorParam = rollout.Status.CanaryStatus.StableRevision
	}

	// Fetch pods
	r2 := o.Builder().
		WithScheme(internalapi.GetScheme(), scheme.Scheme.PrioritizedVersionsAllGroups()...).
		NamespaceParam(o.Namespace).DefaultNamespace().
		ResourceTypes("pods").
		LabelSelectorParam(fmt.Sprintf("%s=%s", labelSelectorParam, SelectorParam)).
		Latest().
		Flatten().
		Do()

	if err := r2.Err(); err != nil {
		return nil, err
	}

	err = r2.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		pod, ok := info.Object.(*corev1.Pod)
		if !ok {
			return fmt.Errorf("expected *corev1.Pod, got %T", info.Object)
		}

		podInfo := struct {
			Name     string
			BatchID  string
			Status   string
			Ready    string
			Age      string
			Restarts string
			Revision string
		}{
			Name:     pod.Name,
			BatchID:  pod.Labels["rollouts.kruise.io/rollout-batch-id"],
			Status:   string(pod.Status.Phase),
			Age:      duration.HumanDuration(time.Since(pod.CreationTimestamp.Time)),
			Restarts: "0",
		}

		if pod.DeletionTimestamp != nil {
			podInfo.Status = "Terminating"
		}

		if len(pod.Status.ContainerStatuses) > 0 {
			restartCount := 0
			for _, containerStatus := range pod.Status.ContainerStatuses {
				restartCount += int(containerStatus.RestartCount)
			}

			podInfo.Restarts = fmt.Sprintf("%v", strconv.Itoa(int(restartCount)))
		}

		// Calculate ready status
		readyContainers := 0
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Ready {
				readyContainers++
			}
		}
		podInfo.Ready = fmt.Sprintf("%d/%d", readyContainers, len(pod.Spec.Containers))

		// Calculate revision
		if pod.Labels["pod-template-hash"] != "" {
			podInfo.Revision = pod.Labels["pod-template-hash"]
		} else if pod.Labels["controller-revision-hash"] != "" {
			podInfo.Revision = pod.Labels["controller-revision-hash"]
		}

		workloadInfo.Pod = append(workloadInfo.Pod, podInfo)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort pods by batch ID and ready count
	sort.Slice(workloadInfo.Pod, func(i, j int) bool {
		if workloadInfo.Pod[i].BatchID != workloadInfo.Pod[j].BatchID {
			return workloadInfo.Pod[i].BatchID < workloadInfo.Pod[j].BatchID
		}

		iReady := strings.Split(workloadInfo.Pod[i].Ready, "/")
		jReady := strings.Split(workloadInfo.Pod[j].Ready, "/")

		iReadyCount, _ := strconv.Atoi(iReady[0])
		jReadyCount, _ := strconv.Atoi(jReady[0])

		return iReadyCount > jReadyCount
	})

	return workloadInfo, nil
}

func (o *DescribeRolloutOptions) colorizeIcon(phase string) string {
	if o.NoColor || phase == "" {
		return ""
	}
	switch phase {
	case string(rolloutsapi.RolloutPhaseHealthy), string(corev1.PodRunning):
		return "\033[32m✔\033[0m"
	case string(rolloutsapi.RolloutPhaseProgressing), string(corev1.PodPending):
		return "\033[33m⚠\033[0m"
	case string(rolloutsapi.RolloutPhaseDisabled), string(corev1.PodUnknown), string(rolloutsapi.RolloutPhaseTerminating), string(rolloutsapi.RolloutPhaseDisabling), string(corev1.PodFailed):
		return "\033[31m✘\033[0m"
	case string(rolloutsapi.RolloutPhaseInitial):
		return "\033[33m⚠\033[0m"
	default:
		return ""
	}
}

func (o *DescribeRolloutOptions) printTrafficRouting(trafficRouting []rolloutsapi.TrafficRoutingRef) {
	fmt.Fprint(o.Out, "Traffic Routings:\n")
	for _, trafficRouting := range trafficRouting {
		fmt.Fprintf(o.Out, tableFormat, "  -  Service: ", trafficRouting.Service)
		if trafficRouting.Ingress != nil {
			fmt.Fprintln(o.Out, `     Ingress: `)
			fmt.Fprintf(o.Out, tableFormat, "      classType: ", trafficRouting.Ingress.ClassType)
			fmt.Fprintf(o.Out, tableFormat, "      name: ", trafficRouting.Ingress.Name)
		}
		if trafficRouting.Gateway != nil {
			fmt.Fprintln(o.Out, `     Gateway: `)
			fmt.Fprintf(o.Out, tableFormat, "      HttpRouteName: ", trafficRouting.Gateway.HTTPRouteName)
		}
		if trafficRouting.CustomNetworkRefs != nil {
			fmt.Fprintln(o.Out, `     CustomNetworkRefs: `)
			for _, customNetworkRef := range trafficRouting.CustomNetworkRefs {
				fmt.Fprintf(o.Out, tableFormat, "      name: ", customNetworkRef.Name)
				fmt.Fprintf(o.Out, tableFormat, "      kind: ", customNetworkRef.Kind)
				fmt.Fprintf(o.Out, tableFormat, "      apiVersion: ", customNetworkRef.APIVersion)
			}
		}
	}
}

func (o *DescribeRolloutOptions) printRolloutInfo(rollout *rolloutsapi.Rollout) {
	// Name, Namespace, Status, Message
	fmt.Fprintf(o.Out, tableFormat, "Name:", rollout.Name)
	fmt.Fprintf(o.Out, tableFormat, "Namespace:", rollout.Namespace)
	if rollout.Status.ObservedGeneration == rollout.GetObjectMeta().GetGeneration() {
		fmt.Fprintf(o.Out, tableFormat, "Status:", o.colorizeIcon(string(rollout.Status.Phase))+" "+string(rollout.Status.Phase))

		if rollout.Status.Message != "" {
			fmt.Fprintf(o.Out, tableFormat, "Message:", rollout.Status.Message)
		}
	}

	// Strategy
	fmt.Fprintf(o.Out, tableFormat, "Strategy:", "Canary")

	if canary := rollout.Spec.Strategy.Canary; canary != nil {
		fmt.Fprintf(o.Out, tableFormat, " Step:", strconv.Itoa(int(rollout.Status.CanaryStatus.CurrentStepIndex))+"/"+strconv.Itoa(len(canary.Steps)))
		fmt.Fprint(o.Out, " Steps:\n")
		currentStepIndex := int(rollout.Status.CanaryStatus.CurrentStepIndex)
		for i, step := range canary.Steps {
			isCurrentStep := (i + 1) == currentStepIndex
			if isCurrentStep {
				fmt.Fprint(o.Out, "\033[32m")
			}

			if step.Replicas != nil {
				fmt.Fprintf(o.Out, tableFormat, "  -  Replicas: ", step.Replicas)
			}
			if step.Traffic != nil {
				fmt.Fprintf(o.Out, tableFormat, "     Traffic: ", *step.Traffic)
			}

			if step.Matches != nil {
				fmt.Fprintln(o.Out, "     Matches: ")
				for _, match := range step.Matches {
					fmt.Fprintln(o.Out, "      - Headers: ")
					for _, path := range match.Headers {
						fmt.Fprintf(o.Out, tableFormat, "       - Name:", path.Name)
						fmt.Fprintf(o.Out, tableFormat, "         Value:", path.Value)
						fmt.Fprintf(o.Out, tableFormat, "         Type:", *path.Type)
					}
				}
			}

			if isCurrentStep {
				fmt.Fprintf(o.Out, tableFormat, "     State:", rollout.Status.CanaryStatus.CurrentStepState)
				fmt.Fprint(o.Out, "\033[0m") // Reset color
			}
		}

		// Traffic Routings
		if len(rollout.Spec.Strategy.Canary.TrafficRoutings) > 0 {
			o.printTrafficRouting(rollout.Spec.Strategy.Canary.TrafficRoutings)
		}

		if rollout.Spec.Strategy.Canary.TrafficRoutingRef != "" {
			// r := o.Builder().
			// 	WithScheme(internalapi.GetScheme(), scheme.Scheme.PrioritizedVersionsAllGroups()...).
			// 	NamespaceParam(o.Namespace).DefaultNamespace().
			// 	ResourceNames("trafficrouting.rollouts.kruise.io", rollout.Spec.Strategy.Canary.TrafficRoutingRef).
			// 	ContinueOnError().
			// 	Latest().
			// 	Flatten().
			// 	Do()

			// if err := r.Err(); err != nil {
			// 	fmt.Fprintf(o.Out, "Error getting TrafficRoutingRef: %v\n", err)
			// 	return
			// }

		}
	}

	// Workload
	info, err := o.GetResources(rollout)
	if err != nil {
		fmt.Fprintf(o.Out, "Error getting resources: %v\n", err)
		return
	}

	// Images
	for i, image := range info.Images {
		if i == 0 {
			fmt.Fprintf(o.Out, tableFormat, "Images:", image)
		} else {
			fmt.Fprintf(o.Out, tableFormat, "", image)
		}
	}

	// Revisions
	fmt.Fprintf(o.Out, tableFormat, "Current Revision:", info.CurrentRevision)
	fmt.Fprintf(o.Out, tableFormat, "Update Revision:", info.UpdateRevision)

	// Replicas
	if rollout.Status.ObservedGeneration == rollout.GetObjectMeta().GetGeneration() {
		fmt.Fprint(o.Out, "Replicas:\n")
		fmt.Fprintf(o.Out, tableFormat, " Desired:", info.Replicas.Desired)
		fmt.Fprintf(o.Out, tableFormat, " Updated:", info.Replicas.Updated)
		fmt.Fprintf(o.Out, tableFormat, " Current:", info.Replicas.Current)
		fmt.Fprintf(o.Out, tableFormat, " Ready:", info.Replicas.Ready)
		fmt.Fprintf(o.Out, tableFormat, " Available:", info.Replicas.Available)
	}

	if len(info.Pod) > 0 {
		w := tabwriter.NewWriter(o.Out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tREADY\tBATCH ID\tREVISION\tAGE\tRESTARTS\tSTATUS")

		for _, pod := range info.Pod {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s %s\n",
				pod.Name,
				pod.Ready,
				pod.BatchID,
				pod.Revision,
				pod.Age,
				pod.Restarts,
				o.colorizeIcon(pod.Status),
				pod.Status,
			)
		}
		w.Flush()
	}

}
