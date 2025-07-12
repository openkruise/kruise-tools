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

package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

func TestSetTargets(t *testing.T) {
	t.Run("should correctly set all target fields when combined", func(t *testing.T) {
		rn := yaml.NewRNode(&yaml.Node{Kind: yaml.MappingNode})
		args := &TargetsArgs{
			AllNamespaces:      true,
			IncludedNamespaces: []string{"include-this"},
			ExcludedNamespaces: []string{"exclude-this"},
			NamespaceLabelSelector: &LabelSelector{
				MatchLabels:      map[string]string{"env": "prod"},
				MatchExpressions: []LabelSelectorRequirement{{Key: "tier", Operator: metav1.LabelSelectorOpIn, Values: []string{"frontend"}}},
			},
		}
		err := setTargets(rn, args)
		require.NoError(t, err)

		type resultStruct struct {
			Spec struct {
				Targets struct {
					AllNamespaces      bool `yaml:"allNamespaces"`
					IncludedNamespaces struct {
						List []string `yaml:"list"`
					} `yaml:"includedNamespaces"`
					ExcludedNamespaces struct {
						List []string `yaml:"list"`
					} `yaml:"excludedNamespaces"`
					NamespaceLabelSelector struct {
						MatchLabels      map[string]string          `yaml:"matchLabels"`
						MatchExpressions []LabelSelectorRequirement `yaml:"matchExpressions"`
					} `yaml:"namespaceLabelSelector"`
				} `yaml:"targets"`
			} `yaml:"spec"`
		}

		var result resultStruct
		err = k8syaml.Unmarshal([]byte(rn.MustString()), &result)
		require.NoError(t, err)

		assert.True(t, result.Spec.Targets.AllNamespaces)
		assert.Equal(t, []string{"include-this"}, result.Spec.Targets.IncludedNamespaces.List)
		assert.Equal(t, []string{"exclude-this"}, result.Spec.Targets.ExcludedNamespaces.List)
		assert.Equal(t, map[string]string{"env": "prod"}, result.Spec.Targets.NamespaceLabelSelector.MatchLabels)
		require.Len(t, result.Spec.Targets.NamespaceLabelSelector.MatchExpressions, 1)
		assert.Equal(t, "tier", result.Spec.Targets.NamespaceLabelSelector.MatchExpressions[0].Key)
	})

	t.Run("should return error if all target args are empty", func(t *testing.T) {
		rn := yaml.NewRNode(&yaml.Node{Kind: yaml.MappingNode})
		args := &TargetsArgs{}
		err := setTargets(rn, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "The targets field of the ResourceDistribution cannot be empty")
	})
}

func TestSetIncludedExcludedNs(t *testing.T) {
	t.Run("should correctly set included namespaces", func(t *testing.T) {
		rn := yaml.NewRNode(&yaml.Node{Kind: yaml.MappingNode})
		err := setIncludedExcludedNs(rn, []string{"ns1", "ns2"}, includedNamespacesPath)
		require.NoError(t, err)
		expectedYAML := `
spec:
  targets:
    includedNamespaces:
      list:
      - ns1
      - ns2
`
		assert.Equal(t, strings.TrimSpace(expectedYAML), strings.TrimSpace(rn.MustString()))
	})

	t.Run("should correctly set excluded namespaces", func(t *testing.T) {
		rn := yaml.NewRNode(&yaml.Node{Kind: yaml.MappingNode})
		err := setIncludedExcludedNs(rn, []string{"kube-system"}, excludedNamespacesPath)
		require.NoError(t, err)
		expectedYAML := `
spec:
  targets:
    excludedNamespaces:
      list:
      - kube-system
`
		assert.Equal(t, strings.TrimSpace(expectedYAML), strings.TrimSpace(rn.MustString()))
	})
}

