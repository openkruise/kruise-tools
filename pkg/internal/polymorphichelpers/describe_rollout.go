package polymorphichelpers

import (
	rolloutschema "github.com/openkruise/kruise-rollout-api/rollouts/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type RolloutViewer func(obj runtime.Unstructured) (*rolloutschema.Rollout, error)
