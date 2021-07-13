package fetcher

import (
	"context"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetResourceInCache(ns, name string, obj runtime.Object, cl client.Reader) (bool, error) {
	err := cl.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: name}, obj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func ListResourceInCache(ns string, obj runtime.Object, cl client.Client) error {
	return cl.List(context.TODO(), obj, client.InNamespace(ns))
}

func GetCloneSetInCache(ns, name string, cl client.Reader) (*kruiseappsv1alpha1.CloneSet, bool, error) {
	cs := &kruiseappsv1alpha1.CloneSet{}
	found, err := GetResourceInCache(ns, name, cs, cl)
	if err != nil || !found {
		cs = nil
	}

	return cs, found, err
}
