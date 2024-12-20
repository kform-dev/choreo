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
	"fmt"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	selectorv1alpha1 "github.com/kform-dev/choreo/apis/selector/v1alpha1"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/server/selector"
	"github.com/kuidio/kuid/apis/backend"
	genericbe "github.com/kuidio/kuid/pkg/backend/generic"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewChoreoGenericBackendstorage(
	indexKind string,
	entryStorage, claimStorage rest.Storage,
	entryObjectFn func(runtime.Unstructured) (backend.EntryObject, error),
	claimObjectFn func(runtime.Unstructured) (backend.ClaimObject, error),
) genericbe.BackendStorage {
	return &kuidgenericbe{
		indexKind:     indexKind,
		entryStorage:  entryStorage,
		claimStorage:  claimStorage,
		claimObjectFn: claimObjectFn,
		entryObjectFn: entryObjectFn,
	}
}

type kuidgenericbe struct {
	indexKind     string
	entryStorage  rest.Storage
	claimStorage  rest.Storage
	entryObjectFn func(runtime.Unstructured) (backend.EntryObject, error)
	claimObjectFn func(runtime.Unstructured) (backend.ClaimObject, error)
}

func (r *kuidgenericbe) ListEntries(ctx context.Context, k store.Key) ([]backend.EntryObject, error) {
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
	entryList := make([]backend.EntryObject, 0)
	items, ok := ul.UnstructuredContent()["items"]
	if ok && len(items.([]interface{})) > 0 {
		ul.EachListItem(func(obj runtime.Object) error {
			ru, ok := obj.(runtime.Unstructured)
			if !ok {
				return fmt.Errorf("not unstructured")
			}
			entryObj, err := r.entryObjectFn(ru)
			if err != nil {
				return fmt.Errorf("not entry object")
			}
			entryList = append(entryList, entryObj)
			return nil
		})
	}
	return entryList, nil
}

func (r *kuidgenericbe) CreateEntry(ctx context.Context, obj backend.EntryObject) error {
	return r.applyEntry(ctx, obj)
}

func (r *kuidgenericbe) UpdateEntry(ctx context.Context, obj, old backend.EntryObject) error {
	return r.applyEntry(ctx, obj)
}

func (r *kuidgenericbe) applyEntry(ctx context.Context, obj backend.EntryObject) error {
	log := log.FromContext(ctx)
	newuobj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}
	newu := &unstructured.Unstructured{
		Object: newuobj,
	}
	newu.SetAPIVersion(obj.GetChoreoAPIVersion())
	if _, err := r.entryStorage.Apply(ctx, newu, &rest.ApplyOptions{
		FieldManager: "backend",
		DryRun:       []string{"recursion"},
	}); err != nil {
		log.Error("cannot apply entry", "error", err)
		return err
	}
	return nil
}

func (r *kuidgenericbe) DeleteEntry(ctx context.Context, obj backend.EntryObject) error {
	log := log.FromContext(ctx)
	if _, err := r.entryStorage.Delete(ctx, obj.GetName(), &rest.DeleteOptions{
		DryRun: []string{"recursion"},
	}); err != nil {
		log.Error("cannot delete entry", "error", err)
		return err
	}
	return nil
}

func (r *kuidgenericbe) ListClaims(ctx context.Context, k store.Key, opts ...genericbe.ListOption) (map[string]backend.ClaimObject, error) {
	o := &genericbe.ListOptions{}
	o.ApplyOptions(opts)

	log := log.FromContext(ctx).With("implementation", "choreo generic backend")
	log.Debug("listClaims")

	match := map[string]string{
		"spec.index": k.Name,
	}
	if o.OwnerKind != "" {
		match[fmt.Sprintf("metadata.ownerReferences.exists(ref, ref.kind == %q)", r.indexKind)] = "true"
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
	claimMap := make(map[string]backend.ClaimObject)
	items, ok := ul.UnstructuredContent()["items"]
	if ok && len(items.([]interface{})) > 0 {
		ul.EachListItem(func(obj runtime.Object) error {
			ru, ok := obj.(runtime.Unstructured)
			if !ok {
				return fmt.Errorf("not unstructured")
			}

			claimObj, err := r.claimObjectFn(ru)
			if err != nil {
				return fmt.Errorf("not claim object")
			}
			claimMap[claimObj.GetNamespacedName().String()] = claimObj
			return nil
		})
	}
	return claimMap, nil
}

func (r *kuidgenericbe) CreateClaim(ctx context.Context, obj backend.ClaimObject) error {
	return r.applyClaim(ctx, obj)
}

func (r *kuidgenericbe) UpdateClaim(ctx context.Context, obj, old backend.ClaimObject) error {
	return r.applyClaim(ctx, obj)
}

func (r *kuidgenericbe) applyClaim(ctx context.Context, obj backend.ClaimObject) error {
	log := log.FromContext(ctx)
	newuobj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}
	newu := &unstructured.Unstructured{
		Object: newuobj,
	}
	newu.SetAPIVersion(obj.GetChoreoAPIVersion())
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

func (r *kuidgenericbe) DeleteClaim(ctx context.Context, obj backend.ClaimObject) error {
	log := log.FromContext(ctx)
	if _, err := r.claimStorage.Delete(ctx, obj.GetName(), &rest.DeleteOptions{
		DryRun: []string{"recursion"},
	}); err != nil {
		log.Error("cannot delete entry", "error", err)
		return err
	}
	return nil
}
