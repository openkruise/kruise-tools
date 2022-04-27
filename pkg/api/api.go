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

package api

import (
	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	kruiserolloutsv1apha1 "github.com/openkruise/rollouts/api/v1alpha1"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubectl/pkg/scheme"
)

var (
	DeploymentKind = apps.SchemeGroupVersion.WithKind("Deployment")
	CloneSetKind   = kruiseappsv1alpha1.SchemeGroupVersion.WithKind("CloneSet")
)

var Scheme = scheme.Scheme

func init() {
	_ = clientgoscheme.AddToScheme(Scheme)
	_ = kruiseappsv1alpha1.AddToScheme(Scheme)
	_ = kruiseappsv1beta1.AddToScheme(Scheme)
	_ = kruiserolloutsv1apha1.AddToScheme(Scheme)
}

func GetScheme() *runtime.Scheme {
	return Scheme
}

type ResourceRef struct {
	// API version of the object.
	APIVersion string
	// Kind of the object.
	Kind string
	// Namespace of the object.
	Namespace string
	// Name of the object.
	Name string
}

func (rf *ResourceRef) GetGroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(rf.APIVersion, rf.Kind)
}

func (rf *ResourceRef) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{Namespace: rf.Namespace, Name: rf.Name}
}

func NewDeploymentRef(namespace, name string) ResourceRef {
	return ResourceRef{
		APIVersion: DeploymentKind.GroupVersion().String(),
		Kind:       DeploymentKind.Kind,
		Namespace:  namespace,
		Name:       name,
	}
}

func NewCloneSetRef(namespace, name string) ResourceRef {
	return ResourceRef{
		APIVersion: CloneSetKind.GroupVersion().String(),
		Kind:       CloneSetKind.Kind,
		Namespace:  namespace,
		Name:       name,
	}
}
