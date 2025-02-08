package polymorphichelpers

import (
	"fmt"

	"github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/scheme"
)

func rolloutRollbackGetter(targetStep int32) func(runtime.Object) ([]byte, error) {
	return func(obj runtime.Object) ([]byte, error) {
		switch rollout := obj.(type) {
		case *v1beta1.Rollout:
			steps := rollout.Spec.Strategy.GetSteps()
			curStep := rollout.Status.CurrentStepIndex
			if len(steps) < int(curStep) {
				return nil, fmt.Errorf("has %d steps, but current step is too large %d", len(steps), curStep)
			}
			if curStep <= 1 {
				return nil, fmt.Errorf("already at the first step, use kubectl-kruise rollout undo to cancel the release")
			}
			if targetStep == -1 {
				s, err := findPreviousStepWithNoTrafficAndMostReplicas(steps, rollout.Status.CurrentStepIndex)
				if err != nil {
					return nil, err
				}
				targetStep = s
			}
			if targetStep >= rollout.Status.CurrentStepIndex {
				return nil, fmt.Errorf("specified step %d is not a previous step (current step is %d)", targetStep, rollout.Status.CurrentStepIndex)
			}
			style := rollout.Spec.Strategy.GetRollingStyle()
			switch style {
			case v1beta1.BlueGreenRollingStyle:
				rollout.Status.BlueGreenStatus.NextStepIndex = targetStep
			default:
				// canary and partition
				rollout.Status.CanaryStatus.NextStepIndex = targetStep
			}
			return runtime.Encode(scheme.Codecs.LegacyCodec(v1beta1.GroupVersion), rollout)
		default:
			return nil, fmt.Errorf("rollback is not supported on given object")
		}
	}
}

func findPreviousStepWithNoTrafficAndMostReplicas(steps []v1beta1.CanaryStep, curStep int32) (int32, error) {
	maxReplicas := 0
	var targetStep int32 = -1
	for i := curStep - 2; i >= 0; i-- {
		step := steps[i]
		if hasTraffic(step) {
			klog.V(5).InfoS("has traffic", "step", i+1)
			continue
		}
		replicas, _ := intstr.GetScaledValueFromIntOrPercent(step.Replicas, 100, true)
		klog.V(5).InfoS("replicas percent", "percent", replicas, "step", i+1, "obj", step)
		if replicas > maxReplicas {
			maxReplicas = replicas
			targetStep = i + 1
		}
	}
	if targetStep == -1 {
		return 0, fmt.Errorf("no previous step with no traffic found")
	}
	return targetStep, nil
}

func hasTraffic(step v1beta1.CanaryStep) bool {
	if step.Traffic == nil {
		return false
	}
	is := intstr.FromString(*step.Traffic)
	trafficPercent, _ := intstr.GetScaledValueFromIntOrPercent(&is, 100, true)
	return trafficPercent != 0
}
