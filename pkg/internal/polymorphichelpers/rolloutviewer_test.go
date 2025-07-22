/*
Copyright 2024 The Kubernetes Authors.

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
	"reflect"
	"testing"

	rolloutv1alpha1 "github.com/openkruise/kruise-rollout-api/rollouts/v1alpha1"
	rolloutv1beta1 "github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type mockObject struct{}

func (m *mockObject) GetObjectKind() schema.ObjectKind { return schema.EmptyObjectKind }
func (m *mockObject) DeepCopyObject() runtime.Object   { return m }

func TestRolloutViewer(t *testing.T) {
	testCases := []struct {
		name         string
		obj          runtime.Object
		expectError  bool
		expectedType reflect.Type
	}{
		{
			name:         "should handle v1beta1.Rollout",
			obj:          &rolloutv1beta1.Rollout{},
			expectError:  false,
			expectedType: reflect.TypeOf(&rolloutv1beta1.Rollout{}),
		},
		{
			name:         "should handle v1alpha1.Rollout",
			obj:          &rolloutv1alpha1.Rollout{},
			expectError:  false,
			expectedType: reflect.TypeOf(&rolloutv1alpha1.Rollout{}),
		},
		{
			name:         "should return an error for an unknown type",
			obj:          &mockObject{},
			expectError:  true,
			expectedType: nil,
		},
		{
			name:         "should return an error for a nil object",
			obj:          nil,
			expectError:  true,
			expectedType: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			viewer, err := rolloutViewer(tc.obj)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error but got: %v", err)
				}
				if viewer == nil {
					t.Errorf("expected a viewer but got nil")
				}
				viewerType := reflect.TypeOf(viewer)
				if viewerType != tc.expectedType {
					t.Errorf("expected viewer of type %v but got %v", tc.expectedType, viewerType)
				}
			}
		})
	}
}
