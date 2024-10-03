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
	"errors"
	"fmt"
	"sort"

	"github.com/henderiw/idxtable/pkg/iptable"
	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	ipamv1alpha1 "github.com/kform-dev/choreo/apis/kuid/backend/ipam/v1alpha1"
	selectorv1alpha1 "github.com/kform-dev/choreo/apis/selector/v1alpha1"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/server/selector"
	"github.com/kform-dev/choreo/pkg/util/object"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func (r *be) restore(ctx context.Context, index *ipamv1alpha1.IPIndex) error {
	log := log.FromContext(ctx)
	k := index.GetKey()

	cacheInstanceCtx, err := r.cache.Get(ctx, k)
	if err != nil {
		log.Error("cannot get index", "error", err.Error())
		return err
	}

	// Fetch the current entries that were stored
	curEntries, err := r.listEntries(ctx, k)
	if err != nil {
		return err
	}

	claimmap, err := r.listClaims(ctx, map[string]string{
		"spec.index": index.Name,
	})
	if err != nil {
		return nil
	}

	/*
		prefixes := make(map[string]ipamv1alpha1.Prefix)
		for _, prefix := range index.Spec.Prefixes {
			prefixes[prefix.Prefix] = prefix
		}
	*/

	/*
		if err := r.restoreIndexPrefixes(ctx, cacheInstanceCtx, curEntries, index, prefixes); err != nil {
			return err
		}
	*/
	if err := r.restoreClaims(ctx, cacheInstanceCtx, curEntries, ipamv1alpha1.IPIndexKind, ipamv1alpha1.IPClaimType_StaticPrefix, claimmap); err != nil {
		return err
	}
	if err := r.restoreClaims(ctx, cacheInstanceCtx, curEntries, ipamv1alpha1.IPClaimKind, ipamv1alpha1.IPClaimType_StaticPrefix, claimmap); err != nil {
		return err
	}
	if err := r.restoreClaims(ctx, cacheInstanceCtx, curEntries, ipamv1alpha1.IPClaimKind, ipamv1alpha1.IPClaimType_StaticRange, claimmap); err != nil {
		return err
	}
	if err := r.restoreClaims(ctx, cacheInstanceCtx, curEntries, ipamv1alpha1.IPClaimKind, ipamv1alpha1.IPClaimType_DynamicPrefix, claimmap); err != nil {
		return err
	}
	if err := r.restoreClaims(ctx, cacheInstanceCtx, curEntries, ipamv1alpha1.IPClaimKind, ipamv1alpha1.IPClaimType_StaticAddress, claimmap); err != nil {
		return err
	}
	if err := r.restoreClaims(ctx, cacheInstanceCtx, curEntries, ipamv1alpha1.IPClaimKind, ipamv1alpha1.IPClaimType_DynamicAddress, claimmap); err != nil {
		return err
	}
	log.Debug("restore prefixes entries left", "items", len(curEntries))

	return nil
}

