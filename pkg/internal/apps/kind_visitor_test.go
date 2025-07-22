/*
Copyright 2021 The Kruise Authors.
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

package apps

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type mockKindVisitor struct {
	visited string
}

func (v *mockKindVisitor) VisitDaemonSet(kind GroupKindElement)  { v.visited = "DaemonSet" }
func (v *mockKindVisitor) VisitDeployment(kind GroupKindElement) { v.visited = "Deployment" }
func (v *mockKindVisitor) VisitJob(kind GroupKindElement)        { v.visited = "Job" }
func (v *mockKindVisitor) VisitPod(kind GroupKindElement)        { v.visited = "Pod" }
func (v *mockKindVisitor) VisitReplicaSet(kind GroupKindElement) { v.visited = "ReplicaSet" }
func (v *mockKindVisitor) VisitReplicationController(kind GroupKindElement) {
	v.visited = "ReplicationController"
}
func (v *mockKindVisitor) VisitStatefulSet(kind GroupKindElement) { v.visited = "StatefulSet" }
func (v *mockKindVisitor) VisitCronJob(kind GroupKindElement)     { v.visited = "CronJob" }
func (v *mockKindVisitor) VisitCloneSet(kind GroupKindElement)    { v.visited = "CloneSet" }
func (v *mockKindVisitor) VisitAdvancedStatefulSet(kind GroupKindElement) {
	v.visited = "AdvancedStatefulSet"
}
func (v *mockKindVisitor) VisitAdvancedDaemonSet(kind GroupKindElement) {
	v.visited = "AdvancedDaemonSet"
}
func (v *mockKindVisitor) VisitRollout(kind GroupKindElement) { v.visited = "Rollout" }

func TestGroupKindElement_Accept(t *testing.T) {
	testCases := []struct {
		name          string
		element       GroupKindElement
		expectedVisit string
		expectError   bool
	}{
		{
			name:          "DaemonSet in apps group",
			element:       GroupKindElement(schema.GroupKind{Group: "apps", Kind: "DaemonSet"}),
			expectedVisit: "DaemonSet",
		},
		{
			name:          "DaemonSet in extensions group",
			element:       GroupKindElement(schema.GroupKind{Group: "extensions", Kind: "DaemonSet"}),
			expectedVisit: "DaemonSet",
		},
		{
			name:          "Deployment in apps group",
			element:       GroupKindElement(schema.GroupKind{Group: "apps", Kind: "Deployment"}),
			expectedVisit: "Deployment",
		},
		{
			name:          "Job in batch group",
			element:       GroupKindElement(schema.GroupKind{Group: "batch", Kind: "Job"}),
			expectedVisit: "Job",
		},
		{
			name:          "Pod in core group",
			element:       GroupKindElement(schema.GroupKind{Group: "", Kind: "Pod"}),
			expectedVisit: "Pod",
		},
		{
			name:          "ReplicaSet in extensions group",
			element:       GroupKindElement(schema.GroupKind{Group: "extensions", Kind: "ReplicaSet"}),
			expectedVisit: "ReplicaSet",
		},
		{
			name:          "ReplicationController in core group",
			element:       GroupKindElement(schema.GroupKind{Group: "core", Kind: "ReplicationController"}),
			expectedVisit: "ReplicationController",
		},
		{
			name:          "StatefulSet in apps group",
			element:       GroupKindElement(schema.GroupKind{Group: "apps", Kind: "StatefulSet"}),
			expectedVisit: "StatefulSet",
		},
		{
			name:          "CronJob in batch group",
			element:       GroupKindElement(schema.GroupKind{Group: "batch", Kind: "CronJob"}),
			expectedVisit: "CronJob",
		},
		{
			name:          "CloneSet",
			element:       GroupKindElement(schema.GroupKind{Group: "apps.kruise.io", Kind: "CloneSet"}),
			expectedVisit: "CloneSet",
		},
		{
			name:          "AdvancedStatefulSet",
			element:       GroupKindElement(schema.GroupKind{Group: "apps.kruise.io", Kind: "StatefulSet"}),
			expectedVisit: "AdvancedStatefulSet",
		},
		{
			name:          "AdvancedDaemonSet",
			element:       GroupKindElement(schema.GroupKind{Group: "apps.kruise.io", Kind: "DaemonSet"}),
			expectedVisit: "AdvancedDaemonSet",
		},
		{
			name:          "Rollout",
			element:       GroupKindElement(schema.GroupKind{Group: "rollouts.kruise.io", Kind: "Rollout"}),
			expectedVisit: "Rollout",
		},
		{
			name:        "Unsupported Kind",
			element:     GroupKindElement(schema.GroupKind{Group: "unsupported.group", Kind: "UnsupportedKind"}),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			visitor := &mockKindVisitor{}
			err := tc.element.Accept(visitor)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error for %v, but got none", tc.element)
				}
				expectedErr := fmt.Sprintf("no visitor method exists for %v", tc.element)
				if err.Error() != expectedErr {
					t.Errorf("Expected error message '%s', but got '%s'", expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error for %v, but got: %v", tc.element, err)
				}
				if visitor.visited != tc.expectedVisit {
					t.Errorf("Expected to visit %s, but visited %s", tc.expectedVisit, visitor.visited)
				}
			}
		})
	}
}

func TestGroupKindElement_GroupMatch(t *testing.T) {
	testCases := []struct {
		name     string
		element  GroupKindElement
		groups   []string
		expected bool
	}{
		{
			name:     "Single group match",
			element:  GroupKindElement(schema.GroupKind{Group: "apps"}),
			groups:   []string{"apps"},
			expected: true,
		},
		{
			name:     "Multiple groups, first one matches",
			element:  GroupKindElement(schema.GroupKind{Group: "extensions"}),
			groups:   []string{"extensions", "apps"},
			expected: true,
		},
		{
			name:     "Multiple groups, last one matches",
			element:  GroupKindElement(schema.GroupKind{Group: "core"}),
			groups:   []string{"apps", "extensions", "core"},
			expected: true,
		},
		{
			name:     "No match",
			element:  GroupKindElement(schema.GroupKind{Group: "batch"}),
			groups:   []string{"apps", "extensions"},
			expected: false,
		},
		{
			name:     "Empty group with empty string match",
			element:  GroupKindElement(schema.GroupKind{Group: ""}),
			groups:   []string{""},
			expected: true,
		},
		{
			name:     "Empty group with non-empty string no match",
			element:  GroupKindElement(schema.GroupKind{Group: ""}),
			groups:   []string{"apps"},
			expected: false,
		},
		{
			name:     "No groups to match",
			element:  GroupKindElement(schema.GroupKind{Group: "apps"}),
			groups:   []string{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.element.GroupMatch(tc.groups...)
			if result != tc.expected {
				t.Errorf("Expected GroupMatch to be %v, but got %v", tc.expected, result)
			}
		})
	}
}
