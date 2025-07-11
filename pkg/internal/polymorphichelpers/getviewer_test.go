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

package polymorphichelpers

import (
	"fmt"
	"testing"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	kruisepolicyv1alpha1 "github.com/openkruise/kruise-api/policy/v1alpha1"
	rolloutv1alpha1 "github.com/openkruise/kruise-rollout-api/rollouts/v1alpha1"
	rolloutv1beta1 "github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

func TestGetViewer(t *testing.T) {
	supportedObjects := []runtime.Object{
		&kruiseappsv1alpha1.CloneSet{},
		&kruiseappsv1beta1.StatefulSet{},
		&kruiseappsv1alpha1.DaemonSet{},
		&rolloutv1beta1.Rollout{},
		&rolloutv1alpha1.Rollout{},
		&kruiseappsv1alpha1.BroadcastJob{},
		&kruiseappsv1alpha1.ContainerRecreateRequest{},
		&kruiseappsv1alpha1.AdvancedCronJob{},
		&kruiseappsv1alpha1.ResourceDistribution{},
		&kruiseappsv1alpha1.UnitedDeployment{},
		&kruiseappsv1alpha1.SidecarSet{},
		&kruiseappsv1alpha1.PodProbeMarker{},
		&kruiseappsv1alpha1.ImagePullJob{},
		&kruisepolicyv1alpha1.PodUnavailableBudget{},
	}

	unsupportedObject := &corev1.Pod{}

	for _, obj := range supportedObjects {
		objType := obj.GetObjectKind().GroupVersionKind().Kind
		if objType == "" {
			t.Run(fmt.Sprintf("SupportedType_%T", obj), func(t *testing.T) {
				viewer, err := getViewer(obj)

				if err != nil {
					t.Errorf("Expected no error for type %T, but got: %v", obj, err)
				}

				if viewer == nil {
					t.Errorf("Expected a viewer for type %T, but got nil", obj)
				}

				if _, ok := viewer.(printers.ResourcePrinter); !ok {
					t.Errorf("Expected viewer to be a ResourcePrinter for type %T", obj)
				}
			})
		}
	}

	t.Run(fmt.Sprintf("UnsupportedType_%T", unsupportedObject), func(t *testing.T) {
		viewer, err := getViewer(unsupportedObject)

		if err == nil {
			t.Errorf("Expected an error for unsupported type %T, but got nil", unsupportedObject)
		}

		expectedErrMsg := fmt.Sprintf("no viewer has been implemented for %T", unsupportedObject)
		if err != nil && err.Error() != expectedErrMsg {
			t.Errorf("Expected error message '%s', but got '%s'", expectedErrMsg, err.Error())
		}

		if viewer != nil {
			t.Errorf("Expected a nil viewer for unsupported type %T, but got a viewer", unsupportedObject)
		}
	})
}
