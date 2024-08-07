/*
Copyright 2016 The Kubernetes Authors.

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

package set

import (
	"fmt"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	"github.com/openkruise/kruise-tools/pkg/internal/polymorphichelpers"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// SetImageOptions ImageOptions is the start of the data required to perform the operation.  As new fields are added, add them here instead of
// referencing the cmd.Flags()
type SetImageOptions struct {
	resource.FilenameOptions

	PrintFlags  *genericclioptions.PrintFlags
	RecordFlags *genericclioptions.RecordFlags

	Infos          []*resource.Info
	Selector       string
	DryRunStrategy cmdutil.DryRunStrategy
	All            bool
	Output         string
	Local          bool
	ResolveImage   ImageResolver

	PrintObj printers.ResourcePrinterFunc
	Recorder genericclioptions.Recorder

	UpdatePodSpecForObject polymorphichelpers.UpdatePodSpecForObjectFunc
	Resources              []string
	ContainerImages        map[string]string

	genericclioptions.IOStreams
}

var (
	imageResources = `
  	pod (po), deployment (deploy), statefulset (sts), daemonset (ds), replicaset (rs), cloneset (clone), statefulset.apps.kruise.io (asts), daemonset.apps.kruise.io (ads)`

	imageLong = templates.LongDesc(`
		Update existing container image(s) of resources.

		Possible resources include (case insensitive):
		` + imageResources)

	imageExample = templates.Examples(`
		# Set a deployment's nginx container image to 'nginx:1.9.1', and its busybox container image to 'busybox'.
		kubectl-kruise set image cloneset/sample busybox=busybox nginx=nginx:1.9.1

		# Update all deployments' and rc's nginx container's image to 'nginx:1.9.1'
		kubectl-kruise set image cloneset,rc nginx=nginx:1.9.1 --all

		# Update image of all containers of cloneset sample to 'nginx:1.9.1'
		kubectl-kruise set image cloneset sample *=nginx:1.9.1

		# Print result (in yaml format) of updating nginx container image from local file, without hitting the server
		kubectl-kruise set image -f path/to/file.yaml nginx=nginx:1.9.1 --local -o yaml`)
)

// NewImageOptions returns an initialized SetImageOptions instance
func NewImageOptions(streams genericclioptions.IOStreams) *SetImageOptions {
	return &SetImageOptions{
		PrintFlags:  genericclioptions.NewPrintFlags("image updated").WithTypeSetter(scheme.Scheme),
		RecordFlags: genericclioptions.NewRecordFlags(),

		Recorder: genericclioptions.NoopRecorder{},

		IOStreams: streams,
	}
}

// NewCmdImage returns an initialized Command instance for the 'set image' sub command
func NewCmdImage(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewImageOptions(streams)

	cmd := &cobra.Command{
		Use:                   "image (-f FILENAME | TYPE NAME) CONTAINER_NAME_1=CONTAINER_IMAGE_1 ... CONTAINER_NAME_N=CONTAINER_IMAGE_N",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Update image of a pod template"),
		Long:                  imageLong,
		Example:               imageExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	o.PrintFlags.AddFlags(cmd)
	o.RecordFlags.AddFlags(cmd)

	usage := "identifying the resource to get from a server."
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
	cmd.Flags().BoolVar(&o.All, "all", o.All, "Select all resources, including uninitialized ones, in the namespace of the specified resource types")
	cmd.Flags().StringVarP(&o.Selector, "selector", "l", o.Selector, "Selector (label query) to filter on, not including uninitialized ones, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().BoolVar(&o.Local, "local", o.Local, "If true, set image will NOT contact api-server but run locally.")
	cmdutil.AddDryRunFlag(cmd)
	return cmd
}

// Complete completes all required options
func (o *SetImageOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	var err error

	err = o.RecordFlags.Complete(cmd)
	if err != nil {
		return err
	}

	o.Recorder, err = o.RecordFlags.ToRecorder()
	if err != nil {
		return err
	}

	o.UpdatePodSpecForObject = polymorphichelpers.UpdatePodSpecForObjectFn
	o.DryRunStrategy, err = cmdutil.GetDryRunStrategy(cmd)
	if err != nil {
		return err
	}
	o.Output = cmdutil.GetFlagString(cmd, "output")
	o.ResolveImage = resolveImageFunc

	cmdutil.PrintFlagsWithDryRunStrategy(o.PrintFlags, o.DryRunStrategy)
	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}

	o.PrintObj = printer.PrintObj

	cmdNamespace, enforceNamespace, err := f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	o.Resources, o.ContainerImages, err = getResourcesAndImages(args)
	if err != nil {
		return err
	}

	builder := f.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		LocalParam(o.Local).
		ContinueOnError().
		NamespaceParam(cmdNamespace).DefaultNamespace().
		FilenameParam(enforceNamespace, &o.FilenameOptions).
		Flatten()

	if !o.Local {
		builder.LabelSelectorParam(o.Selector).
			ResourceTypeOrNameArgs(o.All, o.Resources...).
			Latest()
	} else {
		// if a --local flag was provided, and a resource was specified in the form
		// <resource>/<name>, fail immediately as --local cannot query the api server
		// for the specified resource.
		if len(o.Resources) > 0 {
			return resource.LocalResourceError
		}
	}

	o.Infos, err = builder.Do().Infos()
	if err != nil {
		return err
	}

	return nil
}

// Validate makes sure provided values in SetImageOptions are valid
func (o *SetImageOptions) Validate() error {
	var errors []error
	if o.All && len(o.Selector) > 0 {
		errors = append(errors, fmt.Errorf("cannot set --all and --selector at the same time"))
	}
	if len(o.Resources) < 1 && cmdutil.IsFilenameSliceEmpty(o.Filenames, o.Kustomize) {
		errors = append(errors, fmt.Errorf("one or more resources must be specified as <resource> <name> or <resource>/<name>"))
	}
	if len(o.ContainerImages) < 1 {
		errors = append(errors, fmt.Errorf("at least one image update is required"))
	} else if len(o.ContainerImages) > 1 && hasWildcardKey(o.ContainerImages) {
		errors = append(errors, fmt.Errorf("all containers are already specified by *, but saw more than one container_name=container_image pairs"))
	}
	if o.Local && o.DryRunStrategy == cmdutil.DryRunServer {
		errors = append(errors, fmt.Errorf("cannot specify --local and --dry-run=server - did you mean --dry-run=client?"))
	}
	return utilerrors.NewAggregate(errors)
}

// Run performs the execution of 'set image' sub command
func (o *SetImageOptions) Run() error {
	var allErrs []error

	patches := CalculatePatches(o.Infos, scheme.DefaultJSONEncoder(), func(obj runtime.Object) ([]byte, error) {
		_, err := o.UpdatePodSpecForObject(obj, func(spec *corev1.PodSpec) error {
			for name, image := range o.ContainerImages {
				resolvedImageName, err := o.ResolveImage(image)
				if err != nil {
					allErrs = append(allErrs, fmt.Errorf("error: unable to resolve image %q for container %q: %v", image, name, err))
					if name == "*" {
						break
					}
					continue
				}
				var initContainerFound, containerFound bool
				// Check if the type is kruiseappsv1alpha1.SidecarSet, and if the placeholder is nil.
				if t, ok := obj.(*kruiseappsv1alpha1.SidecarSet); ok && spec == nil {
					initContainerFound = setSideCarImage(t.Spec.InitContainers, name, resolvedImageName)
					containerFound = setSideCarImage(t.Spec.Containers, name, resolvedImageName)
				} else {
					initContainerFound = setImage(spec.InitContainers, name, resolvedImageName)
					containerFound = setImage(spec.Containers, name, resolvedImageName)
				}
				if !containerFound && !initContainerFound {
					allErrs = append(allErrs, fmt.Errorf("error: unable to find container named %q", name))
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		// record this change (for rollout history)
		if err := o.Recorder.Record(obj); err != nil {
			klog.V(4).Infof("error recording current command: %v", err)
		}

		return runtime.Encode(scheme.DefaultJSONEncoder(), obj)
	})

	for _, patch := range patches {
		info := patch.Info
		if patch.Err != nil {
			name := info.ObjectName()
			allErrs = append(allErrs, fmt.Errorf("error: %s %v\n", name, patch.Err))
			continue
		}

		// no changes
		if string(patch.Patch) == "{}" || len(patch.Patch) == 0 {
			continue
		}

		if o.Local || o.DryRunStrategy == cmdutil.DryRunClient {
			if err := o.PrintObj(info.Object, o.Out); err != nil {
				allErrs = append(allErrs, err)
			}
			continue
		}

		// patch the change
		actual, err := resource.
			NewHelper(info.Client, info.Mapping).
			DryRun(o.DryRunStrategy == cmdutil.DryRunServer).
			Patch(info.Namespace, info.Name, types.MergePatchType, patch.After, nil)
		if err != nil {
			allErrs = append(allErrs, fmt.Errorf("failed to patch image update to pod template: %v", err))
			continue
		}

		if err := o.PrintObj(actual, o.Out); err != nil {
			allErrs = append(allErrs, err)
		}

	}
	return utilerrors.NewAggregate(allErrs)
}

func setImage(containers []corev1.Container, containerName string, image string) bool {
	containerFound := false
	// Find the container to update, and update its image
	for i, c := range containers {
		if c.Name == containerName || containerName == "*" {
			containerFound = true
			containers[i].Image = image
		}
	}
	return containerFound
}

// setSideCarImage
func setSideCarImage(containers []kruiseappsv1alpha1.SidecarContainer, containerName string, image string) bool {
	containerFound := false
	// Find the container to update, and update its image
	for i, c := range containers {
		if c.Name == containerName || containerName == "*" {
			containerFound = true
			containers[i].Image = image
		}
	}
	return containerFound
}

// getResourcesAndImages retrieves resources and container name:images pair from given args
func getResourcesAndImages(args []string) (resources []string, containerImages map[string]string, err error) {
	pairType := "image"
	resources, imageArgs, err := cmdutil.GetResourcesAndPairs(args, pairType)
	if err != nil {
		return
	}
	containerImages, _, err = cmdutil.ParsePairs(imageArgs, pairType, false)
	return
}

func hasWildcardKey(containerImages map[string]string) bool {
	_, ok := containerImages["*"]
	return ok
}

// ImageResolver is a func that receives an image name, and
// resolves it to an appropriate / compatible image name.
// Adds flexibility for future image resolving methods.
type ImageResolver func(in string) (string, error)

// implements ImageResolver
func resolveImageFunc(in string) (string, error) {
	return in, nil
}
