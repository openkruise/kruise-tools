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

	kruiserolloutsv1apha1 "github.com/openkruise/rollouts/api/v1alpha1"
	kruiserolloutsv1beta1 "github.com/openkruise/rollouts/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/scheme"
)

// defaultObjectApprover currently only support Kruise Rollout.
func defaultObjectApprover(obj runtime.Object) ([]byte, error) {
	switch obj := obj.(type) {
	case *kruiserolloutsv1apha1.Rollout:
		if obj.Status.CanaryStatus == nil || obj.Status.CanaryStatus.CurrentStepState != kruiserolloutsv1apha1.CanaryStepStatePaused {
			return nil, errors.New("does not allow to approve, because current canary state is not 'StepPaused'")
		}
		obj.Status.CanaryStatus.CurrentStepState = kruiserolloutsv1apha1.CanaryStepStateReady
		return runtime.Encode(scheme.Codecs.LegacyCodec(kruiserolloutsv1apha1.GroupVersion), obj)
	case *kruiserolloutsv1beta1.Rollout:
		if obj.Status.CanaryStatus == nil || obj.Status.CanaryStatus.CurrentStepState != kruiserolloutsv1beta1.CanaryStepStatePaused {
			return nil, errors.New("does not allow to approve, because current canary state is not 'StepPaused'")
		}
		obj.Status.CanaryStatus.CurrentStepState = kruiserolloutsv1beta1.CanaryStepStateReady
		return runtime.Encode(scheme.Codecs.LegacyCodec(kruiserolloutsv1beta1.GroupVersion), obj)
	default:
		return nil, fmt.Errorf("approving is not supported")
	}
}