func (r *be) saveAll(ctx context.Context, k store.Key) error {
	log := log.FromContext(ctx)
	log.Debug("SaveAll", "key", k.String())

	// entries from the memory cache
	newEntries, err := r.getEntriesFromCache(ctx, k)
	if err != nil {
		return err
	}
	// entries in the apiserver
	curEntries, err := r.listEntries(ctx, k)
	if err != nil {
		return err
	}

	news := []string{}
	for _, newEntry := range newEntries {
		news = append(news, newEntry.Name)
	}
	curs := []string{}
	for _, curEntry := range curEntries {
		curs = append(curs, curEntry.Name)
	}
	sort.Strings(news)
	sort.Strings(curs)

	for _, newEntry := range newEntries {
		log.Debug("SaveAll", "newIPEntry", newEntry.GetNamespacedName())
		newEntry := newEntry
		//found := false
		//var entry *ipamv1alpha1.IPEntry
		for idx, curEntry := range curEntries {
			log.Debug("SaveAll", "curEntry", *curEntry)
			idx := idx
			curEntry := curEntry
			if curEntry.GetNamespace() == newEntry.GetNamespace() &&
				curEntry.GetName() == newEntry.GetName() {
				curEntries = append(curEntries[:idx], curEntries[idx+1:]...)
				//found = true
				//entry = curEntry
				break
			}
		}

		uobj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(newEntry)
		if err != nil {
			return err
		}
		newu := &unstructured.Unstructured{
			Object: uobj,
		}

		if _, err := r.entryStorage.Apply(ctx, newu, &rest.ApplyOptions{
			FieldManager: "backend",
		}); err != nil {
			return err
		}
	}
	for _, curEntry := range curEntries {
		if _, err := r.entryStorage.Delete(ctx, curEntry.GetName(), &rest.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}

// Destroy removes the store db
func (r *be) destroy(ctx context.Context, k store.Key) error {
	// no need to delete the index as this is what this fn is supposed to do
	return r.deleteEntries(ctx, k)
}

func (r *be) getEntriesFromCache(ctx context.Context, k store.Key) ([]*ipamv1alpha1.IPEntry, error) {
	//log := log.FromContext(ctx).With("key", k.String())
	cacheInstanceCtx, err := r.cache.Get(ctx, k)
	if err != nil {
		return nil, fmt.Errorf("cache index not initialized")
	}

	entries := make([]*ipamv1alpha1.IPEntry, 0, cacheInstanceCtx.Size())
	// add the main rib entry
	for _, route := range cacheInstanceCtx.rib.GetTable() {
		route := route
		entries = append(entries, ipamv1alpha1.GetIPEntry(ctx, k, route.Prefix(), route.Labels()))
	}
	// add all the range entries
	cacheInstanceCtx.ranges.List(func(key store.Key, t iptable.IPTable) {
		for _, route := range t.GetAll() {
			route := route
			entries = append(entries, ipamv1alpha1.GetIPEntry(ctx, k, route.Prefix(), route.Labels()))
		}
	})

	return entries, nil
}

func (r *be) deleteEntries(ctx context.Context, k store.Key) error {
	log := log.FromContext(ctx)

	entries, err := r.listEntries(ctx, k)
	if err != nil {
		log.Error("cannot list entries", "error", err)
		return err
	}

	var errm error
	for _, entry := range entries {
		if _, err := r.entryStorage.Delete(ctx, entry.GetName(), &rest.DeleteOptions{}); err != nil {
			log.Error("cannot delete entry", "error", err)
			errm = errors.Join(errm, err)
			continue
		}
	}
	return errm
}

func (r *be) listEntries(ctx context.Context, k store.Key) ([]*ipamv1alpha1.IPEntry, error) {
	selector, err := selector.ExprSelectorAsSelector(
		&selectorv1alpha1.ExpressionSelector{
			Match: map[string]string{
				"spec.index": k.Name,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	ul, err := r.entryStorage.List(ctx, &rest.ListOptions{
		Selector:          selector,
		ShowManagedFields: true,
	})
	if err != nil {
		return nil, err
	}
	entryList := make([]*ipamv1alpha1.IPEntry, 0)
	if ul.IsList() {
		ul.EachListItem(func(obj runtime.Object) error {
			ru, ok := obj.(runtime.Unstructured)
			if !ok {
				return fmt.Errorf("not unstructured")
			}
			entryObj := &ipamv1alpha1.IPEntry{}
			runtime.DefaultUnstructuredConverter.FromUnstructured(ru.UnstructuredContent(), entryObj)
			entryList = append(entryList, entryObj)
			return nil
		})
	}
	return entryList, nil
}

func (r *be) listClaims(ctx context.Context, match map[string]string) (map[string]*ipamv1alpha1.IPClaim, error) {
	selector, err := selector.ExprSelectorAsSelector(
		&selectorv1alpha1.ExpressionSelector{
			Match: match,
		},
	)
	if err != nil {
		return nil, err
	}
	ul, err := r.claimStorage.List(ctx, &rest.ListOptions{
		Selector:          selector,
		ShowManagedFields: true,
	})
	if err != nil {
		return nil, err
	}
	claimmap := map[string]*ipamv1alpha1.IPClaim{}
	if ul.IsList() {
		ul.EachListItem(func(obj runtime.Object) error {
			ru, ok := obj.(runtime.Unstructured)
			if !ok {
				return fmt.Errorf("not unstructured")
			}
			claimObj := &ipamv1alpha1.IPClaim{}
			runtime.DefaultUnstructuredConverter.FromUnstructured(ru.UnstructuredContent(), claimObj)
			claimmap[claimObj.GetNamespacedName().String()] = claimObj
			return nil
		})
	}
	return claimmap, nil
}

func (r *be) restoreClaims(ctx context.Context, cacheInstanceCtx *CacheInstanceContext, entries []*ipamv1alpha1.IPEntry, kind string, claimType ipamv1alpha1.IPClaimType, ipclaimmap map[string]*ipamv1alpha1.IPClaim) error {
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		for _, ownerref := range entry.GetOwnerReferences() {
			if ownerref.APIVersion == ipamv1alpha1.SchemeGroupVersion.Identifier() &&
				ownerref.Kind == kind {
				if claimType == entry.Spec.ClaimType {
					nsn := types.NamespacedName{Namespace: entry.GetNamespace(), Name: ownerref.Name}
					claim, ok := ipclaimmap[nsn.String()]
					if ok {
						if err := r.restoreClaim(ctx, cacheInstanceCtx, claim); err != nil {
							return err
						}
						// remove the entry since it is processed
						entries = append(entries[:i], entries[i+1:]...)
						delete(ipclaimmap, nsn.String()) // delete the entry to optimize
					}
				}
			}
		}
	}
	return nil
}

func (r *be) restoreClaim(ctx context.Context, cacheInstanceCtx *CacheInstanceContext, claim *ipamv1alpha1.IPClaim) error {
	ctx = initClaimContext(ctx, "restore", claim)
	a, err := getApplicator(ctx, cacheInstanceCtx, claim)
	if err != nil {
		return err
	}
	// validate is needed, mainly for addresses since the parent route determines
	// e.g. the fact the address belongs to a range or not
	errList := claim.ValidateSyntax("") // needed to expand the createPrefix/prefixLength and owner
	if len(errList) != 0 {
		return fmt.Errorf("invalid syntax %v", errList)
	}
	if err := a.Validate(ctx, claim); err != nil {
		return err
	}
	if err := a.Apply(ctx, claim); err != nil {
		return err
	}
	return nil
}

func (r *be) updateIPIndexClaims(ctx context.Context, index *ipamv1alpha1.IPIndex) error {
	key := index.GetKey()

	newClaims, err := index.GetClaims()
	if err != nil {
		return err
	}

	match := map[string]string{
		"spec.index": key.Name,
		"metadata.ownerReferences.exists(ref, ref.kind == 'IPIndexKind')": "true",
	}
	existingClaims, err := r.listClaims(ctx, match)
	if err != nil {
		return err
	}

	var errm error
	for _, claim := range newClaims {
		u, err := object.GetUnstructructered(claim)
		if err != nil {
			errm = errors.Join(errm, err)
			continue
		}
		if _, err := r.claimStorage.Apply(ctx, u, &rest.ApplyOptions{
			FieldManager: "backend",
		}); err != nil {
			errm = errors.Join(errm, err)
			continue
		}

	}

	for _, claim := range existingClaims {
		if _, err := r.claimStorage.Delete(ctx, claim.GetName()); err != nil {
			errm = errors.Join(errm, err)
			continue
		}
	}

	if errm != nil {
		return errm
	}

	return r.saveAll(ctx, key)
}
