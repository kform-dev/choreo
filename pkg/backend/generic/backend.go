package generic

import (
	"context"
	"fmt"
	"sync"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	condv1alpha1 "github.com/kform-dev/choreo/apis/condition/v1alpha1"
	"github.com/kform-dev/choreo/apis/kuid/backend"
	bebackend "github.com/kform-dev/choreo/pkg/backend"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func New(
	indexGVK schema.GroupVersionKind,
	claimGVK schema.GroupVersionKind,
	entryGVK schema.GroupVersionKind,
	indexObjectFn func(runtime.Unstructured) (backend.IndexObject, error),
	claimObjectFn func(runtime.Unstructured) (backend.ClaimObject, error),
	entryObjectFn func(runtime.Unstructured) (backend.EntryObject, error),
	entryFromCacheFn func(k store.Key, vrange, id string, labels map[string]string) backend.EntryObject,
) bebackend.Backend {

	cache := bebackend.NewCache[*CacheInstanceContext]()

	return &be{
		cache:            cache,
		indexGVK:         indexGVK,
		claimGVK:         claimGVK,
		entryGVK:         entryGVK,
		indexObjectFn:    indexObjectFn,
		claimObjectFn:    claimObjectFn,
		entryObjectFn:    entryObjectFn,
		entryFromCacheFn: entryFromCacheFn,
	}
}

type be struct {
	cache            bebackend.Cache[*CacheInstanceContext]
	m                sync.RWMutex
	indexGVK         schema.GroupVersionKind
	claimGVK         schema.GroupVersionKind
	entryGVK         schema.GroupVersionKind
	indexObjectFn    func(runtime.Unstructured) (backend.IndexObject, error)
	claimObjectFn    func(runtime.Unstructured) (backend.ClaimObject, error)
	entryObjectFn    func(runtime.Unstructured) (backend.EntryObject, error)
	entryFromCacheFn func(k store.Key, vrange, id string, labels map[string]string) backend.EntryObject
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
	//r.m.Lock()
	//defer r.m.Unlock()
	index, err := r.indexObjectFn(obj)
	if err != nil {
		return err
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
		cacheInstanceCtx := NewCacheInstanceContext(index.GetTree(), index.GetType())
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
	objidx, err := r.indexObjectFn(obj)
	if err != nil {
		return err
	}
	ctx = bebackend.InitIndexContext(ctx, "delete", objidx)
	log := log.FromContext(ctx)
	log.Debug("start")
	key := objidx.GetKey()

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
	claim, err := r.claimObjectFn(obj)
	if err != nil {
		return err
	}

	ctx = bebackend.InitClaimContext(ctx, "create", claim)
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
	claim, err := r.claimObjectFn(obj)
	if err != nil {
		return err
	}

	ctx = bebackend.InitClaimContext(ctx, "delete", claim)
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

func getApplicator(_ context.Context, cacheInstanceCtx *CacheInstanceContext, claim backend.ClaimObject) (Applicator, error) {
	claimType := claim.GetClaimType()
	var a Applicator
	switch claimType {
	case backend.ClaimType_DynamicID:
		a = &dynamicApplicator{name: string(claimType), applicator: applicator{cacheInstanceCtx: cacheInstanceCtx}}
	case backend.ClaimType_StaticID:
		a = &staticApplicator{name: string(claimType), applicator: applicator{cacheInstanceCtx: cacheInstanceCtx}}
	case backend.ClaimType_Range:
		a = &rangeApplicator{name: string(claimType), applicator: applicator{cacheInstanceCtx: cacheInstanceCtx}}
	default:
		return nil, fmt.Errorf("invalid addressing, got: %s", string(claimType))
	}

	return a, nil
}
