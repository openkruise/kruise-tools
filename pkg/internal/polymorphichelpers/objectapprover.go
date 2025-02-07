/*
Copyright 2021 The Kruise Authors.

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
	"fmt"

	rolloutsapiv1alpha1 "github.com/openkruise/kruise-rollout-api/rollouts/v1alpha1"
	rolloutsapiv1beta1 "github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/scheme"
)

// defaultObjectApprover currently only support Kruise Rollout.
func defaultObjectApprover(obj runtime.Object) ([]byte, error) {
	switch obj := obj.(type) {
	case *rolloutsapiv1alpha1.Rollout:
		if obj.Status.CanaryStatus == nil || obj.Status.CanaryStatus.CurrentStepState != rolloutsapiv1alpha1.CanaryStepStatePaused {
			return nil, errors.New("does not allow to approve, because current canary state is not 'StepPaused'")
		}
		obj.Status.CanaryStatus.CurrentStepState = rolloutsapiv1alpha1.CanaryStepStateReady
		return runtime.Encode(scheme.Codecs.LegacyCodec(rolloutsapiv1alpha1.GroupVersion), obj)
	case *rolloutsapiv1beta1.Rollout:
		switch {
		case obj.Status.CanaryStatus != nil:
			if obj.Status.CanaryStatus.CurrentStepState != rolloutsapiv1beta1.CanaryStepStatePaused {
				return nil, fmt.Errorf("does not allow to approve, because current canary state is not '%s'", rolloutsapiv1beta1.CanaryStepStatePaused)
			}
			obj.Status.CanaryStatus.CurrentStepState = rolloutsapiv1beta1.CanaryStepStateReady
		case obj.Status.BlueGreenStatus != nil:
			if obj.Status.BlueGreenStatus.CurrentStepState != rolloutsapiv1beta1.CanaryStepStatePaused {
				return nil, fmt.Errorf("does not allow to approve, because current blue-green state is not '%s'", rolloutsapiv1beta1.CanaryStepStatePaused)
			}
			obj.Status.BlueGreenStatus.CurrentStepState = rolloutsapiv1beta1.CanaryStepStateReady
		default:
			return nil, fmt.Errorf("no need to approve: not in canary or blue-green progress")
		}
		return runtime.Encode(scheme.Codecs.LegacyCodec(rolloutsapiv1beta1.GroupVersion), obj)

	default:
		return nil, fmt.Errorf("approving is not supported")
	}
}
