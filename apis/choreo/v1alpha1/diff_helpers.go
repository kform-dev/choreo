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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func BuildDiff(meta metav1.ObjectMeta, spec *DiffSpec, status *DiffStatus) *Diff {
	diffSpec := DiffSpec{}
	if spec != nil {
		diffSpec = *spec
	}
	diffStatus := DiffStatus{}
	if spec != nil {
		diffStatus = *status
	}
	return &Diff{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.Identifier(),
			Kind:       DiffKind,
		},
		ObjectMeta: meta,
		Spec:       diffSpec,
		Status:     diffStatus,
	}
}

func (r *DiffItem) GetStatusSymbol() string {
	switch r.Status {
	case DiffitemStatus_Added:
		return "+"
	case DiffitemStatus_Modified:
		return "~"
	case DiffitemStatus_Deleted:
		return "-"
	case DiffitemStatus_Equal:
		return "="
	default:
		return ""
	}
}
