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
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestStatusViewer(t *testing.T) {
	testCases := []struct {
		name         string
		mapping      *meta.RESTMapping
		expectViewer bool
		expectError  bool
		expectPanic  bool
	}{
		{
			name: "Success for a known type (Deployment)",
			mapping: &meta.RESTMapping{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			},
			expectViewer: true,
			expectError:  false,
			expectPanic:  false,
		},
		{
			name: "Error for an unknown type",
			mapping: &meta.RESTMapping{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "example.com",
					Version: "v1",
					Kind:    "UnknownKind",
				},
			},
			expectViewer: false,
			expectError:  true,
			expectPanic:  false,
		},
		{
			name:         "Panic for a nil mapping",
			mapping:      nil,
			expectViewer: false,
			expectError:  true,
			expectPanic:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected a panic but did not get one")
					}
				}()
			}

			viewer, err := statusViewer(tc.mapping)

			if tc.expectError {
				if err == nil && !tc.expectPanic {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error but got: %v", err)
				}
			}

			// Check for a viewer
			if tc.expectViewer {
				if viewer == nil {
					t.Errorf("expected a StatusViewer but got nil")
				}
			} else {
				if viewer != nil {
					t.Errorf("did not expect a StatusViewer but got one")
				}
			}
		})
	}
}
