package statefulset

import (
	"context"

	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	"github.com/openkruise/kruise-tools/pkg/api"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Options struct {
	// CopyReplicas will mirror the current number of pods.
	CopyReplicas bool
}

// control implements creation.Control for StatefulSet→AdvancedStatefulSet.
type control struct {
	client client.Client
}

// NewControl returns a creator for AdvancedStatefulSet.
func NewControl(cfg *rest.Config) (*control, error) {
	scheme := api.GetScheme()
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	return &control{client: c}, nil
}

func int32Ptr(i int32) *int32 { return &i }

// Create builds and creates an AdvancedStatefulSet from a native StatefulSet.
func (c *control) Create(srcRef, dstRef api.ResourceRef, opts Options) error {
	var ss appsv1.StatefulSet
	if err := c.client.Get(context.Background(), srcRef.GetNamespacedName(), &ss); err != nil {
		return err
	}

	// convert native apps/v1 -> kruise/apps/v1beta1 update strategy
	var ru *kruiseappsv1beta1.RollingUpdateStatefulSetStrategy
	if ss.Spec.UpdateStrategy.RollingUpdate != nil {
		ru = &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
			Partition: ss.Spec.UpdateStrategy.RollingUpdate.Partition,
		}
	}

	ass := &kruiseappsv1beta1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dstRef.Name,
			Namespace: dstRef.Namespace,
			Labels:    ss.Labels,
		},
		Spec: kruiseappsv1beta1.StatefulSetSpec{
			ServiceName:          ss.Spec.ServiceName,
			Selector:             ss.Spec.Selector,
			Template:             ss.Spec.Template,
			VolumeClaimTemplates: ss.Spec.VolumeClaimTemplates,
			UpdateStrategy: kruiseappsv1beta1.StatefulSetUpdateStrategy{
				RollingUpdate: ru,
			},
			PodManagementPolicy:  ss.Spec.PodManagementPolicy,
			RevisionHistoryLimit: ss.Spec.RevisionHistoryLimit,
		},
	}

	// Handle replicas: default to 0, copy only if requested.
	if opts.CopyReplicas && ss.Spec.Replicas != nil {
		ass.Spec.Replicas = ss.Spec.Replicas
	} else {
		ass.Spec.Replicas = int32Ptr(0)
	}

	return c.client.Create(context.Background(), ass)
}
