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

package cloneset

import (
	"context"
	"strings"
	"testing"

	kruiseapi "github.com/openkruise/kruise-api"
	appsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	"github.com/openkruise/kruise-tools/pkg/api"
	"github.com/openkruise/kruise-tools/pkg/creation"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreate(t *testing.T) {
	const (
		testNs          = "test-ns"
		srcDeployName   = "my-deploy"
		dstCloneSetName = "my-cloneset"
	)

	srcDeployment := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNs,
			Name:      srcDeployName,
		},
		Spec: apps.DeploymentSpec{
			Replicas: func() *int32 { r := int32(3); return &r }(),
		},
	}

	dstCloneSetExists := &appsv1alpha1.CloneSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNs,
			Name:      dstCloneSetName,
		},
	}

	srcRef := api.NewDeploymentRef(testNs, srcDeployName)
	dstRef := api.NewCloneSetRef(testNs, dstCloneSetName)

	testCases := []struct {
		name         string
		initObjs     []runtime.Object
		srcRef       api.ResourceRef
		dstRef       api.ResourceRef
		expectErrStr string
		validateFunc func(t *testing.T, c client.Client)
	}{
		{
			name:     "Success: create CloneSet from Deployment",
			initObjs: []runtime.Object{srcDeployment},
			srcRef:   srcRef,
			dstRef:   dstRef,
			validateFunc: func(t *testing.T, c client.Client) {
				cs := &appsv1alpha1.CloneSet{}
				err := c.Get(context.TODO(), dstRef.GetNamespacedName(), cs)
				if err != nil {
					t.Fatalf("Failed to get created CloneSet: %v", err)
				}
				if cs.Name != dstCloneSetName {
					t.Errorf("Expected CloneSet name %q, got %q", dstCloneSetName, cs.Name)
				}
				if *cs.Spec.Replicas != 3 {
					t.Errorf("Expected CloneSet replicas to be 3, got %d", *cs.Spec.Replicas)
				}
			},
		},
		{
			name:     "Error: invalid source type",
			initObjs: []runtime.Object{srcDeployment},
			srcRef: api.ResourceRef{
				APIVersion: "v1",
				Kind:       "Pod",
				Namespace:  testNs,
				Name:       "some-pod",
			},
			dstRef:       dstRef,
			expectErrStr: "invalid src type",
		},
		{
			name:         "Error: invalid destination type",
			initObjs:     []runtime.Object{srcDeployment},
			srcRef:       srcRef,
			dstRef:       api.NewDeploymentRef(testNs, "some-deploy"),
			expectErrStr: "invalid dst type",
		},
		{
			name:         "Error: destination CloneSet already exists",
			initObjs:     []runtime.Object{srcDeployment, dstCloneSetExists},
			srcRef:       srcRef,
			dstRef:       dstRef,
			expectErrStr: "already exists",
		},
		{
			name:         "Error: source Deployment not found",
			initObjs:     []runtime.Object{},
			srcRef:       srcRef,
			dstRef:       dstRef,
			expectErrStr: "failed to get",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			kruiseapi.AddToScheme(scheme)
			apps.AddToScheme(scheme)
			v1.AddToScheme(scheme)

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tc.initObjs...).Build()
			ctrl := &control{client: fakeClient}

			err := ctrl.Create(tc.srcRef, tc.dstRef, creation.Options{})

			if tc.expectErrStr != "" {
				if err == nil {
					t.Fatalf("Expected error containing %q, but got nil", tc.expectErrStr)
				}
				if !strings.Contains(err.Error(), tc.expectErrStr) {
					t.Errorf("Expected error message to contain %q, but got %q", tc.expectErrStr, err.Error())
				}
			} else if err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}

			if tc.validateFunc != nil {
				tc.validateFunc(t, fakeClient)
			}
		})
	}
}
