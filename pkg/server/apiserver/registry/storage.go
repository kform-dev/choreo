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

package registry

import (
	"context"

	"github.com/henderiw/store"
	"github.com/henderiw/store/gitu"
	"github.com/henderiw/store/memoryu"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/server/apiserver/watchermanager"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/managedfields"
)

type NewFn func() runtime.Unstructured
type NewListFn func() runtime.Unstructured

type NewStorageFn func(ctx context.Context, dbpath string) rest.Storage

type StorageCreator struct {
	GroupResource schema.GroupResource
	NewFn         NewFn
	NewListFn     NewListFn
	Strategy      rest.Strategy
	FieldManager  *managedfields.FieldManager
}

func (r StorageCreator) NewStorage(ctx context.Context, pathInRepo, dbpath string) rest.Storage {
	return NewStorage(ctx, pathInRepo, dbpath, r.GroupResource, r.NewFn, r.NewListFn, r.Strategy, r.FieldManager)
}

func NewStorage(ctx context.Context, pathInRepo, dbpath string, gr schema.GroupResource, newFn NewFn, newListFn NewListFn, strategy rest.Strategy, fieldManager *managedfields.FieldManager) rest.Storage {
	watcherManager := watchermanager.New(64)
	go watcherManager.Start(ctx)

	var store store.UnstructuredStore
	if dbpath == "" {
		store = memoryu.NewStore()
	} else {
		var err error
		store, err = gitu.NewStore(&gitu.Config{
			GroupResource: gr,
			PathInRepo:    pathInRepo,
			RootPath:      dbpath,
			NewFunc:       newFn,
		})
		if err != nil {
			panic(err)
		}
	}

	return &storage{
		newFn:          newFn,
		newListFn:      newListFn,
		createStrategy: strategy,
		updateStrategy: strategy,
		deleteStrategy: strategy,
		storage:        store,
		watcherManager: watcherManager,
		fieldManager:   fieldManager,
	}
}

type storage struct {
	// NewFunc returns a new instance of the type this registry returns for a
	// GET of a single object
	newFn NewFn

	// NewListFn returns a new list of the type this registry; it is the
	// type returned when the resource is listed
	newListFn func() runtime.Unstructured

	createStrategy rest.CreateStrategy
	updateStrategy rest.UpdateStrategy
	deleteStrategy rest.DeleteStrategy

	fieldManager *managedfields.FieldManager

	// holds the data
	//storage        store.Storer[runtime.Unstructured]
	storage        store.UnstructuredStore
	watcherManager watchermanager.WatcherManager
}
