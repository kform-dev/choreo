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

package object

import (
	"context"
	"fmt"

	"github.com/henderiw/logger/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/structured-merge-diff/v4/value"
)

func PruneUnmanagedFields(obj runtime.Unstructured, reconcilerName string) {
	// get the managed fields from the resource
	managedFields, found, err := unstructured.NestedSlice(obj.UnstructuredContent(), "metadata", "managedFields")
	if err != nil || !found {
		return
	}

	// 1. before-first-apply: fields set before the SSA was performed
	// since we always use SSA this should always result in NO fields in the map
	// we cannot prune these fields if they were present
	// 2. any field not managed by this reconciler aka fieldmanager should be pruned
	nonManagedFields := map[string]any{}
	var beforeFirstApplyFields map[string]any
	for _, field := range managedFields {
		fieldMap, ok := field.(map[string]interface{})
		if !ok {
			continue
		}
		if fieldMap["manager"] != reconcilerName && fieldMap["operation"] == "Apply" {
			fieldsV1, ok := fieldMap["fieldsV1"].(map[string]interface{})
			if ok {
				mergeFieldMaps(nonManagedFields, fieldsV1)
			}
		}
		if fieldMap["manager"] == "before-first-apply" && fieldMap["operation"] == "Update" {
			beforeFirstApplyFields = fieldMap["fieldsV1"].(map[string]interface{})
		}
	}
	// fields managed using before apply should not be pruned
	pruneFieldsUsingBeforeApply(nonManagedFields, beforeFirstApplyFields)
	// debug only node object for now
	debug := ""
	/*
		if obj.GetObjectKind().GroupVersionKind().Kind == "Link" {
			debug = "link"
			accessor, err := meta.Accessor(obj)
			if err == nil {
				debug = fmt.Sprintf("%s %s %s", r.name, obj.GetObjectKind().GroupVersionKind().Kind, accessor.GetName())
				fmt.Println(debug, "nonManagedFields", nonManagedFields)
				fmt.Println(debug, "beforeFirstApplyFields", beforeFirstApplyFields)
			}
		}
	*/

	data := obj.UnstructuredContent()
	// prune fields from unstructured data
	pruneFieldsFromUnstructured(debug, data, nonManagedFields)
	obj.SetUnstructuredContent(data)
	// SSA does not like to see managed fields in the object
	removeManagedFieldsFromUnstructured(obj)
	// resource version and generation are picked up from the stored resource
	// -> prune data from the resource
	removeResourceVersionAndGenerationFromUnstructured(obj)
	// debug only node object for now
	/*
		if obj.GetObjectKind().GroupVersionKind().Kind == "Link" {
			accessor, err := meta.Accessor(obj)
			if err == nil {
				fmt.Println(r.name, "link", accessor.GetName(), "object\n", obj.UnstructuredContent())
			}
		}
	*/
}

func pruneFieldsFromUnstructured(debug string, obj map[string]interface{}, nonManagedFields map[string]interface{}) {
	for nonManagedFieldKey, nonManagedFieldValue := range nonManagedFields {
		// Remove the "f:" prefix to match against the actual fields in obj
		realKey := nonManagedFieldKey[2:]
		if nonManagedFieldMap, ok := nonManagedFieldValue.(map[string]interface{}); ok {
			if objMap, exists := obj[realKey]; exists {
				if debug != "" {
					fmt.Println(debug, "pruneFieldsFromUnstructured, map exists realKey", realKey, nonManagedFieldValue, objMap, nonManagedFieldMap)
				}
				switch objMap := objMap.(type) {
				case map[string]interface{}:
					// this takes care of the fact that a map contains no element in the nonManagedFieldMap
					// as such we can delete the entry w/o going through the list
					if len(nonManagedFieldMap) == 0 {
						delete(obj, realKey)
					} else {
						pruneFieldsFromUnstructured(debug, objMap, nonManagedFieldMap)
						// when the map is empty is pruned we delete the complete element from the map
						if len(objMap) == 0 {
							delete(obj, realKey)
						}
					}
				case []interface{}:
					// this takes care of the fact that a map contains no element in the nonManagedFieldMap
					// as such we can delete the entry w/o going through the list
					if len(nonManagedFieldMap) == 0 {
						delete(obj, realKey)
					} else {
						obj[realKey] = pruneList(objMap, nonManagedFieldMap)
						// when the last entry is pruned we delete the complete element from the map
						if len(obj[realKey].([]interface{})) == 0 {
							delete(obj, realKey)
						}
					}
				default:
					// the objMap represents the real value
					delete(obj, realKey)
				}
			}
		} else {
			delete(obj, realKey)
		}
	}
}

