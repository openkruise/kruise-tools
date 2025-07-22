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

package conversion

import (
	"testing"

	appsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	"github.com/stretchr/testify/assert"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestDeploymentToCloneSet(t *testing.T) {
	cases := []struct {
		name            string
		deployment      *apps.Deployment
		dstCloneSetName string
		expected        *appsv1alpha1.CloneSet
	}{
		{
			name: "Test basic deployment conversion",
			deployment: &apps.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "default",
					Name:        "test-deploy",
					Labels:      map[string]string{"app": "nginx"},
					Annotations: map[string]string{"foo": "bar"},
				},
				Spec: apps.DeploymentSpec{
					Replicas: int32Ptr(3),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "nginx"}},
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "nginx"}},
						Spec:       v1.PodSpec{Containers: []v1.Container{{Name: "nginx", Image: "nginx:latest"}}},
					},
					MinReadySeconds: 5,
				},
			},
			dstCloneSetName: "my-cloneset",
			expected: &appsv1alpha1.CloneSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "default",
					Name:        "my-cloneset",
					Labels:      map[string]string{"app": "nginx"},
					Annotations: map[string]string{"foo": "bar"},
				},
				Spec: appsv1alpha1.CloneSetSpec{
					Replicas: int32Ptr(3),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "nginx"}},
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "nginx"}},
						Spec:       v1.PodSpec{Containers: []v1.Container{{Name: "nginx", Image: "nginx:latest"}}},
					},
					MinReadySeconds: 5,
					UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
						Type: appsv1alpha1.RecreateCloneSetUpdateStrategyType,
					},
				},
			},
		},
		{
			name: "Test deployment with rolling update strategy and other fields",
			deployment: &apps.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "default",
					Name:       "test-deploy-rolling",
					Finalizers: []string{"finalizer.kruise.io/foo"},
				},
				Spec: apps.DeploymentSpec{
					Replicas:             int32Ptr(2),
					Selector:             &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
					RevisionHistoryLimit: int32Ptr(10),
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test"}},
						Spec:       v1.PodSpec{Containers: []v1.Container{{Name: "test", Image: "test:latest"}}},
					},
					Strategy: apps.DeploymentStrategy{
						Type: apps.RollingUpdateDeploymentStrategyType,
						RollingUpdate: &apps.RollingUpdateDeployment{
							MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
							MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
						},
					},
					Paused: true,
				},
			},
			dstCloneSetName: "my-rolling-cloneset",
			expected: &appsv1alpha1.CloneSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:  "default",
					Name:       "my-rolling-cloneset",
					Finalizers: []string{"finalizer.kruise.io/foo"},
				},
				Spec: appsv1alpha1.CloneSetSpec{
					Replicas:             int32Ptr(2),
					Selector:             &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
					RevisionHistoryLimit: int32Ptr(10),
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test"}},
						Spec:       v1.PodSpec{Containers: []v1.Container{{Name: "test", Image: "test:latest"}}},
					},
					UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
						Type:           appsv1alpha1.RecreateCloneSetUpdateStrategyType,
						Paused:         true,
						MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
						MaxSurge:       &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
					},
				},
			},
		},
		{
			name: "Test deployment with zero values and nil rolling update",
			deployment: &apps.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "default",
					Name:        "test-deploy-zero",
					Labels:      nil,
					Annotations: map[string]string{},
				},
				Spec: apps.DeploymentSpec{
					Replicas:        int32Ptr(0),
					Selector:        &metav1.LabelSelector{},
					MinReadySeconds: 0,
					Template:        v1.PodTemplateSpec{},
					Strategy: apps.DeploymentStrategy{
						Type:          apps.RollingUpdateDeploymentStrategyType,
						RollingUpdate: nil,
					},
				},
			},
			dstCloneSetName: "my-zero-cloneset",
			expected: &appsv1alpha1.CloneSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "default",
					Name:        "my-zero-cloneset",
					Labels:      nil,
					Annotations: map[string]string{},
				},
				Spec: appsv1alpha1.CloneSetSpec{
					Replicas:        int32Ptr(0),
					Selector:        &metav1.LabelSelector{},
					MinReadySeconds: 0,
					Template:        v1.PodTemplateSpec{},
					UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
						Type: appsv1alpha1.RecreateCloneSetUpdateStrategyType,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cs := DeploymentToCloneSet(tc.deployment, tc.dstCloneSetName)
			assert.Equal(t, tc.expected, cs, "The converted CloneSet should match the expected CloneSet.")
		})
	}
}

func int32Ptr(i int32) *int32 { return &i }
