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
	"errors"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
)

type mockRESTClientGetter struct {
	config *rest.Config
	err    error
}

func (m *mockRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return m.config, m.err
}

func (m *mockRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return nil, errors.New("not implemented")
}
func (m *mockRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	return nil, errors.New("not implemented")
}
func (m *mockRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return nil
}

func TestAttachablePodForObject(t *testing.T) {
	// Test objects
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod"},
	}
	testService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "unsupported-service"},
	}

	tests := []struct {
		name          string
		object        runtime.Object
		getter        genericclioptions.RESTClientGetter
		expectedPod   *corev1.Pod
		expectedError string
	}{
		{
			name:        "Direct pod object returns itself",
			object:      testPod,
			getter:      nil,
			expectedPod: testPod,
		},
		{
			name:          "REST config error",
			object:        testService,
			getter:        &mockRESTClientGetter{err: errors.New("fake config error")},
			expectedError: "fake config error",
		},
		{
			name:          "Selector error for unsupported type",
			object:        testService,
			getter:        &mockRESTClientGetter{config: &rest.Config{}},
			expectedError: "cannot attach to *v1.Service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod, err := attachablePodForObject(tt.getter, tt.object, 10*time.Second)

			if tt.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
				}
				return
			}

			// Validate success cases
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if pod.Name != tt.expectedPod.Name {
				t.Errorf("Expected pod %q, got %q", tt.expectedPod.Name, pod.Name)
			}
		})
	}
}
