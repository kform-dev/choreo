/*
Copyright 2024 Nokia.

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

package gotemplate

import (
	"bytes"
	"context"
	"fmt"

	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/controller/reconcile"
	"github.com/kform-dev/choreo/pkg/controller/reconciler/gotemplate/parser"
	"github.com/kform-dev/choreo/pkg/controller/reconciler/resources"
	"github.com/kform-dev/choreo/pkg/proto/grpcerrors"
	"github.com/kform-dev/choreo/pkg/util/object"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

func NewReconcilerFn(client resourceclient.Client, reconcileConfig *choreov1alpha1.Reconciler, branch string) reconcile.TypedReconcilerFn {
	return func() (reconcile.TypedReconciler, error) {
		r := &reconciler{
			name:   reconcileConfig.Name,
			client: client,
			forgvk: reconcileConfig.GetForGVK(),
			owns:   reconcileConfig.GetOwnsGVKs(),
			branch: branch,
		}

		p, err := parser.New(reconcileConfig.Spec.Code)
		if err != nil {
			return nil, err
		}

		r.parser = p
		return r, nil
	}
}

type reconciler struct {
	name   string
	parser *parser.Parser
	client resourceclient.Client
	forgvk schema.GroupVersionKind
	owns   sets.Set[schema.GroupVersionKind]
	branch string
	// dynamic data set on each reconcile
	ctx       context.Context // bad practice but allows for reuse of ctx
	resources *resources.Resources
}

func (r *reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	//log := log.FromContext(ctx)
	// bad practice but allows to reuse the context
	r.ctx = ctx
	// get the resource
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(r.forgvk)
	if err := r.client.Get(r.ctx, req.NamespacedName, u, &resourceclient.GetOptions{
		ShowManagedFields: true,
		Origin:            r.name,
		Branch:            r.branch,
	}); err != nil {
		if !grpcerrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		// stop the reconcile loop since the object dissapeared
		return reconcile.Result{}, nil
	}

	// reinitialize the resource on each reconcile
	r.resources = resources.New(r.name, r.client, u, r.owns, r.branch)
	if u.GetDeletionTimestamp() != nil {
		if err := r.resources.Delete(ctx); err != nil {
			return reconcile.Result{}, fmt.Errorf("gotemplate reconciler %s cannot delete child resource, err: %s", r.name, err.Error())
		}
		object.DeleteFinalizer(u, r.name)

		// removes the fields that are not managed by this reconciler based on the managedFields info in the resource
		// done before conditions are set
		object.PruneUnmanagedFields(u, r.name)

		if err := r.client.Apply(ctx, u, &resourceclient.ApplyOptions{
			FieldManager: r.name,
			Branch:       r.branch,
		}); err != nil {
			return reconcile.Result{}, fmt.Errorf("gotemplate reconciler %s cannot set finalizer, err: %s", r.name, err.Error())
		}
		return reconcile.Result{}, nil
	}

	object.SetFinalizer(u, r.name)

	var buf bytes.Buffer
	if err := r.parser.Render(ctx, "main.tpl", u, &buf); err != nil {
		return reconcile.Result{}, fmt.Errorf("gotemplate reconciler %s parser failed err: %s", r.name, err.Error())
	}

	data := map[string]any{}
	if err := yaml.Unmarshal(buf.Bytes(), &data); err != nil {
		return reconcile.Result{}, fmt.Errorf("gotemplate reconciler %s generate invalid yaml failed err: %s", r.name, err.Error())
	}

	// apply the own resource generated by the template
	newu := &unstructured.Unstructured{}
	newu.Object = data
	r.resources.AddNewResource(ctx, newu)
	if err := r.resources.Apply(ctx); err != nil {
		return reconcile.Result{}, fmt.Errorf("apply failed for gotemplate reconciler %s, err: %s", r.name, err.Error())
	}

	// removes the fields that are not managed by this reconciler based on the managedFields info in the resource
	// done before conditions are set
	object.PruneUnmanagedFields(u, r.name)
	// apply the for resource
	if err := r.client.Apply(ctx, u, &resourceclient.ApplyOptions{
		FieldManager: r.name,
		Branch:       r.branch,
	}); err != nil {
		return reconcile.Result{}, fmt.Errorf("go template reconciler %s cannot set finalizer, err: %s", r.name, err.Error())
	}
	return reconcile.Result{}, nil
}
