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
	"context"
	"fmt"
	"net/netip"

	"github.com/henderiw/iputil"
	"github.com/henderiw/store"
	condv1alpha1 "github.com/kform-dev/choreo/apis/condition/v1alpha1"
	"github.com/kform-dev/choreo/apis/kuid/backend"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

func (r *IPEntry) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.GetNamespace(),
		Name:      r.GetName(),
	}
}

// GetCondition returns the condition based on the condition kind
func (r *IPEntry) GetCondition(t condv1alpha1.ConditionType) condv1alpha1.Condition {
	return r.Status.GetCondition(t)
}

// SetConditions sets the conditions on the resource. it allows for 0, 1 or more conditions
// to be set at once
func (r *IPEntry) SetConditions(c ...condv1alpha1.Condition) {
	r.Status.SetConditions(c...)
}

func GetIPEntry(ctx context.Context, k store.Key, prefix netip.Prefix, labels map[string]string) *IPEntry {
	//log := log.FromContext(ctx)
	pi := iputil.NewPrefixInfo(prefix)

	index := k.Name
	ns := k.Namespace

	spec := &IPEntrySpec{
		Index:         index,
		PrefixType:    GetIPPrefixTypeFromString(labels[backend.KuidIPAMIPPrefixTypeKey]),
		ClaimType:     GetIPClaimTypeFromString(labels[backend.KuidClaimTypeKey]),
		Prefix:        pi.String(),
		AddressFamily: ptr.To(pi.GetAddressFamily()),
	}
	if _, ok := labels[backend.KuidIPAMDefaultGatewayKey]; ok {
		spec.DefaultGateway = ptr.To[bool](true)
	}
	// filter the system defined labels from the labels to prepare for the user defined labels
	udLabels := map[string]string{}
	for k, v := range labels {
		if !backend.BackendSystemKeys.Has(k) && !backend.BackendIPAMSystemKeys.Has(k) {
			udLabels[k] = v
		}
	}
	spec.UserDefinedLabels.Labels = udLabels
	return BuildIPEntry(
		metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s.%s", index, pi.GetSubnetName()),
			Namespace: ns,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: SchemeGroupVersion.Identifier(),
					Kind:       labels[backend.KuidOwnerKindKey],
					Name:       labels[backend.KuidClaimNameKey],
					UID:        types.UID(labels[backend.KuidClaimUIDKey]),
				},
			},
		},
		spec,
		nil,
	)
}
