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

package resourceclient

import (
	"context"

	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/server/selector"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func NewMockClient(storage rest.Storage) Client {
	return &mock{
		storage: storage,
	}
}

type mock struct {
	storage rest.Storage
}

func (r *mock) Get(ctx context.Context, key types.NamespacedName, u runtime.Unstructured, opts ...GetOption) error {
	o := GetOptions{}
	o.ApplyOptions(opts)

	obj, err := r.storage.Get(ctx, key.Name, &rest.GetOptions{
		Commit:            o.Commit,
		ShowManagedFields: o.ShowManagedFields,
		Trace:             o.Trace,
		Origin:            o.Origin,
	})
	if err != nil {
		return err
	}
	u.SetUnstructuredContent(obj.UnstructuredContent())
	return nil
}
func (r *mock) List(ctx context.Context, ul runtime.Unstructured, opts ...ListOption) error {
	o := ListOptions{}
	o.ApplyOptions(opts)

	selector, err := selector.ResourceExprSelectorAsSelector(o.ExprSelector)
	if err != nil {
		return err
	}

	obj, err := r.storage.List(ctx, &rest.ListOptions{
		Commit:            o.Commit,
		ShowManagedFields: o.ShowManagedFields,
		Trace:             o.Trace,
		Origin:            o.Origin,
		Selector:          selector,
		Watch:             false,
	})
	if err != nil {
		return err
	}
	ul.SetUnstructuredContent(obj.UnstructuredContent())
	return nil
}
func (r *mock) Apply(ctx context.Context, u runtime.Unstructured, opts ...ApplyOption) error {
	o := ApplyOptions{}
	o.ApplyOptions(opts)

	obj, err := r.storage.Apply(ctx, u, &rest.ApplyOptions{
		Trace:        o.Trace,
		Origin:       o.Origin,
		DryRun:       o.DryRun,
		FieldManager: o.FieldManager,
		Force:        o.Force,
	})
	if err != nil {
		return err
	}
	u.SetUnstructuredContent(obj.UnstructuredContent())
	return nil
}
func (r *mock) Create(ctx context.Context, u runtime.Unstructured, opts ...CreateOption) error {
	o := CreateOptions{}
	o.ApplyOptions(opts)

	obj, err := r.storage.Create(ctx, u, &rest.CreateOptions{
		Trace:  o.Trace,
		Origin: o.Origin,
		DryRun: o.DryRun,
	})
	if err != nil {
		return err
	}
	u.SetUnstructuredContent(obj.UnstructuredContent())
	return nil
}
func (r *mock) Update(ctx context.Context, u runtime.Unstructured, opts ...UpdateOption) error {
	o := UpdateOptions{}
	o.ApplyOptions(opts)

	obj, err := r.storage.Update(ctx, u, &rest.UpdateOptions{
		Trace:  o.Trace,
		Origin: o.Origin,
		DryRun: o.DryRun,
	})
	if err != nil {
		return err
	}
	u.SetUnstructuredContent(obj.UnstructuredContent())
	return nil
}
func (r *mock) Delete(ctx context.Context, u runtime.Unstructured, opts ...DeleteOption) error {
	o := DeleteOptions{}
	o.ApplyOptions(opts)

	obj := unstructured.Unstructured{
		Object: u.UnstructuredContent(),
	}
	_, err := r.storage.Delete(ctx, obj.GetName(), &rest.DeleteOptions{
		Trace:  o.Trace,
		Origin: o.Origin,
		DryRun: o.DryRun,
	})
	return err
}
func (r *mock) Watch(ctx context.Context, u runtime.Unstructured, opts ...ListOption) chan *resourcepb.Watch_Response {
	ch := make(chan *resourcepb.Watch_Response)
	return ch
}
func (r *mock) Close() error {
	return nil
}
