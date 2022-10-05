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

package generator

import (
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type ResourceDistributionPlugin struct {
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	ResourceArgs     `json:"resource,omitempty" yaml:"resource,omitempty"`
	TargetsArgs      `json:"targets,omitempty" yaml:"targets,omitempty"`

	// Options for the resourcedistribution.
	// GeneratorOptions same as configmap and secret generator Options
	Options *types.GeneratorOptions `json:"options,omitempty" yaml:"options,omitempty"`

	// Behavior of generated resource, must be one of:
	//   'create': create a new one
	//   'replace': replace the existing one
	//   'merge': merge with the existing one
	Behavior string `json:"behavior,omitempty" yaml:"behavior,omitempty"`
}

// ResourceArgs contain arguments for the resource to be distributed.
type ResourceArgs struct {
	// Name of the resource to be distributed.
	ResourceName string `json:"resourceName,omitempty" yaml:"resourceName,omitempty"`
	// Only configmap and secret are available
	ResourceKind string `json:"resourceKind,omitempty" yaml:"resourceKind,omitempty"`

	// KvPairSources defines places to obtain key value pairs.
	// same as configmap and secret generator KvPairSources
	types.KvPairSources `json:",inline,omitempty" yaml:",inline,omitempty"`

	// Options for the resource to be distributed.
	// GeneratorOptions same as configmap and secret generator Options
	ResourceOptions *types.GeneratorOptions `json:"resourceOptions,omitempty" yaml:"resourceOptions,omitempty"`

	// Type of the secret. It can be "Opaque" (default), or "kubernetes.io/tls".
	//
	// If type is "kubernetes.io/tls", then "literals" or "files" must have exactly two
	// keys: "tls.key" and "tls.crt"
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// TargetsArgs defines places to obtain target namespace args.
type TargetsArgs struct {
	// AllNamespaces if true distribute all namespaces
	AllNamespaces bool `json:"allNamespaces,omitempty" yaml:"allNamespaces,omitempty"`

	// ExcludedNamespaces is a list of excluded namespaces name.
	ExcludedNamespaces []string `json:"excludedNamespaces,omitempty" yaml:"excludedNamespaces,omitempty"`

	// IncludedNamespaces is a list of included namespaces name.
	IncludedNamespaces []string `json:"includedNamespaces,omitempty" yaml:"includedNamespaces,omitempty"`

	// NamespaceLabelSelector for the generator.
	NamespaceLabelSelector *metav1.LabelSelector `json:"namespaceLabelSelector,omitempty" yaml:"namespaceLabelSelector,omitempty"`
}

func BuildCmd() *cobra.Command {
	config := new(ResourceDistributionPlugin)
	fn := func(items []*yaml.RNode) ([]*yaml.RNode, error) {
		rn, err := MakeResourceDistribution(config)
		if err != nil {
			return nil, err
		}

		var itemsOutput []*yaml.RNode
		itemsOutput = append(itemsOutput, rn)
		return itemsOutput, nil
	}

	p := framework.SimpleProcessor{Config: config, Filter: kio.FilterFunc(fn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	return cmd
}

// MakeResourceDistribution makes a ResourceDistribution.
//
// ResourceDistribution: https://openkruise.io/docs/user-manuals/resourcedistribution
func MakeResourceDistribution(config *ResourceDistributionPlugin) (*yaml.RNode, error) {
	rn, err := makeBaseNode(&config.ObjectMeta)
	if err != nil {
		return nil, err
	}

	// set Labels and Annotations for ResourceDistribution
	if config.Options != nil {
		err = setLabelsOrAnnotations(rn, config.Options.Annotations, metadataAnnotationsPath)
		if err != nil {
			return nil, err
		}
		err = setLabelsOrAnnotations(rn, config.Options.Labels, metadataLabelsPath)
		if err != nil {
			return nil, err
		}
	}

	if config.ObjectMeta.Name == "" {
		return nil, errors.Errorf("a ResourceDistribution must have a name ")
	}
	err = rn.PipeE(yaml.SetK8sName(config.ObjectMeta.Name))
	if err != nil {
		return nil, err
	}

	err = setResource(rn, &config.ResourceArgs)
	if err != nil {
		return nil, err
	}

	err = setTargets(rn, &config.TargetsArgs)
	if err != nil {
		return nil, err
	}

	return rn, nil
}
