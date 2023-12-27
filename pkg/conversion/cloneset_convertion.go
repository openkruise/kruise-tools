/*
Copyright 2020 The Kruise Authors.

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
	appsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Convert Deployment to CloneSet
func DeploymentToCloneSet(deploy *apps.Deployment, dstCloneSetName string) *appsv1alpha1.CloneSet {
	// Deep copy first
	from := deploy.DeepCopy()

	cs := &appsv1alpha1.CloneSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   from.Namespace,
			Name:        dstCloneSetName,
			Labels:      from.Labels,
			Annotations: from.Annotations,
			Finalizers:  from.Finalizers,
			//ClusterName: from.ClusterName,
		},
		Spec: appsv1alpha1.CloneSetSpec{
			Replicas:             from.Spec.Replicas,
			Selector:             from.Spec.Selector,
			Template:             from.Spec.Template,
			RevisionHistoryLimit: from.Spec.RevisionHistoryLimit,
			MinReadySeconds:      from.Spec.MinReadySeconds,
			UpdateStrategy: appsv1alpha1.CloneSetUpdateStrategy{
				Type:   appsv1alpha1.RecreateCloneSetUpdateStrategyType,
				Paused: deploy.Spec.Paused,
			},
		},
	}

	if from.Spec.Strategy.RollingUpdate != nil {
		if from.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
			cs.Spec.UpdateStrategy.MaxUnavailable = from.Spec.Strategy.RollingUpdate.MaxUnavailable
		}
		if from.Spec.Strategy.RollingUpdate.MaxSurge != nil {
			cs.Spec.UpdateStrategy.MaxSurge = from.Spec.Strategy.RollingUpdate.MaxSurge
		}
	}
	return cs
}
