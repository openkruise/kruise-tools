package statefulset

import (
	"context"
	"fmt"
	"sync"
	"time"

	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	"github.com/openkruise/kruise-tools/pkg/api"
	"github.com/openkruise/kruise-tools/pkg/migration"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// control orchestrates StatefulSet→AdvancedStatefulSet.
type control struct {
	client client.Client
	cache  cache.Cache
	queue  workqueue.RateLimitingInterface
	stop   <-chan struct{}

	mu    sync.RWMutex
	tasks map[types.UID]*task
}

type task struct {
	ID       types.UID
	start    time.Time
	src, dst api.ResourceRef
	result   migration.Result
	mu       sync.Mutex
}

// NewControl sets up the migration loop.
func NewControl(cfg *rest.Config, stopChan <-chan struct{}) (migration.Control, error) {
	scheme := api.GetScheme()
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	cache, err := cache.New(cfg, cache.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	ctrl := &control{
		client: c,
		cache:  cache,
		queue:  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "statefulset-migration"),
		stop:   stopChan,
		tasks:  make(map[types.UID]*task),
	}
	go cache.Start(context.Background())
	cache.WaitForCacheSync(context.Background())
	return ctrl, nil
}

// Submit creates the AdvancedStatefulSet then orphan-deletes the native StatefulSet.
func (c *control) Submit(src api.ResourceRef, dst api.ResourceRef, _ migration.Options) (migration.Result, error) {
	var ss appsv1.StatefulSet
	if err := c.client.Get(context.Background(), src.GetNamespacedName(), &ss); err != nil {
		return migration.Result{}, fmt.Errorf("get native StatefulSet: %w", err)
	}

	// build AdvancedStatefulSet object including all crucial fields
	ass := &kruiseappsv1beta1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dst.Name,
			Namespace: dst.Namespace,
			Labels:    ss.Labels,
		},
		Spec: kruiseappsv1beta1.StatefulSetSpec{
			Replicas:             ss.Spec.Replicas,
			ServiceName:          ss.Spec.ServiceName,
			Selector:             ss.Spec.Selector,
			Template:             ss.Spec.Template,
			VolumeClaimTemplates: ss.Spec.VolumeClaimTemplates,
			UpdateStrategy:       ss.Spec.UpdateStrategy,
			PodManagementPolicy:  ss.Spec.PodManagementPolicy,
			RevisionHistoryLimit: ss.Spec.RevisionHistoryLimit,
		},
	}
	if err := c.client.Create(context.Background(), ass); err != nil {
		return migration.Result{}, fmt.Errorf("create ASS: %w", err)
	}

	prop := metav1.DeletePropagationOrphan
	if err := c.client.Delete(context.Background(), &ss,
		&client.DeleteOptions{PropagationPolicy: &prop}); err != nil {
		return migration.Result{}, fmt.Errorf("orphan-delete SS: %w", err)
	}

	id := types.UID(uuid.NewUUID())
	result := migration.Result{ID: id, State: migration.MigrateSucceeded}
	c.mu.Lock()
	c.tasks[id] = &task{ID: id, start: time.Now(), src: src, dst: dst, result: result}
	c.mu.Unlock()
	return result, nil
}

// Query returns the stored result.
func (c *control) Query(id types.UID) (migration.Result, error) {
	c.mu.RLock()
	t, ok := c.tasks[id]
	c.mu.RUnlock()
	if !ok {
		return migration.Result{}, fmt.Errorf("task %s not found", id)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.result, nil
}
