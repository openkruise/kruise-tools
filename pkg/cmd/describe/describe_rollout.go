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

package describe

import (
	"context"
	"fmt"
	"reflect"
	"time"

	rolloutsv1beta1 "github.com/openkruise/kruise-rollout-api/client/clientset/versioned"
	rolloutsv1beta1types "github.com/openkruise/kruise-rollout-api/client/clientset/versioned/typed/rollouts/v1beta1"
	rolloutsapi "github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	internalapi "github.com/openkruise/kruise-tools/pkg/api"
	internalpolymorphichelpers "github.com/openkruise/kruise-tools/pkg/internal/polymorphichelpers"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

const (
	tableFormat = "%-17s%v\n"
)

var (
	rolloutLong = templates.LongDesc(i18n.T(`
		Get details about and visual representation of a rollout.`))

	rolloutExample = templates.Examples(`
		# Describe the rollout named rollout-demo
		kubectl-kruise describe rollout rollout-demo

		# Watch for changes to the rollout named rollout-demo
		kubectl-kruise describe rollout rollout-demo -w`)
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
			cmdutil.CheckErr(o.Validate())
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
	o.Resources = args
	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
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

func (o *DescribeRolloutOptions) Validate() error {
	if len(o.Resources) == 0 {
		return fmt.Errorf("required rollout name is not specified")
	}
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

func (o *DescribeRolloutOptions) GetResources(rollout *rolloutsapi.RolloutSpec) (*WorkloadInfo, error) {
    resources := []string{rollout.WorkloadRef.Kind + "/" + rollout.WorkloadRef.Name}
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

	info := &WorkloadInfo{}
	objValue := reflect.ValueOf(obj).Elem()
    info.Name = objValue.FieldByName("Name").String()
    info.Kind = objValue.Type().Name()

    podTemplateSpec := objValue.FieldByName("Spec").FieldByName("Template").FieldByName("Spec")
    containers := podTemplateSpec.FieldByName("Containers")
    for i := 0; i < containers.Len(); i++ {
        container := containers.Index(i)
        info.Images = append(info.Images, container.FieldByName("Image").String())
    }

    spec := objValue.FieldByName("Spec")
    status := objValue.FieldByName("Status")

    if replicas := spec.FieldByName("Replicas"); replicas.IsValid() && !replicas.IsNil() {
        info.Replicas.Desired = int32(replicas.Elem().Int())
    }

    info.Replicas.Current = int32(status.FieldByName("Replicas").Int())
    info.Replicas.Updated = int32(status.FieldByName("UpdatedReplicas").Int())
    info.Replicas.Ready = int32(status.FieldByName("ReadyReplicas").Int())
    info.Replicas.Available = int32(status.FieldByName("AvailableReplicas").Int())

    return info, nil
}

func (o *DescribeRolloutOptions) colorizeIcon(phase string) string {
	if o.NoColor || phase == "" {
		return ""
	}
	switch phase {
	case string(rolloutsapi.RolloutPhaseHealthy):
		return "\033[32m✔\033[0m"
	case string(rolloutsapi.RolloutPhaseProgressing):
		return "\033[33m⚠\033[0m"
	case string(rolloutsapi.RolloutPhaseDisabled), string(rolloutsapi.RolloutPhaseTerminating), string(rolloutsapi.RolloutPhaseDisabling):
		return "\033[31m✘\033[0m"
	case string(rolloutsapi.RolloutPhaseInitial):
		return "\033[33m⚠\033[0m"
	default:
		return ""
	}
}

func (o *DescribeRolloutOptions) printRolloutInfo(rollout *rolloutsapi.Rollout) {
	fmt.Fprintf(o.Out, tableFormat, "Name:", rollout.Name)
	fmt.Fprintf(o.Out, tableFormat, "Namespace:", rollout.Namespace)
	fmt.Fprintf(o.Out, tableFormat, "Status:", o.colorizeIcon(string(rollout.Status.Phase))+" "+string(rollout.Status.Phase))
	if rollout.Status.Message != "" {
		fmt.Fprintf(o.Out, tableFormat, "Message:", rollout.Status.Message)
	}

	fmt.Fprintf(o.Out, tableFormat, "Strategy:", "Canary")

	if canary := rollout.Spec.Strategy.Canary; canary != nil {
		fmt.Fprintf(o.Out, tableFormat, " Step:", canary.Steps[0].Replicas)
		fmt.Fprintf(o.Out, tableFormat, " SetWeight:", canary.Steps[1].Replicas)
		fmt.Fprintf(o.Out, tableFormat, " ActualWeight:", canary.Steps[2].Replicas)
	}

	info, err := o.GetResources(&rollout.Spec)
	if err != nil {
		fmt.Fprintf(o.Out, "Error getting resources: %v\n", err)
		return
	}

	for i, image := range info.Images {
		if i == 0 {
			fmt.Fprintf(o.Out, tableFormat, "Images:", image)
		} else {
			fmt.Fprintf(o.Out, tableFormat, "", image)
		}
	}

	
	fmt.Fprint(o.Out, "Replicas:\n")
	fmt.Fprintf(o.Out, tableFormat, " Desired:", info.Replicas.Desired)
	fmt.Fprintf(o.Out, tableFormat, " Updated:", info.Replicas.Updated)
	fmt.Fprintf(o.Out, tableFormat, " Ready:", info.Replicas.Ready)
	fmt.Fprintf(o.Out, tableFormat, " Available:", info.Replicas.Available)
}