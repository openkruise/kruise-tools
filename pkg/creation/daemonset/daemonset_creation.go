package daemonset

import (
	"context"

	appsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openkruise/kruise-tools/pkg/api"
)

// Options for creation
type Options struct {
	// CopyReplicas will mirror the current number of pods.
	CopyReplicas bool
}

// control implements creation.Control for DaemonSet→AdvancedDaemonSet.
type control struct {
	client client.Client
}

// NewControl returns a creator for AdvancedDaemonSet.
func NewControl(cfg *rest.Config) (*control, error) {
	scheme := api.GetScheme()
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	return &control{client: c}, nil
}

// Create builds and creates an AdvancedDaemonSet from a native DaemonSet.
func (c *control) Create(srcRef, dstRef api.ResourceRef, opts Options) error {
	// fetch native DaemonSet
	var ds appsv1.DaemonSet
	if err := c.client.Get(context.Background(), srcRef.GetNamespacedName(), &ds); err != nil {
		return err
	}

	// build ADS object
	ads := &appsv1alpha1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dstRef.Name,
			Namespace: dstRef.Namespace,
			Labels:    ds.Labels,
		},
		Spec: appsv1alpha1.DaemonSetSpec{
			Selector: ds.Spec.Selector,
			Template: ds.Spec.Template,
			UpdateStrategy: appsv1alpha1.DaemonSetUpdateStrategy{
				Type: appsv1alpha1.DaemonSetUpdateStrategyType(ds.Spec.UpdateStrategy.Type),
			},
		},
	}
	// note: DaemonSets don’t use a “replicas” field—CopyReplicas is informational

	return c.client.Create(context.Background(), ads)
}
