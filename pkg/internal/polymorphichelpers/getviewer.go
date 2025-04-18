/*
Copyright 2024 The Kruise Authors.

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

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	kruisepolicyv1alpha1 "github.com/openkruise/kruise-api/policy/v1alpha1"
	rolloutv1alpha1 "github.com/openkruise/kruise-rollout-api/rollouts/v1alpha1"
	rolloutv1beta1 "github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

// getViewer returns a printer for Kruise resources
func getViewer(obj runtime.Object) (interface{}, error) {
	switch obj.(type) {
	case *kruiseappsv1alpha1.CloneSet,
		*kruiseappsv1beta1.StatefulSet,
		*kruiseappsv1alpha1.DaemonSet,
		*rolloutv1beta1.Rollout,
		*rolloutv1alpha1.Rollout,
		*kruiseappsv1alpha1.BroadcastJob,
		*kruiseappsv1alpha1.ContainerRecreateRequest,
		*kruiseappsv1alpha1.AdvancedCronJob,
		*kruiseappsv1alpha1.ResourceDistribution,
		*kruiseappsv1alpha1.UnitedDeployment,
		*kruiseappsv1alpha1.SidecarSet,
		*kruiseappsv1alpha1.PodProbeMarker,
		*kruiseappsv1alpha1.ImagePullJob,
		*kruisepolicyv1alpha1.PodUnavailableBudget:
		return printers.NewTablePrinter(printers.PrintOptions{
			WithKind:      true,
			WithNamespace: true,
		}), nil
	default:
		return nil, fmt.Errorf("no viewer has been implemented for %T", obj)
	}
}
