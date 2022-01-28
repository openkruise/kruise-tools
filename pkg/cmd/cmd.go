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

package cmd

import (
	"flag"
	"io"
	"os"

	cmdexec "github.com/openkruise/kruise-tools/pkg/cmd/exec"
	"github.com/openkruise/kruise-tools/pkg/cmd/expose"
	"github.com/openkruise/kruise-tools/pkg/cmd/migrate"
	krollout "github.com/openkruise/kruise-tools/pkg/cmd/rollout"
	"github.com/openkruise/kruise-tools/pkg/cmd/scaledown"
	kset "github.com/openkruise/kruise-tools/pkg/cmd/set"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/kubectl/pkg/cmd/apiresources"
	"k8s.io/kubectl/pkg/cmd/apply"
	cmdconfig "k8s.io/kubectl/pkg/cmd/config"
	"k8s.io/kubectl/pkg/cmd/diff"
	"k8s.io/kubectl/pkg/cmd/kustomize"
	"k8s.io/kubectl/pkg/cmd/options"
	"k8s.io/kubectl/pkg/cmd/patch"
	"k8s.io/kubectl/pkg/cmd/plugin"
	"k8s.io/kubectl/pkg/cmd/replace"
	"k8s.io/kubectl/pkg/cmd/scale"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/cmd/version"
	"k8s.io/kubectl/pkg/cmd/wait"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

const (
	bashCompletionFunc = `# call kubectl get $1,
__kubectl_debug_out()
{
    local cmd="$1"
    __kubectl_debug "${FUNCNAME[1]}: get completion by ${cmd}"
    eval "${cmd} 2>/dev/null"
}

__kubectl_override_flag_list=(--kubeconfig --cluster --user --context --namespace --server -n -s)
__kubectl_override_flags()
{
    local ${__kubectl_override_flag_list[*]##*-} two_word_of of var
    for w in "${words[@]}"; do
        if [ -n "${two_word_of}" ]; then
            eval "${two_word_of##*-}=\"${two_word_of}=\${w}\""
            two_word_of=
            continue
        fi
        for of in "${__kubectl_override_flag_list[@]}"; do
            case "${w}" in
                ${of}=*)
                    eval "${of##*-}=\"${w}\""
                    ;;
                ${of})
                    two_word_of="${of}"
                    ;;
            esac
        done
    done
    for var in "${__kubectl_override_flag_list[@]##*-}"; do
        if eval "test -n \"\$${var}\""; then
            eval "echo -n \${${var}}' '"
        fi
    done
}

__kubectl_config_get_contexts()
{
    __kubectl_parse_config "contexts"
}

__kubectl_config_get_clusters()
{
    __kubectl_parse_config "clusters"
}

__kubectl_config_get_users()
{
    __kubectl_parse_config "users"
}

# $1 has to be "contexts", "clusters" or "users"
__kubectl_parse_config()
{
    local template kubectl_out
    template="{{ range .$1  }}{{ .name }} {{ end }}"
    if kubectl_out=$(__kubectl_debug_out "kubectl config $(__kubectl_override_flags) -o template --template=\"${template}\" view"); then
        COMPREPLY=( $( compgen -W "${kubectl_out[*]}" -- "$cur" ) )
    fi
}

# $1 is the name of resource (required)
# $2 is template string for kubectl get (optional)
__kubectl_parse_get()
{
    local template
    template="${2:-"{{ range .items  }}{{ .metadata.name }} {{ end }}"}"
    local kubectl_out
    if kubectl_out=$(__kubectl_debug_out "kubectl get $(__kubectl_override_flags) -o template --template=\"${template}\" \"$1\""); then
        COMPREPLY+=( $( compgen -W "${kubectl_out[*]}" -- "$cur" ) )
    fi
}

# Same as __kubectl_get_resources (with s) but allows completion for only one resource name.
__kubectl_get_resource()
{
    if [[ ${#nouns[@]} -eq 0 ]]; then
      __kubectl_get_resource_helper "" "$cur"
      return # the return status is that of the last command executed in the function body
    fi
    __kubectl_parse_get "${nouns[${#nouns[@]} -1]}"
}

# Same as __kubectl_get_resource (without s) but allows completion for multiple, comma-separated resource names.
__kubectl_get_resources()
{
    local SEPARATOR=','
    if [[ ${#nouns[@]} -eq 0 ]]; then
      local kubectl_out HEAD TAIL
      HEAD=""
      TAIL="$cur"
      # if SEPARATOR is contained in $cur, e.g. "pod,sec"
      if [[ "$cur" = *${SEPARATOR}* ]] ; then
        # set HEAD to "pod,"
        HEAD="${cur%${SEPARATOR}*}${SEPARATOR}"
        # set TAIL to "sec"
        TAIL="${cur##*${SEPARATOR}}"
      fi
      __kubectl_get_resource_helper "$HEAD" "$TAIL"
      return # the return status is that of the last command executed in the function body
    fi
    __kubectl_parse_get "${nouns[${#nouns[@]} -1]}"
}

__kubectl_get_resource_helper()
{
    local kubectl_out HEAD TAIL
    HEAD="$1"
    TAIL="$2"
    if kubectl_out=$(__kubectl_debug_out "kubectl api-resources $(__kubectl_override_flags) -o name --cached --request-timeout=5s --verbs=get"); then
        COMPREPLY=( $( compgen -P "$HEAD" -W "${kubectl_out[*]}" -- "$TAIL" ) )
        return 0
    fi
    return 1
}

__kubectl_get_resource_namespace()
{
    __kubectl_parse_get "namespace"
}

__kubectl_get_resource_pod()
{
    __kubectl_parse_get "pod"
}

__kubectl_get_resource_rc()
{
    __kubectl_parse_get "rc"
}

__kubectl_get_resource_node()
{
    __kubectl_parse_get "node"
}

__kubectl_get_resource_clusterrole()
{
    __kubectl_parse_get "clusterrole"
}

# $1 is the name of the pod we want to get the list of containers inside
__kubectl_get_containers()
{
    local template
    template="{{ range .spec.initContainers }}{{ .name }} {{end}}{{ range .spec.containers  }}{{ .name }} {{ end }}"
    __kubectl_debug "${FUNCNAME} nouns are ${nouns[*]}"

    local len="${#nouns[@]}"
    if [[ ${len} -ne 1 ]]; then
        return
    fi
    local last=${nouns[${len} -1]}
    local kubectl_out
    if kubectl_out=$(__kubectl_debug_out "kubectl get $(__kubectl_override_flags) -o template --template=\"${template}\" pods \"${last}\""); then
        COMPREPLY=( $( compgen -W "${kubectl_out[*]}" -- "$cur" ) )
    fi
}

# Require both a pod and a container to be specified
__kubectl_require_pod_and_container()
{
    if [[ ${#nouns[@]} -eq 0 ]]; then
        __kubectl_parse_get pods
        return 0
    fi;
    __kubectl_get_containers
    return 0
}

__kubectl_cp()
{
    if [[ $(type -t compopt) = "builtin" ]]; then
        compopt -o nospace
    fi

    case "$cur" in
        /*|[.~]*) # looks like a path
            return
            ;;
        *:*) # TODO: complete remote files in the pod
            return
            ;;
        */*) # complete <namespace>/<pod>
            local template namespace kubectl_out
            template="{{ range .items }}{{ .metadata.namespace }}/{{ .metadata.name }}: {{ end }}"
            namespace="${cur%%/*}"
            if kubectl_out=$(__kubectl_debug_out "kubectl get $(__kubectl_override_flags) --namespace \"${namespace}\" -o template --template=\"${template}\" pods"); then
                COMPREPLY=( $(compgen -W "${kubectl_out[*]}" -- "${cur}") )
            fi
            return
            ;;
        *) # complete namespaces, pods, and filedirs
            __kubectl_parse_get "namespace" "{{ range .items  }}{{ .metadata.name }}/ {{ end }}"
            __kubectl_parse_get "pod" "{{ range .items  }}{{ .metadata.name }}: {{ end }}"
            _filedir
            ;;
    esac
}

__kubectl_custom_func() {
    case ${last_command} in
        kubectl_get)
            __kubectl_get_resources
            return
            ;;
        kubectl_describe | kubectl_delete | kubectl_label | kubectl_edit | kubectl_patch |\
        kubectl_annotate | kubectl_expose | kubectl_scale | kubectl_autoscale | kubectl_taint | kubectl_rollout_* |\
        kubectl_apply_edit-last-applied | kubectl_apply_view-last-applied)
            __kubectl_get_resource
            return
            ;;
        kubectl_logs)
            __kubectl_require_pod_and_container
            return
            ;;
        kubectl_exec | kubectl_port-forward | kubectl_top_pod | kubectl_attach)
            __kubectl_get_resource_pod
            return
            ;;
        kubectl_cordon | kubectl_uncordon | kubectl_drain | kubectl_top_node)
            __kubectl_get_resource_node
            return
            ;;
        kubectl_config_use-context | kubectl_config_rename-context | kubectl_config_delete-context)
            __kubectl_config_get_contexts
            return
            ;;
        kubectl_config_delete-cluster)
            __kubectl_config_get_clusters
            return
            ;;
        kubectl_cp)
            __kubectl_cp
            return
            ;;
        *)
            ;;
    esac
}
`
)

const kubectlCmdHeaders = "KUBECTL_COMMAND_HEADERS"

var (
	bashCompletionFlags = map[string]string{
		"namespace": "__kubectl_get_resource_namespace",
		"context":   "__kubectl_config_get_contexts",
		"cluster":   "__kubectl_config_get_clusters",
		"user":      "__kubectl_config_get_users",
	}
)

// NewKubectlCommand creates the `kubectl-kruise` command and its nested children.
func NewKubectlCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	warningsAsErrors := false
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "kubectl-kruise",
		Short: i18n.T("kubectl-kruise controls the OpenKruise CRs"),
		Long: templates.LongDesc(`
      kubectl-kruise controls the OpenKruise manager.

      Find more information at:
            https://openkruise.io/`),
		Run: runHelp,
		// Hook before and after Run initialize and write profiles to disk,
		// respectively.
		PersistentPreRunE: func(*cobra.Command, []string) error {
			//rest.SetDefaultWarningHandler(warningHandler)
			return initProfiling()
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			if err := flushProfiling(); err != nil {
				return err
			}
			return nil
		},
		BashCompletionFunction: bashCompletionFunc,
	}

	flags := cmds.PersistentFlags()
	flags.SetNormalizeFunc(cliflag.WarnWordSepNormalizeFunc) // Warn for "_" flags

	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	addProfilingFlags(flags)

	flags.BoolVar(&warningsAsErrors, "warnings-as-errors", warningsAsErrors, "Treat warnings received from the server as errors and exit with a non-zero exit code")

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	kubeConfigFlags.AddFlags(flags)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(cmds.PersistentFlags())

	cmds.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Sending in 'nil' for the getLanguageFn() results in using
	// the LANG environment variable.
	//
	// TODO: Consider adding a flag or file preference for setting
	// the language, instead of just loading from the LANG env. variable.
	i18n.LoadTranslations("kubectl", nil)

	// From this point and forward we get warnings on flags that contain "_" separators
	cmds.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: err}

	groups := templates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				expose.NewCmdExposeService(f, ioStreams),
				cmdWithShortOverwrite(scale.NewCmdScale(f, ioStreams), "Set a new size for a Deployment, ReplicaSet, CloneSet, or Advanced StatefulSet"),
			},
		},
		{
			Message: "Troubleshooting and Debugging Commands:",
			Commands: []*cobra.Command{
				cmdexec.NewCmdExec(f, ioStreams),
			},
		},

		{
			Message: "CloneSet Commands:",
			Commands: []*cobra.Command{
				krollout.NewCmdRollout(f, ioStreams),
				kset.NewCmdSet(f, ioStreams),
				migrate.NewCmdMigrate(f, ioStreams),
			},
		},
		{
			Message: "AdvancedStatefulSet Commands:",
			Commands: []*cobra.Command{
				krollout.NewCmdRollout(f, ioStreams),
				kset.NewCmdSet(f, ioStreams),
			},
		},
		{
			Message: "Scaledown Commands",
			Commands: []*cobra.Command{
				scaledown.NewCmdScaleDown(f, ioStreams),
			},
		},
		{
			Message: "Advanced Commands:",
			Commands: []*cobra.Command{
				diff.NewCmdDiff(f, ioStreams),
				apply.NewCmdApply("kubectl-kruise", f, ioStreams),
				patch.NewCmdPatch(f, ioStreams),
				replace.NewCmdReplace(f, ioStreams),
				wait.NewCmdWait(f, ioStreams),
				kustomize.NewCmdKustomize(ioStreams),
			},
		},
	}
	groups.Add(cmds)

	filters := []string{"options"}

	// Hide the "alpha" subcommand if there are no alpha commands in this build.
	alpha := NewCmdAlpha(f, ioStreams)
	if !alpha.HasSubCommands() {
		filters = append(filters, alpha.Name())
	}

	templates.ActsAsRootCommand(cmds, filters, groups...)

	for name, completion := range bashCompletionFlags {
		if cmds.Flag(name) != nil {
			if cmds.Flag(name).Annotations == nil {
				cmds.Flag(name).Annotations = map[string][]string{}
			}
			cmds.Flag(name).Annotations[cobra.BashCompCustom] = append(
				cmds.Flag(name).Annotations[cobra.BashCompCustom],
				completion,
			)
		}
	}

	cmds.AddCommand(alpha)
	cmds.AddCommand(cmdconfig.NewCmdConfig(f, clientcmd.NewDefaultPathOptions(), ioStreams))
	cmds.AddCommand(plugin.NewCmdPlugin(f, ioStreams))
	cmds.AddCommand(version.NewCmdVersion(f, ioStreams))
	cmds.AddCommand(apiresources.NewCmdAPIVersions(f, ioStreams))
	cmds.AddCommand(apiresources.NewCmdAPIResources(f, ioStreams))
	cmds.AddCommand(options.NewCmdOptions(ioStreams.Out))

	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

// NewDefaultKubectlCommand creates the `kubectl` command with default arguments
func NewDefaultKubectlCommand() *cobra.Command {
	return NewDefaultKubectlCommandWithArgs(os.Args, os.Stdin, os.Stdout, os.Stderr)
}

// NewDefaultKubectlCommandWithArgs creates the `kubectl` command with arguments
func NewDefaultKubectlCommandWithArgs(args []string, in io.Reader, out, errout io.Writer) *cobra.Command {
	cmd := NewKubectlCommand(in, out, errout)

	return cmd
}

func cmdWithShortOverwrite(cmd *cobra.Command, short string) *cobra.Command {
	cmd.Short = i18n.T(short)
	return cmd
}