func TestSetNsLabelSelector(t *testing.T) {
	t.Run("should correctly set both matchLabels and matchExpressions", func(t *testing.T) {
		rn := yaml.NewRNode(&yaml.Node{Kind: yaml.MappingNode})
		selector := &LabelSelector{
			MatchLabels:      map[string]string{"env": "prod"},
			MatchExpressions: []LabelSelectorRequirement{{Key: "tier", Operator: metav1.LabelSelectorOpIn, Values: []string{"frontend"}}},
		}
		err := setNsLabelSelector(rn, selector)
		require.NoError(t, err)
		expectedYAML := `
spec:
  targets:
    namespaceLabelSelector:
      matchExpressions:
      - key: tier
        operator: In
        values:
        - frontend
      matchLabels:
        env: prod
`
		assert.Equal(t, strings.TrimSpace(expectedYAML), strings.TrimSpace(rn.MustString()))
	})
}

func TestSetMatchExpressions(t *testing.T) {
	testcases := []struct {
		name          string
		expressions   []LabelSelectorRequirement
		expectedError string
	}{
		{name: "should succeed with valid 'In' operator", expressions: []LabelSelectorRequirement{{Key: "env", Operator: metav1.LabelSelectorOpIn, Values: []string{"prod"}}}},
		{name: "should fail with invalid operator", expressions: []LabelSelectorRequirement{{Key: "env", Operator: "Equals", Values: []string{"prod"}}}, expectedError: "operator is invalid"},
		{name: "should fail with empty key", expressions: []LabelSelectorRequirement{{Key: "", Operator: metav1.LabelSelectorOpIn, Values: []string{"prod"}}}, expectedError: "key cannot be empty"},
		{name: "should fail with nil values for 'In' operator", expressions: []LabelSelectorRequirement{{Key: "env", Operator: metav1.LabelSelectorOpIn, Values: nil}}, expectedError: "values for In and NotIn cannot be empty"},
		{name: "should fail with non-nil values for 'Exists' operator", expressions: []LabelSelectorRequirement{{Key: "env", Operator: metav1.LabelSelectorOpExists, Values: []string{"prod"}}}, expectedError: "values for Exist and DoesNotExist must be empty"},
		{name: "should fail with duplicate values", expressions: []LabelSelectorRequirement{{Key: "env", Operator: metav1.LabelSelectorOpIn, Values: []string{"prod", "prod"}}}, expectedError: "values cannot be repeated"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rn := yaml.NewRNode(&yaml.Node{Kind: yaml.MappingNode})
			err := setMatchExpressions(rn, tc.expressions)
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewNameListRNode(t *testing.T) {
	t.Run("should produce a simple list of strings", func(t *testing.T) {
		input := []string{"ns1", "kube-system"}
		rnode := newNameListRNode(input...)
		require.NotNil(t, rnode)
		expected := `- ns1
- kube-system`
		assert.Equal(t, expected, strings.TrimSpace(rnode.MustString()))
	})

	t.Run("should produce an empty sequence for an empty list", func(t *testing.T) {
		rnode := newNameListRNode([]string{}...)
		require.NotNil(t, rnode)
		assert.Equal(t, "[]", strings.TrimSpace(rnode.MustString()))
	})
}

func TestSetAllNs(t *testing.T) {
	t.Run("should set allNamespaces to true", func(t *testing.T) {
		rn := yaml.NewRNode(&yaml.Node{Kind: yaml.MappingNode})
		err := setAllNs(rn, true)
		require.NoError(t, err)

		var obj unstructured.Unstructured
		err = k8syaml.Unmarshal([]byte(rn.MustString()), &obj.Object)
		require.NoError(t, err)

		lookupPath := append(targetsPath, allNamespacesField)
		allNs, found, err := unstructured.NestedBool(obj.Object, lookupPath...)
		require.NoError(t, err)
		assert.True(t, found)
		assert.True(t, allNs)
	})

	t.Run("should do nothing if allNamespaces is false", func(t *testing.T) {
		rn := yaml.NewRNode(&yaml.Node{Kind: yaml.MappingNode})
		err := setAllNs(rn, false)
		require.NoError(t, err)
		assert.Equal(t, "{}", strings.TrimSpace(rn.MustString()))
	})
}
