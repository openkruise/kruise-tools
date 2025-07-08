package daemonset

import (
	"context"
	"fmt"
	"sync"
	"time"

	appsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openkruise/kruise-tools/pkg/api"
	"github.com/openkruise/kruise-tools/pkg/migration"
)

// control orchestrates DaemonSet→AdvancedDaemonSet.
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
		queue:  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "daemonset-migration"),
		stop:   stopChan,
		tasks:  make(map[types.UID]*task),
	}
	go cache.Start(context.Background())
	cache.WaitForCacheSync(context.Background())
	return ctrl, nil
}

// Submit creates the ADS then orphan-deletes the native DS.
func (c *control) Submit(src api.ResourceRef, dst api.ResourceRef, _ migration.Options) (migration.Result, error) {
	var ds appsv1.DaemonSet
	if err := c.client.Get(context.Background(), src.GetNamespacedName(), &ds); err != nil {
		return migration.Result{}, fmt.Errorf("get native DaemonSet: %w", err)
	}

	ads := &appsv1alpha1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dst.Name,
			Namespace: dst.Namespace,
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
	if err := c.client.Create(context.Background(), ads); err != nil {
		return migration.Result{}, fmt.Errorf("create ADS: %w", err)
	}

	prop := metav1.DeletePropagationOrphan
	if err := c.client.Delete(context.Background(), &ds,
		&client.DeleteOptions{PropagationPolicy: &prop}); err != nil {
		return migration.Result{}, fmt.Errorf("orphan-delete DS: %w", err)
	}

	id := types.UID(uuid.NewUUID()) // nolint:unconvert
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
