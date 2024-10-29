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

package crdloader

import (
	"context"
	"errors"
	"fmt"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	selectorv1alpha1 "github.com/kform-dev/choreo/apis/selector/v1alpha1"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/server/selector"
	"github.com/kuidio/kuid/apis/backend/ipam"
	ipambe "github.com/kuidio/kuid/pkg/backend/ipam"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewChoreoIPAMBackendstorage(
	entryStorage, claimStorage rest.Storage,
) ipambe.BackendStorage {
	return &kuidbe{
		entryStorage: entryStorage,
		claimStorage: claimStorage,
	}
}

type kuidbe struct {
	entryStorage rest.Storage
	claimStorage rest.Storage
}

func (r *kuidbe) ListEntries(ctx context.Context, k store.Key) ([]*ipam.IPEntry, error) {
	log := log.FromContext(ctx).With("implementation", "choreo ipam backend")
	log.Debug("listEntries")
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
	var errm error
	entryList := make([]*ipam.IPEntry, 0)
	if ul.IsList() {
		ul.EachListItem(func(obj runtime.Object) error {
			ru, ok := obj.(runtime.Unstructured)
			if !ok {
				return fmt.Errorf("not unstructured")
			}
			ipEntry := &ipam.IPEntry{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(ru.UnstructuredContent(), ipEntry)
			if err != nil {
				errm = errors.Join(errm, fmt.Errorf("error converting unstructured to ipEntry: %v", err))
				return nil
			}
			entryList = append(entryList, ipEntry)
			return nil
		})
	}
	return entryList, nil
}

func (r *kuidbe) CreateEntry(ctx context.Context, obj *ipam.IPEntry) error {
	return r.applyEntry(ctx, obj)
}

func (r *kuidbe) UpdateEntry(ctx context.Context, obj, old *ipam.IPEntry) error {
	return r.applyEntry(ctx, obj)
}

func (r *kuidbe) applyEntry(ctx context.Context, obj *ipam.IPEntry) error {
	newuobj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}
	newu := &unstructured.Unstructured{
		Object: newuobj,
	}
	newu.SetAPIVersion(schema.GroupVersion{Group: ipam.GroupName, Version: "ipam"}.String())
	if _, err := r.entryStorage.Apply(ctx, newu, &rest.ApplyOptions{
		FieldManager: "backend",
		DryRun:       []string{"recursion"},
	}); err != nil {
		return err
	}
	return nil
}

func (r *kuidbe) DeleteEntry(ctx context.Context, obj *ipam.IPEntry) error {
	log := log.FromContext(ctx)
	if _, err := r.entryStorage.Delete(ctx, obj.GetName(), &rest.DeleteOptions{
		DryRun: []string{"recursion"},
	}); err != nil {
		log.Error("cannot delete entry", "error", err)
		return err
	}
	return nil
}

func (r *kuidbe) ListClaims(ctx context.Context, k store.Key, opts ...ipambe.ListOption) (map[string]*ipam.IPClaim, error) {
	o := &ipambe.ListOptions{}
	o.ApplyOptions(opts)

	log := log.FromContext(ctx).With("implementation", "choreo ipam backend")
	log.Debug("listClaims")

	match := map[string]string{
		"spec.index": k.Name,
	}
	if o.OwnerKind != "" {
		match["metadata.ownerReferences.exists(ref, ref.kind == 'IPIndexKind')"] = "true"
	}
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
	claimmap := map[string]*ipam.IPClaim{}
	if ul.IsList() {
		ul.EachListItem(func(obj runtime.Object) error {
			ru, ok := obj.(runtime.Unstructured)
			if !ok {
				return fmt.Errorf("not unstructured")
			}
			claimObj := &ipam.IPClaim{}
			runtime.DefaultUnstructuredConverter.FromUnstructured(ru.UnstructuredContent(), claimObj)
			claimmap[claimObj.GetNamespacedName().String()] = claimObj
			return nil
		})
	}
	return claimmap, nil
}

func (r *kuidbe) CreateClaim(ctx context.Context, obj *ipam.IPClaim) error {
	return r.applyClaim(ctx, obj)
}

func (r *kuidbe) UpdateClaim(ctx context.Context, obj, old *ipam.IPClaim) error {
	return r.applyClaim(ctx, obj)
}

func (r *kuidbe) applyClaim(ctx context.Context, obj *ipam.IPClaim) error {
	log := log.FromContext(ctx)
	newuobj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}
	newu := &unstructured.Unstructured{
		Object: newuobj,
	}
	newu.SetAPIVersion(schema.GroupVersion{Group: ipam.GroupName, Version: "ipam"}.String())
	log.Debug("choreo apply claim", "obj", newu.Object)
	if _, err := r.claimStorage.Apply(ctx, newu, &rest.ApplyOptions{
		FieldManager: "backend",
		DryRun:       []string{"recursion"},
	}); err != nil {
		log.Error("choreo apply claim", "error", err)
		return err
	}
	return nil
}

func (r *kuidbe) DeleteClaim(ctx context.Context, obj *ipam.IPClaim) error {
	log := log.FromContext(ctx)
	if _, err := r.claimStorage.Delete(ctx, obj.GetName(), &rest.DeleteOptions{
		DryRun: []string{"recursion"},
	}); err != nil {
		log.Error("cannot delete entry", "error", err)
		return err
	}
	return nil
}
