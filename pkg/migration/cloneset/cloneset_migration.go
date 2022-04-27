/*
Copyright 2020 The Kruise Authors.

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

package cloneset

import (
	"context"
	"fmt"
	"sync"
	"time"

	appsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	"github.com/openkruise/kruise-tools/pkg/api"
	"github.com/openkruise/kruise-tools/pkg/migration"
	"github.com/openkruise/kruise-tools/pkg/utils"

	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	// TODO: make it as an option
	maxConcurrentReconciles = 5
)

type control struct {
	client   client.Client
	cache    cache.Cache
	queue    workqueue.RateLimitingInterface
	stopChan <-chan struct{}

	sync.RWMutex
	tasks          map[types.UID]*task
	executingTasks map[api.ResourceRef]*task
	handledGVKs    map[schema.GroupVersionKind]struct{}
}

type task struct {
	ID                types.UID
	creationTimestamp metav1.Time

	src  api.ResourceRef
	dst  api.ResourceRef
	opts migration.Options

	srcUpdatedGeneration int64
	dstUpdatedGeneration int64

	mu     sync.Mutex
	result migration.Result
}

var _ migration.Control = &control{}

func NewControl(cfg *rest.Config, stopChan <-chan struct{}) (migration.Control, error) {
	scheme := api.GetScheme()
	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		return nil, err
	}

	ctrl := &control{
		stopChan:       stopChan,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cloneset-migration-control"),
		tasks:          make(map[types.UID]*task),
		executingTasks: make(map[api.ResourceRef]*task),
		handledGVKs:    make(map[schema.GroupVersionKind]struct{}),
	}

	if ctrl.client, err = client.New(cfg, client.Options{Scheme: scheme, Mapper: mapper}); err != nil {
		return nil, err
	}

	if ctrl.cache, err = cache.New(cfg, cache.Options{Scheme: scheme, Mapper: mapper}); err != nil {
		return nil, err
	}

	go func() {
		_ = ctrl.cache.Start(context.TODO())
	}()
	// Wait for the caches to sync.
	ctrl.cache.WaitForCacheSync(context.TODO())

	for i := 0; i < maxConcurrentReconciles; i++ {
		// Process work items
		go wait.Until(ctrl.worker, time.Second, stopChan)
	}

	return ctrl, nil
}

func (c *control) Submit(src api.ResourceRef, dst api.ResourceRef, opts migration.Options) (migration.Result, error) {
	if opts.Replicas != nil && *opts.Replicas <= 0 {
		return migration.Result{}, fmt.Errorf("invalid replicas %v", *opts.Replicas)
	} else if src.GetGroupVersionKind() != api.DeploymentKind {
		return migration.Result{}, fmt.Errorf("invalid src type, currently only support %v", api.DeploymentKind.String())
	} else if dst.GetGroupVersionKind() != api.CloneSetKind {
		return migration.Result{}, fmt.Errorf("invalid dst type, must be %v", api.CloneSetKind.String())
	}

	srcGVK := src.GetGroupVersionKind()
	dstGVK := dst.GetGroupVersionKind()

	srcDeployment, dstCloneSet, err := getDeploymentAndCloneSetObjects(c.client, &src, &dst)
	if err != nil {
		return migration.Result{}, err
	}

	if opts.Replicas == nil {
		opts.Replicas = srcDeployment.Spec.Replicas
	}
	if opts.MaxSurge == nil {
		opts.MaxSurge = func() *int32 { var i int32 = 1; return &i }()
	}
	if *opts.MaxSurge <= 0 {
		return migration.Result{}, fmt.Errorf("maxSurge must be integar more than zore")
	}

	c.Lock()
	defer c.Unlock()
	if _, ok := c.executingTasks[src]; ok {
		return migration.Result{}, fmt.Errorf("already existing migration task for %v", src)
	} else if _, ok := c.executingTasks[dst]; ok {
		return migration.Result{}, fmt.Errorf("already existing migration task for %v", dst)
	}

	if err := c.addEventHandler(srcGVK); err != nil {
		return migration.Result{}, err
	}
	if err := c.addEventHandler(dstGVK); err != nil {
		return migration.Result{}, err
	}

	id := uuid.NewUUID()
	t := task{
		ID:                id,
		creationTimestamp: metav1.Now(),

		src:  src,
		dst:  dst,
		opts: opts,

		srcUpdatedGeneration: srcDeployment.Generation,
		dstUpdatedGeneration: dstCloneSet.Generation,

		result: migration.Result{ID: id, State: migration.MigrateExecuting},
	}
	c.tasks[t.ID] = &t
	c.executingTasks[t.src] = &t
	c.executingTasks[t.dst] = &t

	// must enqueue once
	c.queue.Add(id)

	return t.result, nil
}

func (c *control) Query(ID types.UID) (migration.Result, error) {
	t := c.getTask(ID)
	if t == nil {
		return migration.Result{}, fmt.Errorf("not found ID %v", ID)
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	return t.result, nil
}

func (c *control) addEventHandler(gvk schema.GroupVersionKind) error {
	if _, ok := c.handledGVKs[gvk]; !ok {
		informer, err := c.cache.GetInformerForKind(context.Background(), gvk)
		if err != nil {
			return fmt.Errorf("failed to get informer for %v: %v", gvk, err)
		}
		if gvk == api.DeploymentKind {
			informer.AddEventHandler(&deploymentHandler{ctrl: c})
		} else if gvk == api.CloneSetKind {
			informer.AddEventHandler(&cloneSetHandler{ctrl: c})
		} else {
			return fmt.Errorf("unsupported gvk %v", gvk)
		}
		c.handledGVKs[gvk] = struct{}{}
	}
	return nil
}

func (c *control) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *control) processNextWorkItem() bool {
	obj, shutdown := c.queue.Get()
	if shutdown {
		// Stop working
		return false
	}
	defer c.queue.Done(obj)

	err := c.reconcile(obj.(types.UID))
	if err == nil {
		c.queue.Forget(obj)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("sync %q failed with %v", obj, err))
	c.queue.AddRateLimited(obj)

	return true
}

func (c *control) reconcile(ID types.UID) error {
	task := c.getTask(ID)
	if task.result.State != migration.MigrateExecuting {
		return nil
	} else if task.result.DstMigratedReplicas == *task.opts.Replicas && task.result.SrcMigratedReplicas == *task.opts.Replicas {
		c.finishTask(task, migration.MigrateSucceeded, "")
		return nil
	} else if task.opts.TimeoutSeconds != nil && time.Since(task.creationTimestamp.Time) > time.Duration(*task.opts.TimeoutSeconds)*time.Second {
		c.finishTask(task, migration.MigrateFailed, fmt.Sprintf("task timeout exceeded"))
		return nil
	}

	srcDeployment, dstCloneSet, err := getDeploymentAndCloneSetObjects(c.cache, &task.src, &task.dst)
	if err != nil {
		c.finishTask(task, migration.MigrateFailed, err.Error())
		return nil
	}

	if srcDeployment.Generation < task.srcUpdatedGeneration || dstCloneSet.Generation < task.dstUpdatedGeneration {
		// cache has not synced
		return nil
	} else if srcDeployment.Generation != srcDeployment.Status.ObservedGeneration || dstCloneSet.Generation != dstCloneSet.Status.ObservedGeneration {
		// workload controller has not reconciled
		return nil
	}

	// dst need scale out
	if task.result.DstMigratedReplicas < *task.opts.Replicas {
		deltaSurge := *task.opts.MaxSurge - (task.result.DstMigratedReplicas - task.result.SrcMigratedReplicas)
		deltaReplicas := *task.opts.Replicas - task.result.DstMigratedReplicas
		maxScaleOut := utils.Int32Min(deltaSurge, deltaReplicas)

		if maxScaleOut > 0 {
			*dstCloneSet.Spec.Replicas += maxScaleOut
			if err := c.client.Update(context.TODO(), dstCloneSet); err != nil {
				return err
			}
			task.dstUpdatedGeneration = dstCloneSet.Generation
			c.updateTask(task, 0, maxScaleOut)
			return nil
		}
	}

	// src need scale in
	if task.result.SrcMigratedReplicas < *task.opts.Replicas {
		deltaReplicas := *task.opts.Replicas - task.result.SrcMigratedReplicas
		deltaMigrated := task.result.DstMigratedReplicas - task.result.SrcMigratedReplicas
		maxScaleIn := utils.Int32Min(*srcDeployment.Spec.Replicas, deltaReplicas, deltaMigrated)

		// must wait for all pods in CloneSet available
		if maxScaleIn > 0 && *dstCloneSet.Spec.Replicas == dstCloneSet.Status.AvailableReplicas {
			*srcDeployment.Spec.Replicas -= maxScaleIn
			if err := c.client.Update(context.TODO(), srcDeployment); err != nil {
				return err
			}
			task.srcUpdatedGeneration = srcDeployment.Generation
			c.updateTask(task, maxScaleIn, 0)
			return nil
		}
	}

	return nil
}

func (c *control) getTask(ID types.UID) *task {
	c.RLock()
	defer c.RUnlock()
	return c.tasks[ID]
}

func (c *control) updateTask(t *task, srcMigratedReplicas, dstMigratedReplicas int32) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.result.SrcMigratedReplicas += srcMigratedReplicas
	t.result.DstMigratedReplicas += dstMigratedReplicas
}

func (c *control) finishTask(t *task, state migration.MigrateState, message string) {
	func() {
		t.mu.Lock()
		defer t.mu.Unlock()
		t.result.State = state
		t.result.Message = message
	}()

	c.Lock()
	defer c.Unlock()
	delete(c.executingTasks, t.src)
	delete(c.executingTasks, t.dst)
}

func getDeploymentAndCloneSetObjects(reader client.Reader, src, dst *api.ResourceRef) (*apps.Deployment, *appsv1alpha1.CloneSet, error) {
	srcDeployment := apps.Deployment{}
	if err := reader.Get(context.TODO(), src.GetNamespacedName(), &srcDeployment); err != nil {
		return nil, nil, fmt.Errorf("failed to get %v: %v", src, err)
	}

	dstCloneSet := appsv1alpha1.CloneSet{}
	if err := reader.Get(context.TODO(), dst.GetNamespacedName(), &dstCloneSet); err != nil {
		return nil, nil, fmt.Errorf("failed to get %v: %v", dst, err)
	}

	return &srcDeployment, &dstCloneSet, nil
}
