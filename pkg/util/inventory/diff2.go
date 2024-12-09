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
	"fmt"

	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/proto/snapshotpb"
	"github.com/kform-dev/choreo/pkg/util/object"
	"github.com/sdcio/config-server/apis/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (inv Inventory) Diff(invBefore Inventory, diff *choreov1alpha1.Diff, opts *snapshotpb.Diff_Options) error {
	diff.Status.Items = []*choreov1alpha1.DiffItem{}

	invSetsBefore := invBefore.Sets()

	for ref, treeNode := range inv {
		if !opts.ShowChoreoAPIs && treeNode.ChoreoAPI {
			continue
		}

		if opts.ShowFinalConfig {
			if treeNode.Resource == nil {
				continue
			}
			labels := treeNode.Resource.GetLabels()
			if _, ok := labels["config.sdcio.dev/finalconfig"]; !ok {
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

			fmt.Println("running config check")

			if rctreeNode := inv.GetResource(ref.APIVersion, config.RunningConfigKind, ref.Name, ref.Namespace); rctreeNode != nil {
				fmt.Println("running config exists")
				copiedBeforeObj := rctreeNode.Resource.DeepCopy()
				copiedAfterObj := treeNode.Resource.DeepCopy()

				runningConfig, _, err := unstructured.NestedFieldCopy(copiedBeforeObj.Object, "status", "value")
				if err != nil {
					return err
				}

				/*
					runningConfigBytes, err := yaml.Marshal(runningConfig)
					if err != nil {
						return fmt.Errorf("cannot marshal running config")
					}
				*/

				// Access spec.config[0].value
				configSlice, found, err := unstructured.NestedSlice(copiedAfterObj.Object, "spec", "config")
				if err != nil {
					return fmt.Errorf("error accessing spec.config in new config resource")
				}
				if !found {
					return fmt.Errorf("error accessing spec.config not found new config resource")
				}

				// Check if the slice has at least one element
				if len(configSlice) == 0 {
					return fmt.Errorf("error accessing spec.config is empty new config resource")
				}

				// Access the first element in the slice
				firstConfig, ok := configSlice[0].(map[string]interface{})
				if !ok {
					return fmt.Errorf("error spec.config[0] is not a map for new config resource")
				}

				// Access the "value" field in the first element
				newConfig, found, err := unstructured.NestedFieldNoCopy(firstConfig, "value")
				if err != nil {
					return fmt.Errorf("error accessing new config resource")
				}

				/*
					newConfigBytes, err := yaml.Marshal(newConfig)
					if err != nil {
						return fmt.Errorf("cannot marshal newconfig")
					}
				*/

				diffStr, err := diffconfig(runningConfig, newConfig)
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
				diff.Status.Items = append(diff.Status.Items, diffItem)
			}
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
			copiedAfterObj := treeNode.Resource.DeepCopy()
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
			copiedAfterObj := treeNode.Resource.DeepCopy()
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
		if opts.ShowFinalConfig {
			if treeNode.Resource == nil {
				continue
			}
			labels := treeNode.Resource.GetLabels()
			if _, ok := labels["config.sdcio.dev/finalconfig"]; !ok {
				continue
			}
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
