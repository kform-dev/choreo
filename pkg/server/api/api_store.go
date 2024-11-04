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
	"sort"

	"github.com/henderiw/store"
	"github.com/henderiw/store/memory"
	"github.com/henderiw/store/watch"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

func NewAPIStore() *APIStore {
	return &APIStore{
		memory.NewStore(func() *ResourceContext { return &ResourceContext{} }),
	}
}

type ResourceContext struct {
	Internal *discoverypb.APIResource
	External *discoverypb.APIResource
	Storage  rest.Storage
	CRD      *apiextensionsv1.CustomResourceDefinition
}

func (r *ResourceContext) GV() schema.GroupKind {
	return schema.GroupKind{
		Group: r.External.Group,
		Kind:  r.External.Kind,
	}
}

// ExternalGVK return the gvk as seen by the client
func (r *ResourceContext) ExternalGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   r.External.Group,
		Version: r.External.Version,
		Kind:    r.External.Kind,
	}
}

// InternalGVK return the gvk as seen by the storage layer
func (r *ResourceContext) InternalGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   r.Internal.Group,
		Version: r.Internal.Version,
		Kind:    r.Internal.Kind,
	}
}

// key 1 is gvk.String; key 2 is the name of the choreoinst
type APIStore struct {
	store.Storer[*ResourceContext]
}

func (r APIStore) Apply(gk schema.GroupKind, rctx *ResourceContext) error {
	gkKey := store.ToKey(gk.String())
	return r.Storer.Apply(gkKey, rctx)
}

func (r APIStore) Delete(gk schema.GroupKind) error {
	gkKey := store.ToKey(gk.String())
	return r.Storer.Delete(gkKey)
}

func (r *APIStore) GetAPIResources() []*discoverypb.APIResource {
	apiResources := []*discoverypb.APIResource{}
	r.List(func(k store.Key, rctx *ResourceContext) {
		apiResources = append(apiResources, rctx.External)
	})

	sort.Slice(apiResources, func(i, j int) bool {
		return apiResources[i].Resource < apiResources[j].Resource
	})

	return apiResources
}

func (r *APIStore) Has(gvk schema.GroupVersionKind) bool {
	gkKey := store.ToKey(gvk.GroupKind().String())
	rctx, err := r.Storer.Get(gkKey)
	if err != nil {
		return false
	}
	if rctx.External.Kind == gvk.Kind {
		return true
	}
	if rctx.Internal.Kind == gvk.Kind {
		return true
	}
	return false
}

func (r *APIStore) Get(gk schema.GroupKind) (*ResourceContext, error) {
	gkKey := store.ToKey(gk.String())
	return r.Storer.Get(gkKey)
}

func (r *APIStore) GetStorage(gk schema.GroupKind) (rest.Storage, error) {
	gkKey := store.ToKey(gk.String())
	resctx, err := r.Storer.Get(gkKey)
	if err != nil {
		return nil, err
	}
	return resctx.Storage, nil
}

// return the gvks as seen by the client
func (r *APIStore) GetExternalGVKSet() sets.Set[schema.GroupVersionKind] {
	gvkset := sets.New[schema.GroupVersionKind]()
	r.List(func(k store.Key, rc *ResourceContext) {
		gvkset.Insert(rc.ExternalGVK())
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
