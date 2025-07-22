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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

func TestMakeResourceDistribution(t *testing.T) {
	testcases := []struct {
		name          string
		config        *ResourceDistributionPlugin
		expectedYAML  string
		expectedError string
	}{
		{
			name: "Test successful generation with full config",
			config: &ResourceDistributionPlugin{
				ObjectMeta: types.ObjectMeta{Name: "my-rd-resource"},
				Options:    &types.GeneratorOptions{Labels: map[string]string{"app": "nginx"}, Annotations: map[string]string{"author": "kruise"}},
				ResourceArgs: ResourceArgs{
					ResourceName: "my-secret",
					ResourceKind: "Secret",
					Type:         "Opaque",
				},
				TargetsArgs: TargetsArgs{
					IncludedNamespaces: []string{"ns1", "ns2"},
					ExcludedNamespaces: []string{"kube-system"},
					NamespaceLabelSelector: &LabelSelector{
						MatchLabels:      map[string]string{"foo": "bar"},
						MatchExpressions: []LabelSelectorRequirement{{Key: "env", Operator: metav1.LabelSelectorOpIn, Values: []string{"prod", "staging"}}},
					},
				},
			},
			expectedYAML: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
metadata:
  name: my-rd-resource
  labels:
    app: nginx
  annotations:
    author: kruise
spec:
  resource:
    apiVersion: v1
    kind: Secret
    metadata:
      name: my-secret
    data: {}
    type: Opaque
  targets:
    includedNamespaces:
      list:
      - name: ns1
      - name: ns2
    excludedNamespaces:
      list:
      - name: kube-system
    namespaceLabelSelector:
      matchLabels:
        foo: bar
      matchExpressions:
      - key: env
        operator: In
        values:
        - prod
        - staging
`,
		},
		{
			name: "Test resource with literal data sources",
			config: &ResourceDistributionPlugin{
				ObjectMeta: types.ObjectMeta{Name: "rd-with-literals"},
				ResourceArgs: ResourceArgs{
					ResourceName: "my-cm",
					ResourceKind: "ConfigMap",
					KvPairSources: types.KvPairSources{
						LiteralSources: []string{"key1=value1", "key2=value2"},
					},
				},
				TargetsArgs: TargetsArgs{
					AllNamespaces: true,
				},
			},
			expectedYAML: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
metadata:
  name: rd-with-literals
spec:
  resource:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: my-cm
    data:
      key1: value1
      key2: value2
  targets:
    allNamespaces: true
`,
		},
		{
			name: "Test with labels/annotations on the inner resource",
			config: &ResourceDistributionPlugin{
				ObjectMeta: types.ObjectMeta{Name: "rd-with-resource-options"},
				ResourceArgs: ResourceArgs{
					ResourceName: "my-cm",
					ResourceKind: "ConfigMap",
					ResourceOptions: &types.GeneratorOptions{
						Labels:      map[string]string{"internal-label": "yes"},
						Annotations: map[string]string{"internal-annotation": "true"},
					},
				},
				TargetsArgs: TargetsArgs{
					AllNamespaces: true,
				},
			},
			expectedYAML: `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
metadata:
  name: rd-with-resource-options
spec:
  resource:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: my-cm
      labels:
        internal-label: "yes"
      annotations:
        internal-annotation: "true"
  targets:
    allNamespaces: true
`,
		},
		{
			name:          "Test error on missing name",
			config:        &ResourceDistributionPlugin{},
			expectedError: "a ResourceDistribution must have a name",
		},
		{
			name: "Test error on invalid resource kind",
			config: &ResourceDistributionPlugin{
				ObjectMeta: types.ObjectMeta{Name: "invalid-kind"},
				ResourceArgs: ResourceArgs{
					ResourceName: "foo",
					ResourceKind: "Deployment",
				},
			},
			expectedError: "resourceKind must be ConfigMap or Secret",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rn, err := MakeResourceDistribution(tc.config)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				return
			}

			require.NoError(t, err)

			actualYAML, err := rn.String()
			require.NoError(t, err)

			var expectedObj, actualObj unstructured.Unstructured
			err = yaml.Unmarshal([]byte(tc.expectedYAML), &expectedObj.Object)
			require.NoError(t, err, "Failed to unmarshal expected YAML")

			err = yaml.Unmarshal([]byte(actualYAML), &actualObj.Object)
			require.NoError(t, err, "Failed to unmarshal actual YAML")

			assert.Equal(t, expectedObj.Object, actualObj.Object)
		})
	}
}
