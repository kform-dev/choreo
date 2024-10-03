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
	"errors"
	"fmt"

	"github.com/henderiw/idxtable/pkg/tree/gtree"
	"github.com/henderiw/iputil"
	"github.com/henderiw/store"
	condv1alpha1 "github.com/kform-dev/choreo/apis/condition/v1alpha1"
	"github.com/kform-dev/choreo/apis/kuid/backend"
	commonv1alpha1 "github.com/kuidio/kuid/apis/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

var _ backend.IndexObject = &IPIndex{}

func (r *IPIndex) GetKey() store.Key {
	return store.KeyFromNSN(r.GetNamespacedName())
}

func (r *IPIndex) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.GetNamespace(),
		Name:      r.GetName(),
	}
}

// GetCondition returns the condition based on the condition kind
func (r *IPIndex) GetCondition(t condv1alpha1.ConditionType) condv1alpha1.Condition {
	return r.Status.GetCondition(t)
}

// SetConditions sets the conditions on the resource. it allows for 0, 1 or more conditions
// to be set at once
func (r *IPIndex) SetConditions(c ...condv1alpha1.Condition) {
	r.Status.SetConditions(c...)
}

func (r *IPIndex) ValidateSyntax(s string) field.ErrorList {
	var allErrs field.ErrorList

	if len(r.Spec.Prefixes) == 0 {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec.minID"),
			r,
			fmt.Errorf("a ipindex needs a prefix").Error(),
		))

	}

	return allErrs
}

// GetTree satisfies the interface but should not be used
func (r *IPIndex) GetTree() gtree.GTree { return nil }

func (r *IPIndex) GetType() string { return "" }

// GetMinID satisfies the interface but should not be used
func (r *IPIndex) GetMinID() *uint64 { return nil }

// GetMaxID satisfies the interface but should not be used
func (r *IPIndex) GetMaxID() *uint64 { return nil }

// GetMinClaim satisfies the interface but should not be used
func (r *IPIndex) GetMinClaim() backend.ClaimObject { return nil }

// GetMaxClaim satisfies the interface but should not be used
func (r *IPIndex) GetMaxClaim() backend.ClaimObject { return nil }

func (r *IPIndex) GetMax() uint64 { return 0 }

func (r *IPIndex) GetStatus() (map[string]any, error) {
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

func (r *IPIndex) GetClaims() ([]*IPClaim, error) {
	ipclaims := make([]*IPClaim, len(r.Spec.Prefixes))
	var errm, err error
	for i, prefix := range r.Spec.Prefixes {
		ipclaims[i], err = r.GetClaim(prefix)
		if err != nil {
			errm = errors.Join(errm, err)
		}
	}
	if errm != nil {
		return nil, errm
	}
	return ipclaims, nil
}

func (r *IPIndex) GetClaim(prefix Prefix) (*IPClaim, error) {
	pi, err := iputil.New(prefix.Prefix)
	if err != nil {
		return nil, err
	}

	return BuildIPClaim(
		metav1.ObjectMeta{
			Namespace: r.GetNamespace(),
			Name:      fmt.Sprintf("%s.%s", r.Name, pi.GetSubnetName()),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: r.APIVersion,
					Kind:       r.Kind,
					Name:       r.Name,
					UID:        r.UID,
				},
			},
		},
		&IPClaimSpec{
			Index:        r.Name,
			PrefixType:   prefix.PrefixType,
			Prefix:       ptr.To(prefix.Prefix),
			PrefixLength: ptr.To(uint32(pi.GetPrefixLength())),
			CreatePrefix: ptr.To(true),
			ClaimLabels: commonv1alpha1.ClaimLabels{
				UserDefinedLabels: prefix.UserDefinedLabels,
			},
		},
		nil,
	), nil
}
