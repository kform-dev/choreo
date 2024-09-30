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

package api

import (
	"context"

	"github.com/henderiw/store"
	"github.com/henderiw/store/memory"
	"github.com/henderiw/store/watch"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

// NewAPIResourceStore stores
// gvk -> multiple storages
// embedded API(s) are stored twice
// internal/dynamic API(s) can only be stored once
func NewAPIStore() *APIStore {
	return &APIStore{
		memory.NewStore[*ResourceContext](func() *ResourceContext { return &ResourceContext{} }),
	}
}

type ResourceContext struct {
	*discoverypb.APIResource
	Storage rest.Storage
	CRD     *apiextensionsv1.CustomResourceDefinition
}

func (r *ResourceContext) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   r.Group,
		Version: r.Version,
		Kind:    r.Kind,
	}
}

// key 1 is gvk.String; key 2 is the name of the choreoinst
type APIStore struct {
	store.Storer[*ResourceContext]
}

func (r APIStore) Apply(gvk schema.GroupVersionKind, rctx *ResourceContext) error {
	gvkKey := store.ToKey(gvk.String())
	return r.Storer.Apply(gvkKey, rctx)
}

func (r APIStore) Delete(gvk schema.GroupVersionKind) error {
	gvkKey := store.ToKey(gvk.String())
	return r.Storer.Delete(gvkKey)
}

func (r *APIStore) GetAPIResources() []*discoverypb.APIResource {
	apiResources := []*discoverypb.APIResource{}
	r.List(func(k store.Key, rctx *ResourceContext) {
		apiResources = append(apiResources, rctx.APIResource)
	})
	return apiResources
}

func (r *APIStore) Has(gvk schema.GroupVersionKind) bool {
	gvkKey := store.ToKey(gvk.String())
	_, err := r.Storer.Get(gvkKey)
	return err == nil
}

func (r *APIStore) Get(gvk schema.GroupVersionKind) (*ResourceContext, error) {
	gvkKey := store.ToKey(gvk.String())
	return r.Storer.Get(gvkKey)
}

func (r *APIStore) GetStorage(gvk schema.GroupVersionKind) (rest.Storage, error) {
	gvkKey := store.ToKey(gvk.String())
	resctx, err := r.Storer.Get(gvkKey)
	if err != nil {
		return nil, err
	}
	return resctx.Storage, nil
}

func (r *APIStore) GetGVKSet() sets.Set[schema.GroupVersionKind] {
	gvkset := sets.New[schema.GroupVersionKind]()
	r.List(func(k store.Key, rc *ResourceContext) {
		gvkset.Insert(rc.GVK())
	})
	return gvkset
}

func (r *APIStore) Import(other *APIStore) {
	other.Storer.List(func(k store.Key, rc *ResourceContext) {
		r.Storer.Apply(k, rc)
	})
}

func (r *APIStore) Watch(ctx context.Context, opts ...store.ListOption) (watch.WatchInterface[*ResourceContext], error) {
	return r.Storer.Watch(ctx, opts...)
}
