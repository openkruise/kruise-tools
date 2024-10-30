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

	appsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	"github.com/openkruise/kruise-tools/pkg/api"
	"github.com/openkruise/kruise-tools/pkg/conversion"
	"github.com/openkruise/kruise-tools/pkg/creation"

	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type control struct {
	client client.Client
}

func NewControl(cfg *rest.Config) (creation.Control, error) {
	scheme := api.GetScheme()
	c, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}
	mapper, err := apiutil.NewDynamicRESTMapper(cfg, c)
	if err != nil {
		return nil, err
	}

	ctrl := &control{}
	if ctrl.client, err = client.New(cfg, client.Options{Scheme: scheme, Mapper: mapper}); err != nil {
		return nil, err
	}

	return ctrl, nil
}

func (c *control) Create(src api.ResourceRef, dst api.ResourceRef, opts creation.Options) error {
	if src.GetGroupVersionKind() != api.DeploymentKind {
		return fmt.Errorf("invalid src type, currently only support %v", api.DeploymentKind.String())
	} else if dst.GetGroupVersionKind() != api.CloneSetKind {
		return fmt.Errorf("invalid dst type, must be %v", api.CloneSetKind.String())
	}

	if err := c.ensureCloneSetNotExists(dst); err != nil {
		return err
	}
	srcDeployment, err := c.getDeployment(src)
	if err != nil {
		return err
	}

	dstCloneSet := conversion.DeploymentToCloneSet(srcDeployment, dst.Name)
	return c.client.Create(context.TODO(), dstCloneSet)
}

func (c *control) getDeployment(ref api.ResourceRef) (*apps.Deployment, error) {
	d := &apps.Deployment{}
	if err := c.client.Get(context.TODO(), ref.GetNamespacedName(), d); err != nil {
		return nil, fmt.Errorf("failed to get %v: %v", ref, err)
	}
	return d, nil
}

func (c *control) ensureCloneSetNotExists(ref api.ResourceRef) error {
	cs := &appsv1alpha1.CloneSet{}
	if err := c.client.Get(context.TODO(), ref.GetNamespacedName(), cs); err == nil {
		return fmt.Errorf("cloneset %v already exists", ref.GetNamespacedName())
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get %v: %v", ref, err)
	}
	return nil
}
