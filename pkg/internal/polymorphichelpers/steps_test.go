package polymorphichelpers

import (
	"encoding/json"
	"testing"

	"github.com/openkruise/kruise-rollout-api/rollouts/v1alpha1"
	"github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	_ = v1beta1.AddToScheme(scheme.Scheme)
}

func TestRolloutRollbackGetter(t *testing.T) {
	getRollout := func(currentIdx int32, steps []v1beta1.CanaryStep) []client.Object {
		canary := &v1beta1.Rollout{
			Status: v1beta1.RolloutStatus{
				CurrentStepIndex: currentIdx,
			},
		}
		canary.Spec.Strategy.Canary = &v1beta1.CanaryStrategy{
			Steps:                        steps,
			EnableExtraWorkloadForCanary: true,
		}
		canary.Status.CanaryStatus = &v1beta1.CanaryStatus{}

		blueGreen := canary.DeepCopy()
		blueGreen.Spec.Strategy.BlueGreen = &v1beta1.BlueGreenStrategy{
			Steps: steps,
		}
		blueGreen.Status.BlueGreenStatus = &v1beta1.BlueGreenStatus{}

		return []client.Object{canary, blueGreen}
	}
	newStep := func(replicas string, traffic string) v1beta1.CanaryStep {
		r := intstr.FromString(replicas)
		var trafficPtr *string
		if traffic != "" {
			trafficPtr = &traffic
		}
		return v1beta1.CanaryStep{
			Replicas: &r,
			TrafficRoutingStrategy: v1beta1.TrafficRoutingStrategy{
				Traffic: trafficPtr,
			},
		}
	}

	tests := []struct {
		name         string
		rollout      []client.Object
		targetStep   int32
		expectedStep int32
		expectedErr  string
	}{
		{
			name: "valid rollback to previous step",
			rollout: getRollout(3, []v1beta1.CanaryStep{
				newStep("10%", ""),
				newStep("20%", ""),
				newStep("30%", ""),
			}),
			targetStep:   2,
			expectedStep: 2,
		},
		{
			name: "invalid rollback to same or future step",
			rollout: getRollout(3, []v1beta1.CanaryStep{
				newStep("10%", ""),
				newStep("20%", ""),
				newStep("30%", ""),
			}),
			targetStep:  3,
			expectedErr: "specified step 3 is not a previous step (current step is 3)",
		},
		{
			name: "rollback to previous step with no traffic and most replicas",
			rollout: getRollout(4, []v1beta1.CanaryStep{
				newStep("10%", ""),
				newStep("20%", ""),
				newStep("30%", "50%"),
				newStep("40%", "50%"),
			}),
			targetStep:   -1,
			expectedStep: 2,
		},
		{
			name: "no previous step with no traffic found",
			rollout: getRollout(5, []v1beta1.CanaryStep{
				newStep("10%", "10%"),
				newStep("20%", "10%"),
				newStep("30%", "10%"),
				newStep("40%", "10%"),
				newStep("50%", "10%"),
			}),
			targetStep:  -1,
			expectedErr: "no previous step with no traffic found",
		},
		{
			name: "already at the first step",
			rollout: getRollout(1, []v1beta1.CanaryStep{
				newStep("10%", ""),
			}),
			targetStep:  2,
			expectedErr: "already at the first step, use kubectl-kruise rollout undo to cancel the release",
		},
		{
			name: "current step index out of range",
			rollout: getRollout(3, []v1beta1.CanaryStep{
				newStep("10%", ""),
				newStep("20%", ""),
			}),
			targetStep:  2,
			expectedErr: "has 2 steps, but current step is too large 3",
		},
		{
			name:        "not supported object",
			rollout:     []client.Object{&v1alpha1.Rollout{}},
			expectedErr: "rollback is not supported on given object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rollback := rolloutRollbackGetter(tt.targetStep)
			for _, rollout := range tt.rollout {
				data, err := rollback(rollout)
				if tt.expectedErr != "" {
					if err == nil || err.Error() != tt.expectedErr {
						t.Errorf("expected error %v, got %v", tt.expectedErr, err)
					}
					return
				}
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				var updatedRollout v1beta1.Rollout
				if err := json.Unmarshal(data, &updatedRollout); err != nil {
					t.Errorf("failed to unmarshal updated rollout: %v", err)
					return
				}

				var nextStepIndex int32
				if updatedRollout.Spec.Strategy.GetRollingStyle() == v1beta1.BlueGreenRollingStyle {
					nextStepIndex = updatedRollout.Status.BlueGreenStatus.NextStepIndex
				} else {
					nextStepIndex = updatedRollout.Status.CanaryStatus.NextStepIndex
				}

				if nextStepIndex != tt.expectedStep {
					t.Errorf("expected next step index %d, got %d", tt.expectedStep, nextStepIndex)
				}
			}
		})
	}
}
