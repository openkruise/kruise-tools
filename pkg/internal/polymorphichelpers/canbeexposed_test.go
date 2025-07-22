/*
Copyright 2025 The Kubernetes Authors.

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

package polymorphichelpers

import (
	"fmt"
	"testing"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCanBeExposed(t *testing.T) {
	tests := []struct {
		name        string
		kind        schema.GroupKind
		expectError bool
	}{
		{
			name:        "ReplicationController should be exposable",
			kind:        corev1.SchemeGroupVersion.WithKind("ReplicationController").GroupKind(),
			expectError: false,
		},
		{
			name:        "Service should be exposable",
			kind:        corev1.SchemeGroupVersion.WithKind("Service").GroupKind(),
			expectError: false,
		},
		{
			name:        "Pod should be exposable",
			kind:        corev1.SchemeGroupVersion.WithKind("Pod").GroupKind(),
			expectError: false,
		},
		{
			name:        "apps/v1 Deployment should be exposable",
			kind:        appsv1.SchemeGroupVersion.WithKind("Deployment").GroupKind(),
			expectError: false,
		},
		{
			name:        "apps/v1 ReplicaSet should be exposable",
			kind:        appsv1.SchemeGroupVersion.WithKind("ReplicaSet").GroupKind(),
			expectError: false,
		},
		{
			name:        "extensions/v1beta1 Deployment should be exposable",
			kind:        extensionsv1beta1.SchemeGroupVersion.WithKind("Deployment").GroupKind(),
			expectError: false,
		},
		{
			name:        "extensions/v1beta1 ReplicaSet should be exposable",
			kind:        extensionsv1beta1.SchemeGroupVersion.WithKind("ReplicaSet").GroupKind(),
			expectError: false,
		},
		{
			name:        "CloneSet should be exposable",
			kind:        kruiseappsv1alpha1.SchemeGroupVersion.WithKind("CloneSet").GroupKind(),
			expectError: false,
		},
		{
			name:        "ConfigMap should not be exposable",
			kind:        corev1.SchemeGroupVersion.WithKind("ConfigMap").GroupKind(),
			expectError: true,
		},
		{
			name:        "Secret should not be exposable",
			kind:        corev1.SchemeGroupVersion.WithKind("Secret").GroupKind(),
			expectError: true,
		},
		{
			name:        "StatefulSet should not be exposable",
			kind:        appsv1.SchemeGroupVersion.WithKind("StatefulSet").GroupKind(),
			expectError: true,
		},
		{
			name:        "DaemonSet should not be exposable",
			kind:        appsv1.SchemeGroupVersion.WithKind("DaemonSet").GroupKind(),
			expectError: true,
		},
		{
			name:        "Custom resource should not be exposable",
			kind:        schema.GroupKind{Group: "example.com", Kind: "CustomResource"},
			expectError: true,
		},
		{
			name:        "Empty kind should not be exposable",
			kind:        schema.GroupKind{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := canBeExposed(tt.kind)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for kind %v, but got nil", tt.kind)
				}
				// Verify error message format
				expectedMsg := fmt.Sprintf("cannot expose a %s", tt.kind)
				if err.Error() != expectedMsg {
					t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error for kind %v, but got: %v", tt.kind, err)
				}
			}
		})
	}
}

func TestCanBeExposedErrorMessage(t *testing.T) {
	unsupportedKind := schema.GroupKind{Group: "test.example.com", Kind: "UnsupportedKind"}
	err := canBeExposed(unsupportedKind)

	if err == nil {
		t.Fatal("expected error for unsupported kind")
	}

	expectedMsg := "cannot expose a UnsupportedKind.test.example.com"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// Benchmark test to ensure the function performs well
func BenchmarkCanBeExposed(b *testing.B) {
	supportedKind := appsv1.SchemeGroupVersion.WithKind("Deployment").GroupKind()
	unsupportedKind := schema.GroupKind{Group: "example.com", Kind: "CustomResource"}

	b.Run("Supported", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = canBeExposed(supportedKind)
		}
	})

	b.Run("Unsupported", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = canBeExposed(unsupportedKind)
		}
	})
}
