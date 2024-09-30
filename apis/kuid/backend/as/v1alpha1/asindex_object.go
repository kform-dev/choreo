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

package v1alpha1

import (
	"encoding/json"
	"fmt"

	"github.com/henderiw/idxtable/pkg/tree/gtree"
	"github.com/henderiw/idxtable/pkg/tree/tree32"
	"github.com/henderiw/store"
	condv1alpha1 "github.com/kform-dev/choreo/apis/condition/v1alpha1"
	"github.com/kform-dev/choreo/apis/kuid/backend"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

var _ backend.IndexObject = &ASIndex{}

func (r *ASIndex) GetKey() store.Key {
	return store.KeyFromNSN(r.GetNamespacedName())
}

func (r *ASIndex) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.GetNamespace(),
		Name:      r.GetName(),
	}
}

// GetCondition returns the condition based on the condition kind
func (r *ASIndex) GetCondition(t condv1alpha1.ConditionType) condv1alpha1.Condition {
	return r.Status.GetCondition(t)
}

// SetConditions sets the conditions on the resource. it allows for 0, 1 or more conditions
// to be set at once
func (r *ASIndex) SetConditions(c ...condv1alpha1.Condition) {
	r.Status.SetConditions(c...)
}

func (r *ASIndex) ValidateSyntax(s string) field.ErrorList {
	var allErrs field.ErrorList

	if r.Spec.MinID != nil {
		if err := validateASID(int(*r.Spec.MinID)); err != nil {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec.minID"),
				r,
				fmt.Errorf("invalid vlan ID %d", *r.Spec.MinID).Error(),
			))
		}
	}
	if r.Spec.MaxID != nil {
		if err := validateASID(int(*r.Spec.MaxID)); err != nil {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec.maxID"),
				r,
				fmt.Errorf("invalid vlan ID %d", *r.Spec.MaxID).Error(),
			))
		}
	}
	if r.Spec.MinID != nil && r.Spec.MaxID != nil {
		if *r.Spec.MinID > *r.Spec.MaxID {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec.maxID"),
				r,
				fmt.Errorf("min vlan ID %d cannot be bigger than max vlan ID %d", *r.Spec.MinID, *r.Spec.MaxID).Error(),
			))
		}
	}
	return allErrs
}

func (r *ASIndex) GetTree() gtree.GTree {
	tree, err := tree32.New(32)
	if err != nil {
		panic(err)
	}
	return tree
}

func (r *ASIndex) GetType() string {
	return ""
}

func (r *ASIndex) GetMinID() *uint64 {
	if r.Spec.MinID == nil {
		return nil
	}
	return ptr.To[uint64](uint64(*r.Spec.MinID))
}

func (r *ASIndex) GetMaxID() *uint64 {
	if r.Spec.MaxID == nil {
		return nil
	}
	return ptr.To[uint64](uint64(*r.Spec.MaxID))
}

func (r *ASIndex) GetMax() uint64 {
	return ASID_Max
}

func GetMinClaimRange(id uint32) string {
	return fmt.Sprintf("%d-%d", ASID_Min, id-1)
}

func GetMaxClaimRange(id uint32) string {
	return fmt.Sprintf("%d-%d", id+1, ASID_Max)
}

func (r *ASIndex) GetMinClaimNSN() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      fmt.Sprintf("%s.%s", r.Name, backend.IndexReservedMinName),
	}
}

func (r *ASIndex) GetMaxClaimNSN() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      fmt.Sprintf("%s.%s", r.Name, backend.IndexReservedMaxName),
	}
}

func (r *ASIndex) GetMinClaim() backend.ClaimObject {
	return BuildASClaim(
		metav1.ObjectMeta{
			Namespace: r.GetNamespace(),
			Name:      r.GetMinClaimNSN().Name,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: r.APIVersion,
					Kind:       r.Kind,
					Name:       r.Name,
					UID:        r.UID,
				},
			},
		},
		&ASClaimSpec{
			Index: r.Name,
			Range: ptr.To[string](GetMinClaimRange(*r.Spec.MinID)),
		},
		nil,
	)
}

func (r *ASIndex) GetMaxClaim() backend.ClaimObject {
	return BuildASClaim(
		metav1.ObjectMeta{
			Namespace: r.GetNamespace(),
			Name:      r.GetMaxClaimNSN().Name,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: r.APIVersion,
					Kind:       r.Kind,
					Name:       r.Name,
					UID:        r.UID,
				},
			},
		},
		&ASClaimSpec{
			Index: r.Name,
			Range: ptr.To[string](GetMaxClaimRange(*r.Spec.MaxID)),
		},
		nil,
	)
}

func (r *ASIndex) GetStatus() (map[string]any, error) {
	b, err := json.Marshal(r.Status)
	if err != nil {
		return nil, err
	}
	status := map[string]any{}
	if err := json.Unmarshal(b, &status); err != nil {
		return nil, err
	}
	return status, nil
}

func ASIndexFromUnstructured(ru runtime.Unstructured) (backend.IndexObject, error) {
	obj := &ASIndex{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(ru.UnstructuredContent(), obj)
	if err != nil {
		return nil, fmt.Errorf("error converting unstructured to asIndex: %v", err)
	}
	return obj, nil
}
