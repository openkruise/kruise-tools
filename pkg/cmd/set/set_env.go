/*
Copyright 2017 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	"github.com/openkruise/kruise-tools/pkg/api"
	"github.com/openkruise/kruise-tools/pkg/internal/polymorphichelpers"
	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	envutil "k8s.io/kubectl/pkg/cmd/set/env"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	validEnvNameRegexp = regexp.MustCompile("[^a-zA-Z0-9_]")
	envResources       = `
  	pod (po), deployment (deploy), statefulset (sts), daemonset (ds), replicaset (rs), cloneset (clone), statefulset.apps.kruise.io (asts), daemonset.apps.kruise.io (ads)`

	envLong = templates.LongDesc(`
		Update environment variables on a pod template.

		List environment variable definitions in one or more pods, pod templates.
		Add, update, or remove container environment variable definitions in one or
		more pod templates (within replication controllers or deployment configurations).
		View or modify the environment variable definitions on all containers in the
		specified pods or pod templates, or just those that match a wildcard.

		If "--env -" is passed, environment variables can be read from STDIN using the standard env
		syntax.

		Possible resources include (case insensitive):
		` + envResources)

	envExample = templates.Examples(`
      # Update cloneset 'sample' with a new environment variable
	  kubectl-kruise set env cloneset/sample STORAGE_DIR=/local

	  # List the environment variables defined on a cloneset 'sample'
	  kubectl-kruise set env cloneset/sample --list

	  # List the environment variables defined on all pods
	  kubectl-kruise set env pods --all --list

	  # Output modified cloneset in YAML, and does not alter the object on the server
	  kubectl-kruise set env cloneset/sample STORAGE_DIR=/data -o yaml

	  # Update all containers in all replication controllers in the project to have ENV=prod
	  kubectl-kruise set env rc --all ENV=prod

	  # Import environment from a secret
	  kubectl-kruise set env --from=secret/mysecret cloneset/sample

	  # Import environment from a config map with a prefix
	  kubectl-kruise set env --from=configmap/myconfigmap --prefix=MYSQL_ cloneset/sample

      # Import specific keys from a config map
      kubectl-kruise set env --keys=my-example-key --from=configmap/myconfigmap cloneset/sample

	  # Remove the environment variable ENV from container 'c1' in all deployment configs
	  kubectl-kruise set env clonesets --all --containers="c1" ENV-

	  # Remove the environment variable ENV from a deployment definition on disk and
	  # update the deployment config on the server
	  kubectl-kruise set env -f deploy.json ENV-

	  # Set some of the local shell environment into a deployment config on the server
	  env | grep RAILS_ | kubectl-kruise set env -e - cloneset/sample`)
)

// EnvOptions holds values for 'set env' command-lone options
type EnvOptions struct {
	PrintFlags *genericclioptions.PrintFlags
	resource.FilenameOptions

	EnvParams         []string
	All               bool
	Resolve           bool
	List              bool
	Local             bool
	Overwrite         bool
	ContainerSelector string
	Selector          string
	From              string
	Prefix            string
	Keys              []string

	PrintObj printers.ResourcePrinterFunc

	envArgs                []string
	resources              []string
	output                 string
	dryRunStrategy         cmdutil.DryRunStrategy
	builder                func() *resource.Builder
	updatePodSpecForObject polymorphichelpers.UpdatePodSpecForObjectFunc
	namespace              string
	enforceNamespace       bool
	clientset              *kubernetes.Clientset
	resRef                 api.ResourceRef

	genericclioptions.IOStreams
}

// NewEnvOptions returns an EnvOptions indicating all containers in the selected
// pod templates are selected by default and allowing environment to be overwritten
func NewEnvOptions(streams genericclioptions.IOStreams) *EnvOptions {
	return &EnvOptions{
		PrintFlags: genericclioptions.NewPrintFlags("env updated").WithTypeSetter(scheme.Scheme),

		ContainerSelector: "*",
		Overwrite:         true,

		IOStreams: streams,
	}
}

// NewCmdEnv implements the OpenShift cli env command
func NewCmdEnv(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewEnvOptions(streams)
	cmd := &cobra.Command{
		Use:                   "env RESOURCE/NAME KEY_1=VAL_1 ... KEY_N=VAL_N",
		DisableFlagsInUseLine: true,
		Short:                 "Update environment variables on a pod template",
		Long:                  envLong,
		Example:               fmt.Sprintf(envExample),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.RunEnv())
		},
	}
	usage := "the resource to update the env"
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
	cmd.Flags().StringVarP(&o.ContainerSelector, "containers", "c", o.ContainerSelector, "The names of containers in the selected pod templates to change - may use wildcards")
	cmd.Flags().StringVarP(&o.From, "from", "", "", "The name of a resource from which to inject environment variables")
	cmd.Flags().StringVarP(&o.Prefix, "prefix", "", "", "Prefix to append to variable names")
	cmd.Flags().StringArrayVarP(&o.EnvParams, "env", "e", o.EnvParams, "Specify a key-value pair for an environment variable to set into each container.")
	cmd.Flags().StringSliceVarP(&o.Keys, "keys", "", o.Keys, "Comma-separated list of keys to import from specified resource")
	cmd.Flags().BoolVar(&o.List, "list", o.List, "If true, display the environment and any changes in the standard format. this flag will removed when we have kubectl view env.")
	cmd.Flags().BoolVar(&o.Resolve, "resolve", o.Resolve, "If true, show secret or configmap references when listing variables")
	cmd.Flags().StringVarP(&o.Selector, "selector", "l", o.Selector, "Selector (label query) to filter on")
	cmd.Flags().BoolVar(&o.Local, "local", o.Local, "If true, set env will NOT contact api-server but run locally.")
	cmd.Flags().BoolVar(&o.All, "all", o.All, "If true, select all resources in the namespace of the specified resource types")
	cmd.Flags().BoolVar(&o.Overwrite, "overwrite", o.Overwrite, "If true, allow environment to be overwritten, otherwise reject updates that overwrite existing environment.")

	o.PrintFlags.AddFlags(cmd)

	cmdutil.AddDryRunFlag(cmd)
	return cmd
}

func validateNoOverwrites(existing []v1.EnvVar, env []v1.EnvVar) error {
	for _, e := range env {
		if current, exists := findEnv(existing, e.Name); exists && current.Value != e.Value {
			return fmt.Errorf("'%s' already has a value (%s), and --overwrite is false", current.Name, current.Value)
		}
	}
	return nil
}

func keyToEnvName(key string) string {
	return strings.ToUpper(validEnvNameRegexp.ReplaceAllString(key, "_"))
}

func contains(key string, keyList []string) bool {
	if len(keyList) == 0 {
		return true
	}

	for _, k := range keyList {
		if k == key {
			return true
		}
	}
	return false
}

// Complete completes all required options
func (o *EnvOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if o.All && len(o.Selector) > 0 {
		return fmt.Errorf("cannot set --all and --selector at the same time")
	}
	ok := false
	o.resources, o.envArgs, ok = envutil.SplitEnvironmentFromResources(args)
	if !ok {
		return fmt.Errorf("all resources must be specified before environment changes: %s", strings.Join(args, " "))
	}

	o.updatePodSpecForObject = polymorphichelpers.UpdatePodSpecForObjectFn
	o.output = cmdutil.GetFlagString(cmd, "output")
	var err error
	o.dryRunStrategy, err = cmdutil.GetDryRunStrategy(cmd)
	if err != nil {
		return err
	}

	cmdutil.PrintFlagsWithDryRunStrategy(o.PrintFlags, o.dryRunStrategy)
	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.PrintObj = printer.PrintObj

	o.clientset, err = f.KubernetesClientSet()
	if err != nil {
		return err
	}
	o.namespace, o.enforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	o.builder = f.NewBuilder

	return nil
}

// Validate makes sure provided values for EnvOptions are valid
func (o *EnvOptions) Validate() error {
	if o.Local && o.dryRunStrategy == cmdutil.DryRunServer {
		return fmt.Errorf("cannot specify --local and --dry-run=server - did you mean --dry-run=client?")
	}
	if len(o.Filenames) == 0 && len(o.resources) < 1 {
		return fmt.Errorf("one or more resources must be specified as <resource> <name> or <resource>/<name>")
	}
	if o.List && len(o.output) > 0 {
		return fmt.Errorf("--list and --output may not be specified together")
	}
	if len(o.Keys) > 0 && len(o.From) == 0 {
		return fmt.Errorf("when specifying --keys, a configmap or secret must be provided with --from")
	}
	return nil
}

// RunEnv contains all the necessary functionality for the OpenShift cli env command
func (o *EnvOptions) RunEnv() error {
	env, remove, _, err := envutil.ParseEnv(append(o.EnvParams, o.envArgs...), o.In)

	if err != nil {
		return err
	}

	if len(o.From) != 0 {
		b := o.builder().
			WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
			LocalParam(o.Local).
			ContinueOnError().
			NamespaceParam(o.namespace).DefaultNamespace().
			FilenameParam(o.enforceNamespace, &o.FilenameOptions).
			Flatten()
		if !o.Local {
			b = b.
				LabelSelectorParam(o.Selector).
				ResourceTypeOrNameArgs(o.All, o.From).
				Latest()
		}

		infos, err := b.Do().Infos()
		if err != nil {
			return err
		}

		for _, info := range infos {
			switch from := info.Object.(type) {
			case *v1.Secret:
				for key := range from.Data {
					if contains(key, o.Keys) {
						envVar := v1.EnvVar{
							Name: keyToEnvName(key),
							ValueFrom: &v1.EnvVarSource{
								SecretKeyRef: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: from.Name,
									},
									Key: key,
								},
							},
						}
						env = append(env, envVar)
					}
				}
			case *v1.ConfigMap:
				for key := range from.Data {
					if contains(key, o.Keys) {
						envVar := v1.EnvVar{
							Name: keyToEnvName(key),
							ValueFrom: &v1.EnvVarSource{
								ConfigMapKeyRef: &v1.ConfigMapKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: from.Name,
									},
									Key: key,
								},
							},
						}
						env = append(env, envVar)
					}
				}
			default:
				return fmt.Errorf("unsupported resource specified in --from")
			}
		}
	}

	if len(o.Prefix) != 0 {
		for i := range env {
			env[i].Name = fmt.Sprintf("%s%s", o.Prefix, env[i].Name)
		}
	}

	b := o.builder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		LocalParam(o.Local).
		ContinueOnError().
		NamespaceParam(o.namespace).DefaultNamespace().
		FilenameParam(o.enforceNamespace, &o.FilenameOptions).
		Flatten()

	if !o.Local {
		b.LabelSelectorParam(o.Selector).
			ResourceTypeOrNameArgs(o.All, o.resources...).
			Latest()
	}

	infos, err := b.Do().Infos()
	if err != nil {
		return err
	}

	if len(infos) == 0 {
		return nil
	}

	switch infos[0].Object.(type) {
	case *kruiseappsv1alpha1.CloneSet:

		obj, err := resource.
			NewHelper(infos[0].Client, infos[0].Mapping).
			Get(o.namespace, infos[0].Name)
		if err != nil {
			return err
		}
		res := obj.(*kruiseappsv1alpha1.CloneSet)

		resolutionErrorsEncountered := false
		containers, _ := selectContainers(res.Spec.Template.Spec.Containers, o.ContainerSelector)

		objName, err := meta.NewAccessor().Name(res)
		if err != nil {
			return err
		}

		gvks, _, err := scheme.Scheme.ObjectKinds(res)
		if err != nil {
			return err
		}

		objKind := res.GetObjectKind().GroupVersionKind().Kind
		if len(objKind) == 0 {
			for _, gvk := range gvks {
				if len(gvk.Kind) == 0 {
					continue
				}
				if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
					continue
				}

				objKind = gvk.Kind
				break
			}
		}

		if len(containers) == 0 {
			if gvks, _, err := scheme.Scheme.ObjectKinds(res); err == nil {
				objKind := res.GetObjectKind().GroupVersionKind().Kind
				if len(objKind) == 0 {
					for _, gvk := range gvks {
						if len(gvk.Kind) == 0 {
							continue
						}
						if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
							continue
						}

						objKind = gvk.Kind
						break
					}
				}

				fmt.Fprintf(o.ErrOut, "warning: %s/%s does not have any containers matching %q\n", objKind, objName, o.ContainerSelector)
			}
			return nil
		}

		for _, c := range containers {
			if !o.Overwrite {
				if err := validateNoOverwrites(c.Env, env); err != nil {
					return err
				}
			}

			c.Env = updateEnv(c.Env, env, remove)
			if o.List {
				resolveErrors := map[string][]string{}
				store := envutil.NewResourceStore()

				fmt.Fprintf(o.Out, "# %s %s, container %s\n", objKind, objName, c.Name)
				for _, env := range c.Env {
					// Print the simple value
					if env.ValueFrom == nil {
						fmt.Fprintf(o.Out, "%s=%s\n", env.Name, env.Value)
						continue
					}

					// Print the reference version
					if !o.Resolve {
						fmt.Fprintf(o.Out, "# %s from %s\n", env.Name, envutil.GetEnvVarRefString(env.ValueFrom))
						continue
					}

					value, err := envutil.GetEnvVarRefValue(o.clientset, o.namespace, store, env.ValueFrom, res, c)
					// Print the resolved value
					if err == nil {
						fmt.Fprintf(o.Out, "%s=%s\n", env.Name, value)
						continue
					}

					// Print the reference version and save the resolve error
					fmt.Fprintf(o.Out, "# %s from %s\n", env.Name, envutil.GetEnvVarRefString(env.ValueFrom))
					errString := err.Error()
					resolveErrors[errString] = append(resolveErrors[errString], env.Name)
					resolutionErrorsEncountered = true
				}

				// Print any resolution errors
				var errs []string
				for err, vars := range resolveErrors {
					sort.Strings(vars)
					errs = append(errs, fmt.Sprintf("error retrieving reference for %s: %v", strings.Join(vars, ", "), err))
				}
				sort.Strings(errs)
				for _, err := range errs {
					_, _ = fmt.Fprintln(o.ErrOut, err)
				}
			}
		}

		if !o.Local {
			_, err := resource.
				NewHelper(infos[0].Client, infos[0].Mapping).
				DryRun(o.dryRunStrategy == cmdutil.DryRunServer).
				Replace(infos[0].Namespace, infos[0].Name, true, res)
			if err != nil {
				return fmt.Errorf("failed to patch env update to pod template: %v", err)
			}
		}

		if resolutionErrorsEncountered {
			return errors.New("failed to retrieve valueFrom references")
		}

		if o.List {
			return nil
		}

		if err := o.PrintObj(res, o.Out); err != nil {
			return errors.New(err.Error())
		}

		return nil
	case *kruiseappsv1beta1.StatefulSet:
		obj, err := resource.
			NewHelper(infos[0].Client, infos[0].Mapping).
			DryRun(o.dryRunStrategy == cmdutil.DryRunServer).
			Get(o.namespace, infos[0].Name)
		if err != nil {
			return err
		}
		res := obj.(*kruiseappsv1beta1.StatefulSet)

		resolutionErrorsEncountered := false
		containers, _ := selectContainers(res.Spec.Template.Spec.Containers, o.ContainerSelector)

		objName, err := meta.NewAccessor().Name(res)
		if err != nil {
			return err
		}

		gvks, _, err := scheme.Scheme.ObjectKinds(res)
		if err != nil {
			return err
		}

		objKind := res.GetObjectKind().GroupVersionKind().Kind
		if len(objKind) == 0 {
			for _, gvk := range gvks {
				if len(gvk.Kind) == 0 {
					continue
				}
				if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
					continue
				}

				objKind = gvk.Kind
				break
			}
		}

		if len(containers) == 0 {
			if gvks, _, err := scheme.Scheme.ObjectKinds(res); err == nil {
				objKind := res.GetObjectKind().GroupVersionKind().Kind
				if len(objKind) == 0 {
					for _, gvk := range gvks {
						if len(gvk.Kind) == 0 {
							continue
						}
						if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
							continue
						}

						objKind = gvk.Kind
						break
					}
				}

				fmt.Fprintf(o.ErrOut, "warning: %s/%s does not have any containers matching %q\n", objKind, objName, o.ContainerSelector)
			}
			return nil
		}

		for _, c := range containers {
			if !o.Overwrite {
				if err := validateNoOverwrites(c.Env, env); err != nil {
					return err
				}
			}

			c.Env = updateEnv(c.Env, env, remove)
			if o.List {
				resolveErrors := map[string][]string{}
				store := envutil.NewResourceStore()

				fmt.Fprintf(o.Out, "# %s %s, container %s\n", objKind, objName, c.Name)
				for _, env := range c.Env {
					// Print the simple value
					if env.ValueFrom == nil {
						fmt.Fprintf(o.Out, "%s=%s\n", env.Name, env.Value)
						continue
					}

					// Print the reference version
					if !o.Resolve {
						fmt.Fprintf(o.Out, "# %s from %s\n", env.Name, envutil.GetEnvVarRefString(env.ValueFrom))
						continue
					}

					value, err := envutil.GetEnvVarRefValue(o.clientset, o.namespace, store, env.ValueFrom, res, c)
					// Print the resolved value
					if err == nil {
						fmt.Fprintf(o.Out, "%s=%s\n", env.Name, value)
						continue
					}

					// Print the reference version and save the resolve error
					fmt.Fprintf(o.Out, "# %s from %s\n", env.Name, envutil.GetEnvVarRefString(env.ValueFrom))
					errString := err.Error()
					resolveErrors[errString] = append(resolveErrors[errString], env.Name)
					resolutionErrorsEncountered = true
				}

				// Print any resolution errors
				var errs []string
				for err, vars := range resolveErrors {
					sort.Strings(vars)
					errs = append(errs, fmt.Sprintf("error retrieving reference for %s: %v", strings.Join(vars, ", "), err))
				}
				sort.Strings(errs)
				for _, err := range errs {
					_, _ = fmt.Fprintln(o.ErrOut, err)
				}
			}
		}

		if !o.Local {
			_, err := resource.
				NewHelper(infos[0].Client, infos[0].Mapping).
				DryRun(o.dryRunStrategy == cmdutil.DryRunServer).
				Replace(infos[0].Namespace, infos[0].Name, true, res)
			if err != nil {
				return fmt.Errorf("failed to patch env update to pod template: %v", err)
			}
		}

		if resolutionErrorsEncountered {
			return errors.New("failed to retrieve valueFrom references")
		}

		if o.List {
			return nil
		}

		if err := o.PrintObj(res, o.Out); err != nil {
			return errors.New(err.Error())
		}

		return nil
	default:
		patches := CalculatePatches(infos, scheme.DefaultJSONEncoder(), func(obj runtime.Object) ([]byte, error) {
			_, err := o.updatePodSpecForObject(obj, func(spec *v1.PodSpec) error {
				resolutionErrorsEncountered := false
				containers, _ := selectContainers(spec.Containers, o.ContainerSelector)
				objName, err := meta.NewAccessor().Name(obj)
				if err != nil {
					return err
				}

				gvks, _, err := scheme.Scheme.ObjectKinds(obj)
				if err != nil {
					return err
				}
				objKind := obj.GetObjectKind().GroupVersionKind().Kind
				if len(objKind) == 0 {
					for _, gvk := range gvks {
						if len(gvk.Kind) == 0 {
							continue
						}
						if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
							continue
						}

						objKind = gvk.Kind
						break
					}
				}

				if len(containers) == 0 {
					if gvks, _, err := scheme.Scheme.ObjectKinds(obj); err == nil {
						objKind := obj.GetObjectKind().GroupVersionKind().Kind
						if len(objKind) == 0 {
							for _, gvk := range gvks {
								if len(gvk.Kind) == 0 {
									continue
								}
								if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
									continue
								}

								objKind = gvk.Kind
								break
							}
						}

						fmt.Fprintf(o.ErrOut, "warning: %s/%s does not have any containers matching %q\n", objKind, objName, o.ContainerSelector)
					}
					return nil
				}
				for _, c := range containers {
					if !o.Overwrite {
						if err := validateNoOverwrites(c.Env, env); err != nil {
							return err
						}
					}

					c.Env = updateEnv(c.Env, env, remove)
					if o.List {
						resolveErrors := map[string][]string{}
						store := envutil.NewResourceStore()

						fmt.Fprintf(o.Out, "# %s %s, container %s\n", objKind, objName, c.Name)
						for _, env := range c.Env {
							// Print the simple value
							if env.ValueFrom == nil {
								fmt.Fprintf(o.Out, "%s=%s\n", env.Name, env.Value)
								continue
							}

							// Print the reference version
							if !o.Resolve {
								fmt.Fprintf(o.Out, "# %s from %s\n", env.Name, envutil.GetEnvVarRefString(env.ValueFrom))
								continue
							}

							value, err := envutil.GetEnvVarRefValue(o.clientset, o.namespace, store, env.ValueFrom, obj, c)
							// Print the resolved value
							if err == nil {
								fmt.Fprintf(o.Out, "%s=%s\n", env.Name, value)
								continue
							}

							// Print the reference version and save the resolve error
							fmt.Fprintf(o.Out, "# %s from %s\n", env.Name, envutil.GetEnvVarRefString(env.ValueFrom))
							errString := err.Error()
							resolveErrors[errString] = append(resolveErrors[errString], env.Name)
							resolutionErrorsEncountered = true
						}

						// Print any resolution errors
						var errs []string
						for err, vars := range resolveErrors {
							sort.Strings(vars)
							errs = append(errs, fmt.Sprintf("error retrieving reference for %s: %v", strings.Join(vars, ", "), err))
						}
						sort.Strings(errs)
						for _, err := range errs {
							_, _ = fmt.Fprintln(o.ErrOut, err)
						}
					}
				}
				if resolutionErrorsEncountered {
					return errors.New("failed to retrieve valueFrom references")
				}
				return nil
			})

			if err == nil {
				return runtime.Encode(scheme.DefaultJSONEncoder(), obj)
			}
			return nil, err
		})

		if o.List {
			return nil
		}

		var allErrs []error

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

			if o.Local || o.dryRunStrategy == cmdutil.DryRunClient {
				if err := o.PrintObj(info.Object, o.Out); err != nil {
					allErrs = append(allErrs, err)
				}
				continue
			}

			actual, err := resource.
				NewHelper(info.Client, info.Mapping).
				DryRun(o.dryRunStrategy == cmdutil.DryRunServer).
				Patch(info.Namespace, info.Name, types.StrategicMergePatchType, patch.Patch, nil)
			if err != nil {
				allErrs = append(allErrs, fmt.Errorf("failed to patch env update to pod template: %v", err))
				continue
			}

			// make sure arguments to set or replace environment variables are set
			// before returning a successful message
			if len(env) == 0 && len(o.envArgs) == 0 {
				return fmt.Errorf("at least one environment variable must be provided")
			}

			if err := o.PrintObj(actual, o.Out); err != nil {
				allErrs = append(allErrs, err)
			}
		}
		return utilerrors.NewAggregate(allErrs)
	}
}
