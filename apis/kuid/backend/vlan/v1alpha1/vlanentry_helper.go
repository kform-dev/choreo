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
	"github.com/kform-dev/choreo/apis/kuid/backend"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildVLANEntry returns a reource from a client Object a Spec/Status
func BuildVLANEntry(meta metav1.ObjectMeta, spec *VLANEntrySpec, status *VLANEntryStatus) backend.EntryObject {
	aspec := VLANEntrySpec{}
	if spec != nil {
		aspec = *spec
	}
	astatus := VLANEntryStatus{}
	if status != nil {
		astatus = *status
	}
	return &VLANEntry{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.Identifier(),
			Kind:       VLANEntryKind,
		},
		ObjectMeta: meta,
		Spec:       aspec,
		Status:     astatus,
	}
}
