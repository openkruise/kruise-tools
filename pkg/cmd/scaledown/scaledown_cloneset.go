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
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/openkruise/kruise-tools/pkg/fetcher"
	"k8s.io/klog"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (o *ScaleDownOptions) ScaleDownCloneSet(f cmdutil.Factory, cloneSetName string, cr client.Reader, cl client.Client) error {
	res, found, err := fetcher.GetCloneSetInCache(o.Namespace, cloneSetName, cr)
	if err != nil || !found {
		klog.Error(err)
		return fmt.Errorf("failed to retrieve CloneSet %s: %s", cloneSetName, err.Error())
	}

	podsSlc := strings.Split(o.Pods, ",")
	afterReplicas := *res.Spec.Replicas - int32(len(podsSlc))
	res.Spec.ScaleStrategy.PodsToDelete = append(res.Spec.ScaleStrategy.PodsToDelete, podsSlc...)
	res.Spec.Replicas = &afterReplicas

	err = cl.Update(context.TODO(), res)
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
