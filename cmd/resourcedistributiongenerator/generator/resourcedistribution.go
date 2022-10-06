/*
Copyright 2022 The Kruise Authors.
Copyright 2015 The Kubernetes Authors.

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
	//
	// The features of fields DisableNameSuffixHash and Immutable in Options are not implemented yet
	Options *types.GeneratorOptions `json:"options,omitempty" yaml:"options,omitempty"`

	// Behavior of generated resource, must be one of:
	//   'create': create a new one
	//   'replace': replace the existing one
	//   'merge': merge with the existing one
	//
	//   The feature of this field is not implemented yet
	//Behavior string `json:"behavior,omitempty" yaml:"behavior,omitempty"`
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
	//
	// The feature of field DisableNameSuffixHash in ResourceOptions is not implemented yet
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
	NamespaceLabelSelector *LabelSelector `json:"namespaceLabelSelector,omitempty" yaml:"namespaceLabelSelector,omitempty"`
}

// It is the same as metav1.LabelSelector except that the YAMl tag is added after each field
//
// A label selector is a label query over a set of resources. The result of matchLabels and
// matchExpressions are ANDed. An empty label selector matches all objects. A null
// label selector matches no objects.
// +structType=atomic
type LabelSelector struct {
	// matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
	// map is equivalent to an element of matchExpressions, whose key field is "key", the
	// operator is "In", and the values array contains only "value". The requirements are ANDed.
	// +optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"  protobuf:"bytes,1,rep,name=matchLabels" yaml:"matchLabels,omitempty"`
	// matchExpressions is a list of label selector requirements. The requirements are ANDed.
	// +optional
	MatchExpressions []LabelSelectorRequirement `json:"matchExpressions,omitempty" protobuf:"bytes,2,rep,name=matchExpressions" yaml:"matchExpressions,omitempty"`
}

// A label selector requirement is a selector that contains values, a key, and an operator that
// relates the key and values.
type LabelSelectorRequirement struct {
	// key is the label key that the selector applies to.
	// +patchMergeKey=key
	// +patchStrategy=merge
	Key string `json:"key" patchStrategy:"merge" patchMergeKey:"key" protobuf:"bytes,1,opt,name=key" yaml:"key,omitempty"`
	// operator represents a key's relationship to a set of values.
	// Valid operators are In, NotIn, Exists and DoesNotExist.
	Operator metav1.LabelSelectorOperator `json:"operator"  protobuf:"bytes,2,opt,name=operator,casttype=LabelSelectorOperator" yaml:"operator,omitempty"`
	// values is an array of string values. If the operator is In or NotIn,
	// the values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty. This array is replaced during a strategic
	// merge patch.
	// +optional
	Values []string `json:"values,omitempty" protobuf:"bytes,3,rep,name=values"  yaml:"values,omitempty"`
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
