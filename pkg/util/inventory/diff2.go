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

package inventory

import (
	"context"

	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/util/object"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (inv Inventory) Diff(invBefore Inventory, diff *choreov1alpha1.Diff, opts *choreopb.Diff_Options) error {
	diff.Status.Items = []*choreov1alpha1.DiffItem{}

	invSetsBefore := invBefore.Sets()

	for ref, treenode := range inv {
		if !opts.ShowChoreoAPIs && treenode.ChoreoAPI {
			continue
		}
		diffItem := &choreov1alpha1.DiffItem{
			ResourceGVK: choreov1alpha1.ResourceGVK{
				Group:   ref.GroupVersionKind().Group,
				Version: ref.GroupVersionKind().Version,
				Kind:    ref.GroupVersionKind().Kind,
			},
			Name:      ref.Name,
			Namespace: ref.Namespace,
		}

		if invSetsBefore.Has(ref) {
			//diffStr, err := diff(invBefore[ref].Resource, treenode.Resource)
			copiedBeforeObj := invBefore[ref].Resource.DeepCopy()
			copiedAfterObj := treenode.Resource.DeepCopy()
			if !opts.ShowManagedField {
				object.RemoveManagedFieldsFromUnstructured(context.Background(), copiedBeforeObj)
				object.RemoveManagedFieldsFromUnstructured(context.Background(), copiedAfterObj)
			}

			diffStr, err := diff2(copiedBeforeObj, copiedAfterObj)
			if err != nil {
				return err
			}
			if diffStr == "" {
				diffItem.Status = choreov1alpha1.DiffitemStatus_Equal
			} else {
				diffItem.Status = choreov1alpha1.DiffitemStatus_Modified
				diffItem.Diff = &diffStr
			}
			invSetsBefore.Delete(ref)
		} else {

			copiedAfterObj := treenode.Resource.DeepCopy()
			if !opts.ShowManagedField {
				object.RemoveManagedFieldsFromUnstructured(context.Background(), copiedAfterObj)
			}

			diffItem.Status = choreov1alpha1.DiffitemStatus_Added
			diffStr, err := diff2(&unstructured.Unstructured{}, copiedAfterObj)
			if err != nil {
				return err
			}
			diffItem.Diff = &diffStr
		}
		diff.Status.Items = append(diff.Status.Items, diffItem)
	}
	for _, ref := range invSetsBefore.UnsortedList() {
		treeNode := invBefore[ref]
		if !opts.ShowChoreoAPIs && treeNode.ChoreoAPI {
			continue
		}

		diffItem := &choreov1alpha1.DiffItem{
			ResourceGVK: choreov1alpha1.ResourceGVK{
				Group:   ref.GroupVersionKind().Group,
				Version: ref.GroupVersionKind().Version,
				Kind:    ref.GroupVersionKind().Kind,
			},
			Name:      ref.Name,
			Namespace: ref.Namespace,
			Status:    choreov1alpha1.DiffitemStatus_Deleted,
		}

		copiedBeforeObj := treeNode.Resource.DeepCopy()
		if !opts.ShowManagedField {
			object.RemoveManagedFieldsFromUnstructured(context.Background(), copiedBeforeObj)
		}

		diffStr, err := diff2(copiedBeforeObj, &unstructured.Unstructured{})
		if err != nil {
			return err
		}
		diffItem.Diff = &diffStr

		diff.Status.Items = append(diff.Status.Items, diffItem)
	}
	return nil
}
