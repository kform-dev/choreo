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

package ipam

import (
	"context"
	"fmt"
	"sync"

	"github.com/henderiw/logger/log"
	condv1alpha1 "github.com/kform-dev/choreo/apis/condition/v1alpha1"
	ipamv1alpha1 "github.com/kform-dev/choreo/apis/kuid/backend/ipam/v1alpha1"
	bebackend "github.com/kform-dev/choreo/pkg/backend"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"k8s.io/apimachinery/pkg/runtime"
)

func New() bebackend.Backend {
	cache := bebackend.NewCache[*CacheInstanceContext]()
	return &be{
		cache: cache,
	}
}

type be struct {
	cache bebackend.Cache[*CacheInstanceContext]
	m     sync.RWMutex
	// added later
	entryStorage rest.Storage
	claimStorage rest.Storage
}

func (r *be) AddStorage(entryStorage, claimStorage rest.Storage) {
	r.entryStorage = entryStorage
	r.claimStorage = claimStorage
}

// CreateIndex creates a backend index
func (r *be) CreateIndex(ctx context.Context, obj runtime.Unstructured) error {
	r.m.Lock()
	defer r.m.Unlock()
	index := &ipamv1alpha1.IPIndex{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), index)
	if err != nil {
		return fmt.Errorf("error converting unstructured to ipIndex: %v", err)
	}

	ctx = bebackend.InitIndexContext(ctx, "create", index)
	log := log.FromContext(ctx)
	log.Debug("start")
	key := index.GetKey()

	log.Debug("start", "isInitialized", r.cache.IsInitialized(ctx, key))
	// if the Cache is not initialized -> restore the cache
	// this happens upon initialization or backend restart
	if _, err := r.cache.Get(ctx, key); err != nil {
		// if it does not exist create the cache
		cacheInstanceCtx := NewCacheInstanceContext()
		r.cache.Create(ctx, key, cacheInstanceCtx)
	}

	if !r.cache.IsInitialized(ctx, key) {
		if err := r.restore(ctx, index); err != nil {
			log.Error("cannot restore index", "error", err.Error())
			return err
		}
		log.Debug("restored")
		index.SetConditions(condv1alpha1.Ready())
		status, err := index.GetStatus()
		if err != nil {
			return err
		}
		objdata := obj.UnstructuredContent()
		objdata["status"] = status
		obj.SetUnstructuredContent(objdata)

		return r.cache.SetInitialized(ctx, key)
	}
	log.Debug("finished")
	return nil
}

// DeleteIndex deletes a backend index
func (r *be) DeleteIndex(ctx context.Context, obj runtime.Unstructured) error {
	r.m.Lock()
	defer r.m.Unlock()
	index := &ipamv1alpha1.IPIndex{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), index)
	if err != nil {
		return fmt.Errorf("error converting unstructured to ipIndex: %v", err)
	}

	ctx = bebackend.InitIndexContext(ctx, "delete", index)
	log := log.FromContext(ctx)
	log.Debug("start")
	key := index.GetKey()

	log.Debug("start", "isInitialized", r.cache.IsInitialized(ctx, key))
	// delete the data from the backend
	if err := r.destroy(ctx, key); err != nil {
		log.Error("cannot delete Index", "error", err.Error())
		return err
	}
	r.cache.Delete(ctx, key)

	log.Debug("finished")
	return nil
}

func (r *be) Claim(ctx context.Context, obj runtime.Unstructured) error {
	r.m.Lock()
	defer r.m.Unlock()
	claim := &ipamv1alpha1.IPClaim{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), claim)
	if err != nil {
		return fmt.Errorf("error converting unstructured to ipclaim: %v", err)
	}

	ctx = initClaimContext(ctx, "create", claim)
	log := log.FromContext(ctx)
	log.Debug("start")

	cacheCtx, err := r.cache.Get(ctx, claim.GetKey())
	if err != nil {
		return err
	}
	if !r.cache.IsInitialized(ctx, claim.GetKey()) {
		return fmt.Errorf("cache not initialized")
	}

	a, err := getApplicator(ctx, cacheCtx, claim)
	if err != nil {
		return err
	}
	if err := a.Validate(ctx, claim); err != nil {
		return err
	}
	if err := a.Apply(ctx, claim); err != nil {
		return err
	}
	// store the resources in the backend
	if err := r.saveAll(ctx, claim.GetKey()); err != nil {
		return err
	}
	status, err := claim.GetStatus()
	if err != nil {
		return err
	}
	objdata := obj.UnstructuredContent()
	objdata["status"] = status
	obj.SetUnstructuredContent(objdata)
	return nil

}

func (r *be) Release(ctx context.Context, obj runtime.Unstructured) error {
	r.m.Lock()
	defer r.m.Unlock()
	claim := &ipamv1alpha1.IPClaim{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), claim)
	if err != nil {
		return fmt.Errorf("error converting unstructured to ipclaim: %v", err)
	}

	ctx = initClaimContext(ctx, "delete", claim)
	log := log.FromContext(ctx)
	log.Debug("start")

	cacheCtx, err := r.cache.Get(ctx, claim.GetKey())
	if err != nil {
		return err
	}
	if !r.cache.IsInitialized(ctx, claim.GetKey()) {
		return fmt.Errorf("cache not initialized")
	}

	a, err := getApplicator(ctx, cacheCtx, claim)
	if err != nil {
		return err
	}
	if err := a.Delete(ctx, claim); err != nil {
		return err
	}

	return r.saveAll(ctx, claim.GetKey())
}

func getApplicator(_ context.Context, cacheInstanceCtx *CacheInstanceContext, claim *ipamv1alpha1.IPClaim) (Applicator, error) {
	ipClaimType, err := claim.GetIPClaimType()
	if err != nil {
		return nil, err
	}
	var a Applicator
	switch ipClaimType {
	case ipamv1alpha1.IPClaimType_StaticAddress:
		a = &staticAddressApplicator{name: string(ipamv1alpha1.IPClaimType_StaticAddress), applicator: applicator{cacheInstanceCtx: cacheInstanceCtx}}
	case ipamv1alpha1.IPClaimType_StaticPrefix:
		a = &staticPrefixApplicator{name: string(ipamv1alpha1.IPClaimType_StaticPrefix), applicator: applicator{cacheInstanceCtx: cacheInstanceCtx}}
	case ipamv1alpha1.IPClaimType_StaticRange:
		a = &staticRangeApplicator{name: string(ipamv1alpha1.IPClaimType_StaticRange), applicator: applicator{cacheInstanceCtx: cacheInstanceCtx}}
	case ipamv1alpha1.IPClaimType_DynamicAddress:
		a = &dynamicAddressApplicator{name: string(ipamv1alpha1.IPClaimType_DynamicAddress), applicator: applicator{cacheInstanceCtx: cacheInstanceCtx}}
	case ipamv1alpha1.IPClaimType_DynamicPrefix:
		a = &dynamicPrefixApplicator{name: string(ipamv1alpha1.IPClaimType_DynamicPrefix), applicator: applicator{cacheInstanceCtx: cacheInstanceCtx}}
	default:
		return nil, fmt.Errorf("invalid addressing, got: %s", string(ipClaimType))
	}

	return a, nil
}
