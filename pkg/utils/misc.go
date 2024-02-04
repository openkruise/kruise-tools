package utils

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
)

const InRolloutProgressingAnnotation = "rollouts.kruise.io/in-progressing"
const DeploymentStrategyAnnotation = "rollouts.kruise.io/deployment-strategy"

func IsKruiseRolloutsAnnotation(s *string) bool {
	if s == nil {
		return false
	}
	const prefix = "rollouts.kruise.io/"
	return strings.Contains(*s, prefix)
}

func InCanaryProgress(deployment *appsv1.Deployment) bool {
	if !deployment.Spec.Paused {
		return false
	}
	//if deployment has InRolloutProgressingAnnotation, it is under kruise-rollout control
	if _, ok := deployment.Annotations[InRolloutProgressingAnnotation]; !ok {
		return false
	}
	// only if deployment strategy is 'partition', webhook would add DeploymentStrategyAnnotation
	if _, ok := deployment.Annotations[DeploymentStrategyAnnotation]; ok {
		return false
	}
	return true
}
