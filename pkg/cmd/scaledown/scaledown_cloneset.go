/*
Copyright 2021 The Kruise Authors.

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

package scaledown

import (
	"errors"
	"fmt"
	"strings"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	"k8s.io/cli-runtime/pkg/resource"
)

func (o *ScaleDownOptions) ScaleDownCloneSet(info *resource.Info) error {
	cloneSetName := info.Name
	obj, err := resource.
		NewHelper(info.Client, info.Mapping).
		Get(info.Namespace, info.Name)
	if err != nil {
		return err
	}
	res := obj.(*kruiseappsv1alpha1.CloneSet)

	podsSlc := strings.Split(o.Pods, ",")
	afterReplicas := *res.Spec.Replicas - int32(len(podsSlc))
	res.Spec.ScaleStrategy.PodsToDelete = append(res.Spec.ScaleStrategy.PodsToDelete, podsSlc...)
	res.Spec.Replicas = &afterReplicas

	_, err = resource.
		NewHelper(info.Client, info.Mapping).
		Replace(info.Namespace, info.Name, true, res)
	if err != nil {
		fmt.Fprintf(o.Out, "%s delete pods %s failed\n", cloneSetName, podsSlc)
		return fmt.Errorf("scaledown cloneset %s failed, error is %v", res.Name, err)
	}

	fmt.Fprintf(o.Out, "# %s delete pods %s successfully\n", cloneSetName, podsSlc)
	if err := o.PrintObj(res, o.Out); err != nil {
		return errors.New(err.Error())
	}

	return nil
}
