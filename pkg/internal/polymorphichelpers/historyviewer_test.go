/*
Copyright 2018 The Kubernetes Authors.
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

package polymorphichelpers

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes"
	corefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"

	kruiseclientsets "github.com/openkruise/kruise-api/client/clientset/versioned"
)

// fakeRESTGetterErr always fails ToRESTConfig
type fakeRESTGetterErr struct{}

func (f *fakeRESTGetterErr) ToRESTConfig() (*rest.Config, error) {
	return nil, errors.New("cfg fail")
}
func (f *fakeRESTGetterErr) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return nil, nil
}
func (f *fakeRESTGetterErr) ToRESTMapper() (meta.RESTMapper, error) {
	return nil, nil
}
func (f *fakeRESTGetterErr) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return nil
}

// fakeRESTGetterOK returns a dummy RESTConfig
type fakeRESTGetterOK struct{}

func (f *fakeRESTGetterOK) ToRESTConfig() (*rest.Config, error) {
	return &rest.Config{Host: "https://example"}, nil
}
func (f *fakeRESTGetterOK) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return memory.NewMemCacheClient(corefake.NewSimpleClientset().Discovery()), nil
}
func (f *fakeRESTGetterOK) ToRESTMapper() (meta.RESTMapper, error) {
	return nil, nil
}
func (f *fakeRESTGetterOK) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return nil
}

func TestHistoryViewer_ErrorPaths(t *testing.T) {
	mapping := &meta.RESTMapping{
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
	}

	if _, err := historyViewer(&fakeRESTGetterErr{}, mapping); err == nil || err.Error() != "cfg fail" {
		t.Fatalf("expected cfg fail, got %v", err)
	}

	origCore := coreNewForConfig
	coreNewForConfig = func(cfg *rest.Config) (*kubernetes.Clientset, error) {
		return nil, errors.New("core fail")
	}
	if _, err := historyViewer(&fakeRESTGetterOK{}, mapping); err == nil || err.Error() != "core fail" {
		t.Fatalf("expected core fail, got %v", err)
	}
	coreNewForConfig = origCore

	origKruise := kruiseNewForConfig
	kruiseNewForConfig = func(cfg *rest.Config) (*kruiseclientsets.Clientset, error) {
		return nil, errors.New("kruise fail")
	}
	if _, err := historyViewer(&fakeRESTGetterOK{}, mapping); err == nil || err.Error() != "kruise fail" {
		t.Fatalf("expected kruise fail, got %v", err)
	}
	kruiseNewForConfig = origKruise
}