func pruneList(objList []interface{}, nonManagedFields map[string]interface{}) []interface{} {
	for i := len(objList) - 1; i >= 0; i-- {
		item := objList[i]
		// Handle "k:" for list items
		for nonManagedFieldKey := range nonManagedFields {
			pe, err := NewPathElement(nonManagedFieldKey)
			if err != nil {
				continue
			}
			switch {
			case pe.Key != nil:
				// Handle "k:" prefix for list items (keys)
				objMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				matches := true
				for _, v := range *pe.Key {
					if objVal, exists := objMap[v.Name]; !exists || fmt.Sprintf("%v", objVal) != v.Value.AsString() {
						matches = false
						break
					}
				}
				if matches {
					objList = append(objList[:i], objList[i+1:]...)
				}
			case pe.Value != nil:
				// Handle "v:" prefix for list items (values)
				if fmt.Sprintf("%q", item) == value.ToString(*pe.Value) {
					// Remove the item if it matches the managed value
					objList = append(objList[:i], objList[i+1:]...)
				}
			}
		}
	}
	return objList
}

// pruneFieldsUsingBeforeApply prunes fields in nonManagedFields that were set using beforeApplyFields
// beforeApplyFields are fields set before the first SSA was done, so these are owned by a special manager
// (before-first-apply)
func pruneFieldsUsingBeforeApply(nonManagedFields, beforeApplyFields map[string]interface{}) {
	for beforeKey, beforeValue := range beforeApplyFields {
		if value, found := nonManagedFields[beforeKey]; found {
			// If it's a nested map, recursively prune
			if nestedMap, ok := value.(map[string]interface{}); ok {
				pruneFieldsUsingBeforeApply(nestedMap, beforeValue.(map[string]interface{}))
				if len(nestedMap) == 0 {
					delete(nonManagedFields, beforeKey)
				}
			}
		}
	}
}

// mergeFieldMaps merges several managed fieldsmaps together
func mergeFieldMaps(dest, src map[string]interface{}) {
	for key, value := range src {
		if existingValue, exists := dest[key]; exists {
			// If the key exists and both are maps, merge recursively
			if existingMap, ok := existingValue.(map[string]interface{}); ok {
				if valueMap, ok := value.(map[string]interface{}); ok {
					mergeFieldMaps(existingMap, valueMap)
					continue
				}
			}
		}
		// Otherwise, just set the value
		dest[key] = value
	}
}

func removeManagedFieldsFromUnstructured(obj runtime.Unstructured) {
	log := log.FromContext(context.Background())
	// Access the unstructured content
	unstructuredContent := obj.UnstructuredContent()

	// Access the metadata section
	metadata, found, err := unstructured.NestedMap(unstructuredContent, "metadata")
	if err != nil || !found {
		return
	}

	// Remove the managedFields key from metadata
	delete(metadata, "managedFields")

	// Set the updated metadata back to the unstructured content
	err = unstructured.SetNestedMap(unstructuredContent, metadata, "metadata")
	if err != nil {
		log.Error("error setting metadata", "error", err)
	}
}

func removeResourceVersionAndGenerationFromUnstructured(obj runtime.Unstructured) {
	log := log.FromContext(context.Background())
	// Access the unstructured content
	unstructuredContent := obj.UnstructuredContent()

	// Access the metadata section
	metadata, found, err := unstructured.NestedMap(unstructuredContent, "metadata")
	if err != nil || !found {
		return
	}

	// Remove the resourceVersion key from metadata
	delete(metadata, "resourceVersion")
	delete(metadata, "generation")

	// Set the updated metadata back to the unstructured content
	err = unstructured.SetNestedMap(unstructuredContent, metadata, "metadata")
	if err != nil {
		log.Error("error setting metadata", "error", err)
	}
}
