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
	appsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	"github.com/openkruise/kruise-tools/pkg/api"

	apps "k8s.io/api/apps/v1"
	toolscache "k8s.io/client-go/tools/cache"
)

type cloneSetHandler struct {
	ctrl *control
}

var _ toolscache.ResourceEventHandler = &cloneSetHandler{}

func (ch *cloneSetHandler) OnAdd(obj interface{}, isInInitialList bool) {
	d, ok := obj.(*appsv1alpha1.CloneSet)
	if !ok {
		return
	}
	ref := api.ResourceRef{
		APIVersion: api.CloneSetKind.GroupVersion().String(),
		Kind:       api.CloneSetKind.Kind,
		Namespace:  d.Namespace,
		Name:       d.Name,
	}

	ch.ctrl.RLock()
	defer ch.ctrl.RUnlock()
	if task, ok := ch.ctrl.executingTasks[ref]; ok {
		ch.ctrl.queue.Add(task.ID)
	}
}

func (ch *cloneSetHandler) OnUpdate(oldObj interface{}, newObj interface{}) {
	d, ok := newObj.(*appsv1alpha1.CloneSet)
	if !ok {
		return
	}

	ref := api.ResourceRef{
		APIVersion: api.CloneSetKind.GroupVersion().String(),
		Kind:       api.CloneSetKind.Kind,
		Namespace:  d.Namespace,
		Name:       d.Name,
	}

	ch.ctrl.RLock()
	defer ch.ctrl.RUnlock()
	if task, ok := ch.ctrl.executingTasks[ref]; ok {
		ch.ctrl.queue.Add(task.ID)
	}
}

func (ch *cloneSetHandler) OnDelete(obj interface{}) {
	d, ok := obj.(*appsv1alpha1.CloneSet)
	if !ok {
		return
	}
	ref := api.ResourceRef{
		APIVersion: api.CloneSetKind.GroupVersion().String(),
		Kind:       api.CloneSetKind.Kind,
		Namespace:  d.Namespace,
		Name:       d.Name,
	}

	ch.ctrl.RLock()
	defer ch.ctrl.RUnlock()
	if task, ok := ch.ctrl.executingTasks[ref]; ok {
		ch.ctrl.queue.Add(task.ID)
	}
}

type deploymentHandler struct {
	ctrl *control
}

var _ toolscache.ResourceEventHandler = &deploymentHandler{}

func (dh *deploymentHandler) OnAdd(obj interface{}, isInInitialList bool) {
	d, ok := obj.(*apps.Deployment)
	if !ok {
		return
	}
	ref := api.ResourceRef{
		APIVersion: api.DeploymentKind.GroupVersion().String(),
		Kind:       api.DeploymentKind.Kind,
		Namespace:  d.Namespace,
		Name:       d.Name,
	}

	dh.ctrl.RLock()
	defer dh.ctrl.RUnlock()
	if task, ok := dh.ctrl.executingTasks[ref]; ok {
		dh.ctrl.queue.Add(task.ID)
	}
}

func (dh *deploymentHandler) OnUpdate(oldObj interface{}, newObj interface{}) {
	d, ok := newObj.(*apps.Deployment)
	if !ok {
		return
	}

	ref := api.ResourceRef{
		APIVersion: api.DeploymentKind.GroupVersion().String(),
		Kind:       api.DeploymentKind.Kind,
		Namespace:  d.Namespace,
		Name:       d.Name,
	}

	dh.ctrl.RLock()
	defer dh.ctrl.RUnlock()
	if task, ok := dh.ctrl.executingTasks[ref]; ok {
		dh.ctrl.queue.Add(task.ID)
	}
}

func (dh *deploymentHandler) OnDelete(obj interface{}) {
	d, ok := obj.(*apps.Deployment)
	if !ok {
		return
	}
	ref := api.ResourceRef{
		APIVersion: api.DeploymentKind.GroupVersion().String(),
		Kind:       api.DeploymentKind.Kind,
		Namespace:  d.Namespace,
		Name:       d.Name,
	}

	dh.ctrl.RLock()
	defer dh.ctrl.RUnlock()
	if task, ok := dh.ctrl.executingTasks[ref]; ok {
		dh.ctrl.queue.Add(task.ID)
	}
}
