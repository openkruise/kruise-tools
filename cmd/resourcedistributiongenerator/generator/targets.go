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
	"sort"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func setTargets(rn *yaml.RNode, args *TargetsArgs) error {
	if args.AllNamespaces == false && args.IncludedNamespaces == nil && args.ExcludedNamespaces == nil &&
		(args.NamespaceLabelSelector == nil ||
			args.NamespaceLabelSelector.MatchLabels == nil && args.NamespaceLabelSelector.MatchExpressions == nil) {
		return errors.Errorf("The targets field of the ResourceDistribution cannot be empty")
	}

	err := setAllNs(rn, args.AllNamespaces)
	if err != nil {
		return err
	}

	err = setIncludedExcludedNs(rn, args.ExcludedNamespaces, excludedNamespacesPath)
	if err != nil {
		return err
	}

	err = setIncludedExcludedNs(rn, args.IncludedNamespaces, includedNamespacesPath)
	if err != nil {
		return err
	}

	err = setNsLabelSelector(rn, args.NamespaceLabelSelector)
	if err != nil {
		return err
	}

	return nil
}

// setIncludedExcludedNs set IncludedNamespaces Or ExcludedNamespaces for targets field
func setIncludedExcludedNs(rn *yaml.RNode, v []string, inExNsPath []string) error {
	if v == nil {
		return nil
	}
	if err := rn.SetMapField(newNameListRNode(v...), append(inExNsPath, listField)...); err != nil {
		return err
	}
	return nil
}

// newNameListRNode returns a new List *RNode
// containing the provided scalar values prefixed with a string of name.
func newNameListRNode(values ...string) *yaml.RNode {
	matchSeq := &yaml.Node{Kind: yaml.SequenceNode}
	for _, v := range values {
		node := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: v,
		}
		matchSeq.Content = append(matchSeq.Content, node)

	}
	return yaml.NewRNode(matchSeq)
}

// setAllNs set AllNamespaces true for targets field
// The allNamespaces field is false by default and is not displayed
func setAllNs(rn *yaml.RNode, allNs bool) error {
	if !allNs {
		return nil
	}
	allNamespaces := strconv.FormatBool(allNs)
	n := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: allNamespaces,
		Tag:   yaml.NodeTagBool,
	}
	if err := rn.SetMapField(yaml.NewRNode(n), append(targetsPath, allNamespacesField)...); err != nil {
		return err
	}
	return nil
}

func setNsLabelSelector(rn *yaml.RNode, sel *LabelSelector) error {
	if sel == nil {
		return nil
	}

	err := setMatchExpressions(rn, sel.MatchExpressions)
	if err != nil {
		return err
	}

	err = setMatchLabels(rn, sel.MatchLabels)
	if err != nil {
		return err
	}

	return nil
}

func setMatchLabels(rn *yaml.RNode, matchLabels map[string]string) error {
	if matchLabels == nil {
		return nil
	}
	for _, k := range yaml.SortedMapKeys(matchLabels) {
		v := matchLabels[k]
		if err := rn.SetMapField(yaml.NewStringRNode(v), append(MatchLabelsPath, k)...); err != nil {
			return err
		}
	}
	return nil
}

func setMatchExpressions(rn *yaml.RNode, args []LabelSelectorRequirement) error {
	if args == nil {
		return nil
	}

	sort.Slice(args, func(i, j int) bool {
		return strings.Compare(args[i].Key, args[j].Key) < 0
	})

	// matchExpList will be seted to the matchExpressions field
	matchExpList := &yaml.Node{Kind: yaml.SequenceNode}
	for _, matchExpArgs := range args {
		matchExpElement := &yaml.Node{
			Kind: yaml.MappingNode,
		}

		// add key for matchExpression
		if matchExpArgs.Key == "" {
			return errors.Errorf("the field " +
				"ResourceDistribution.targets.namespaceLabelSelector.matchExpressions.key cannot be empty")
		}
		matchExpElement.Content = append(matchExpElement.Content, newNode(keyField), newNode(matchExpArgs.Key))

		// add operator for matchExpression
		if matchExpArgs.Operator != metav1.LabelSelectorOpIn &&
			matchExpArgs.Operator != metav1.LabelSelectorOpNotIn &&
			matchExpArgs.Operator != metav1.LabelSelectorOpExists &&
			matchExpArgs.Operator != metav1.LabelSelectorOpDoesNotExist {
			return errors.Errorf("the field " +
				"ResourceDistribution.targets.namespaceLabelSelector.matchExpressions.operator is invalid")
		}
		operator := string(matchExpArgs.Operator)
		matchExpElement.Content = append(matchExpElement.Content, newNode(operatorField), newNode(operator))

		if matchExpArgs.Operator == metav1.LabelSelectorOpIn ||
			matchExpArgs.Operator == metav1.LabelSelectorOpNotIn {
			// add values for matchExpression
			if matchExpArgs.Values == nil {
				return errors.Errorf("the field " +
					"ResourceDistribution.targets.namespaceLabelSelector.matchExpressions.values for In and NotIn cannot be empty")
			}
			valSeq := &yaml.Node{Kind: yaml.SequenceNode}
			knowValue := make(map[string]bool)
			sort.Strings(matchExpArgs.Values)
			for _, val := range matchExpArgs.Values {
				if val == "" {
					return errors.Errorf("the element of field " +
						"ResourceDistribution.targets.namespaceLabelSelector.matchExpressions.values is invalid")
				}
				if knowValue[val] {
					return errors.Errorf("the element of field " +
						"ResourceDistribution.targets.namespaceLabelSelector.matchExpressions.values cannot be repeated")
				}
				valSeq.Content = append(valSeq.Content, newNode(val))
				knowValue[val] = true
			}
			matchExpElement.Content = append(matchExpElement.Content, newNode(valuesField), valSeq)
		} else {
			if matchExpArgs.Values != nil {
				return errors.Errorf("the field " +
					"ResourceDistribution.targets.namespaceLabelSelector.matchExpressions.values for Exist and DoesNotExist must be empty")
			}
		}

		// element is added to the list
		matchExpList.Content = append(matchExpList.Content, matchExpElement)
	}

	err := rn.SetMapField(yaml.NewRNode(matchExpList), append(NamespaceLabelSelectorPath, matchExpressionsField)...)
	if err != nil {
		return err
	}
	return nil
}

func newNode(value string) *yaml.Node {
	return &yaml.Node{
		Kind: yaml.ScalarNode, Value: value,
	}
}
