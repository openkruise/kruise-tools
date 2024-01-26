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

	rolloutsapi "github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/scheme"
)

// defaultObjectApprover currently only support Kruise Rollout.
func defaultObjectApprover(obj runtime.Object) ([]byte, error) {
	switch obj := obj.(type) {
	case *rolloutsapi.Rollout:
		if obj.Status.CanaryStatus == nil || obj.Status.CanaryStatus.CurrentStepState != rolloutsapi.CanaryStepStatePaused {
			return nil, errors.New("does not allow to approve, because current canary state is not 'StepPaused'")
		}
		obj.Status.CanaryStatus.CurrentStepState = rolloutsapi.CanaryStepStateReady
		return runtime.Encode(scheme.Codecs.LegacyCodec(rolloutsapi.GroupVersion), obj)

	default:
		return nil, fmt.Errorf("approving is not supported")
	}
}
