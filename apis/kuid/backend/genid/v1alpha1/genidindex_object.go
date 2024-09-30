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
	"github.com/henderiw/idxtable/pkg/tree/tree16"
	"github.com/henderiw/idxtable/pkg/tree/tree32"
	"github.com/henderiw/idxtable/pkg/tree/tree64"
	"github.com/henderiw/store"
	condv1alpha1 "github.com/kform-dev/choreo/apis/condition/v1alpha1"
	"github.com/kform-dev/choreo/apis/kuid/backend"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

var _ backend.IndexObject = &GENIDIndex{}

func (r *GENIDIndex) GetKey() store.Key {
	return store.KeyFromNSN(r.GetNamespacedName())
}

func (r *GENIDIndex) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.GetNamespace(),
		Name:      r.GetName(),
	}
}

// GetCondition returns the condition based on the condition kind
func (r *GENIDIndex) GetCondition(t condv1alpha1.ConditionType) condv1alpha1.Condition {
	return r.Status.GetCondition(t)
}

// SetConditions sets the conditions on the resource. it allows for 0, 1 or more conditions
// to be set at once
func (r *GENIDIndex) SetConditions(c ...condv1alpha1.Condition) {
	r.Status.SetConditions(c...)
}

func (r *GENIDIndex) ValidateSyntax(s string) field.ErrorList {
	var allErrs field.ErrorList

	if GetGenIDType(r.Spec.Type) == GENIDType_Invalid {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec.type"),
			r,
			fmt.Errorf("invalid GENID Type %s", r.Spec.Type).Error(),
		))
	}

	if r.Spec.MinID != nil {
		if err := validateGENIDID(GetGenIDType(r.Spec.Type), *r.Spec.MinID); err != nil {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec.minID"),
				r,
				fmt.Errorf("invalid GENID ID %d", *r.Spec.MinID).Error(),
			))
		}
	}
	if r.Spec.MaxID != nil {
		if err := validateGENIDID(GetGenIDType(r.Spec.Type), *r.Spec.MaxID); err != nil {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec.maxID"),
				r,
				fmt.Errorf("invalid GENID ID %d", *r.Spec.MaxID).Error(),
			))
		}
	}
	if r.Spec.MinID != nil && r.Spec.MaxID != nil {
		if *r.Spec.MinID > *r.Spec.MaxID {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec.maxID"),
				r,
				fmt.Errorf("min GENID ID %d cannot be bigger than max GENID ID %d", *r.Spec.MinID, *r.Spec.MaxID).Error(),
			))
		}
	}
	return allErrs
}

func (r *GENIDIndex) GetTree() gtree.GTree {
	switch GetGenIDType(r.Spec.Type) {
	case GENIDType_16bit:
		tree, err := tree16.New(16)
		if err != nil {
			return nil
		}
		return tree
	case GENIDType_32bit:
		tree, err := tree32.New(32)
		if err != nil {
			return nil
		}
		return tree
	case GENIDType_48bit:
		tree, err := tree64.New(48)
		if err != nil {
			return nil
		}
		return tree
	case GENIDType_64bit:
		tree, err := tree64.New(64)
		if err != nil {
			return nil
		}
		return tree
	default:
		return nil
	}
}

func (r *GENIDIndex) GetType() string {
	return r.Spec.Type
}

func (r *GENIDIndex) GetMinID() *uint64 {
	if r.Spec.MinID == nil {
		return nil
	}
	return ptr.To(*r.Spec.MinID)
}

func (r *GENIDIndex) GetMaxID() *uint64 {
	if r.Spec.MinID == nil {
		return nil
	}
	return ptr.To(*r.Spec.MaxID)
}

func (r *GENIDIndex) GetMax() uint64 {
	return GENIDID_MaxValue[GetGenIDType(r.Spec.Type)]
}

func GetMinClaimRange(id uint64) string {
	return fmt.Sprintf("%d-%d", GENIDID_Min, id-1)
}

func GetMaxClaimRange(genidType GENIDType, id uint64) string {
	return fmt.Sprintf("%d-%d", id+1, GENIDID_MaxValue[genidType])
}

func (r *GENIDIndex) GetMinClaimNSN() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      fmt.Sprintf("%s.%s", r.Name, backend.IndexReservedMinName),
	}
}

func (r *GENIDIndex) GetMaxClaimNSN() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      fmt.Sprintf("%s.%s", r.Name, backend.IndexReservedMaxName),
	}
}

func (r *GENIDIndex) GetMinClaim() backend.ClaimObject {
	return BuildGENIDClaim(
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
		&GENIDClaimSpec{
			Index: r.Name,
			Range: ptr.To(GetMinClaimRange(*r.Spec.MinID)),
		},
		nil,
	)
}

func (r *GENIDIndex) GetMaxClaim() backend.ClaimObject {
	return BuildGENIDClaim(
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
		&GENIDClaimSpec{
			Index: r.Name,
			Range: ptr.To(GetMaxClaimRange(GetGenIDType(r.Spec.Type), *r.Spec.MaxID)),
		},
		nil,
	)
}

func (r *GENIDIndex) GetStatus() (map[string]any, error) {
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

func GENIDIndexFromUnstructured(ru runtime.Unstructured) (backend.IndexObject, error) {
	obj := &GENIDIndex{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(ru.UnstructuredContent(), obj)
	if err != nil {
		return nil, fmt.Errorf("error converting unstructured to GENIDIndex: %v", err)
	}
	return obj, nil
}
