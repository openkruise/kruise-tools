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

const tmpl = `
apiVersion: apps.kruise.io/v1alpha1
kind: ResourceDistribution
spec:
  resource:
    apiVersion: v1
`

// Field names
const (
	kindField             = "kind"
	nameField             = "name"
	listField             = "list"
	allNamespacesField    = "allNamespaces"
	immutableField        = "immutable"
	typeField             = "type"
	matchExpressionsField = "matchExpressions"
	keyField              = "key"
	operatorField         = "operator"
	valuesField           = "values"
)

var metadataLabelsPath = []string{"metadata", "labels"}
var metadataAnnotationsPath = []string{"metadata", "annotations"}
var resourcePath = []string{"spec", "resource"}
var metadataPath = []string{"spec", "resource", "metadata"}
var resourceLabelsPath = []string{"spec", "resource", "metadata", "labels"}
var resourceAnnotationsPath = []string{"spec", "resource", "metadata", "annotations"}
var targetsPath = []string{"spec", "targets"}
var includedNamespacesPath = []string{"spec", "targets", "includedNamespaces"}
var excludedNamespacesPath = []string{"spec", "targets", "excludedNamespaces"}
var NamespaceLabelSelectorPath = []string{"spec", "targets", "namespaceLabelSelector"}
var MatchLabelsPath = []string{"spec", "targets", "namespaceLabelSelector", "matchLabels"}
