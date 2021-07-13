package fetcher

import (
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetAdvancedStsInCache(ns, name string, cl client.Reader) (*kruiseappsv1beta1.StatefulSet, bool, error) {
	asts := &kruiseappsv1beta1.StatefulSet{}
	found, err := GetResourceInCache(ns, name, asts, cl)
	if err != nil || !found {
		asts = nil
	}
	return asts, found, err
}
