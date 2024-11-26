/*
Copyright 2018 The Kubernetes Authors.

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

	"k8s.io/apimachinery/pkg/runtime"

	rolloutv1alpha1 "github.com/openkruise/kruise-rollout-api/rollouts/v1alpha1"
	rolloutv1beta1 "github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
)

// statusViewer returns a StatusViewer for printing rollout status.
func rolloutViewer(obj runtime.Object) (interface{}, error) {
	if rollout, ok := obj.(*rolloutv1beta1.Rollout); ok {
		return rollout, nil
	}
	if rollout, ok := obj.(*rolloutv1alpha1.Rollout); ok {
		return rollout, nil
	}

	return nil, fmt.Errorf("unknown rollout type: %T", obj)
}
