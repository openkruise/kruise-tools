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

package utils

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Helper function for string pointers
func strPtr(s string) *string { return &s }

func TestIsKruiseRolloutsAnnotation(t *testing.T) {
	testCases := []struct {
		name     string
		input    *string
		expected bool
	}{
		{name: "Nil string", input: nil, expected: false},
		{name: "Kruise prefix", input: strPtr("rollouts.kruise.io/annotation"), expected: true},
		{name: "Non-matching string", input: strPtr("other.domain/key"), expected: false},
		{name: "Exact prefix", input: strPtr("rollouts.kruise.io/"), expected: true},
		{name: "Empty string", input: strPtr(""), expected: false},
		{name: "Substring match", input: strPtr("pre/rollouts.kruise.io/suffix"), expected: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsKruiseRolloutsAnnotation(tc.input); got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestInCanaryProgress(t *testing.T) {
	testCases := []struct {
		name       string
		deployment *appsv1.Deployment
		expected   bool
	}{
		{
			name: "Not paused",
			deployment: &appsv1.Deployment{
				Spec:       appsv1.DeploymentSpec{Paused: false},
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{InRolloutProgressingAnnotation: "true"}},
			},
			expected: false,
		},
		{
			name: "Missing progressing annotation",
			deployment: &appsv1.Deployment{
				Spec:       appsv1.DeploymentSpec{Paused: true},
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
			},
			expected: false,
		},
		{
			name: "Has strategy annotation",
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{Paused: true},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						InRolloutProgressingAnnotation: "true",
						DeploymentStrategyAnnotation:   "partition",
					},
				},
			},
			expected: false,
		},
		{
			name: "Valid canary state",
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{Paused: true},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{InRolloutProgressingAnnotation: "true"},
				},
			},
			expected: true,
		},
		{
			name: "Nil annotations",
			deployment: &appsv1.Deployment{
				Spec:       appsv1.DeploymentSpec{Paused: true},
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: false,
		},
		{
			name: "Explicit empty annotations",
			deployment: &appsv1.Deployment{
				Spec:       appsv1.DeploymentSpec{Paused: true},
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
			},
			expected: false,
		},
		{
			name: "Only strategy annotation present",
			deployment: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{Paused: true},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{DeploymentStrategyAnnotation: "partition"},
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := InCanaryProgress(tc.deployment); got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
