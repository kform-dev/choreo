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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

type ContentType string

const (
	ContentTypeYAML ContentType = "yaml"
	ContentTypeJSON ContentType = "json"
)

func GetUnstructuredListContent(b []byte, ct ContentType) (*unstructured.UnstructuredList, error) {
	// Unmarshal JSON data into a map
	var data map[string]interface{}
	switch ct {
	case ContentTypeYAML:
		if err := yaml.Unmarshal(b, &data); err != nil {
			return nil, err
		}
	case ContentTypeJSON:
		if err := json.Unmarshal(b, &data); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid contentType got %s, expecting %s or %s", string(ct), ContentTypeYAML, ContentTypeJSON)
	}

	// Create an empty Unstructured object
	u := &unstructured.UnstructuredList{}
	// Set the map to the Unstructured object
	u.SetUnstructuredContent(data)
	return u, nil
}

func GetUnstructuredContent(b []byte, ct ContentType) (*unstructured.Unstructured, error) {
	// Unmarshal JSON data into a map
	var data map[string]interface{}
	switch ct {
	case ContentTypeYAML:
		if err := yaml.Unmarshal(b, &data); err != nil {
			return nil, err
		}
	case ContentTypeJSON:
		if err := json.Unmarshal(b, &data); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid contentType got %s, expecting %s or %s", string(ct), ContentTypeYAML, ContentTypeJSON)
	}

	// Create an empty Unstructured object
	u := &unstructured.Unstructured{}
	// Set the map to the Unstructured object
	u.SetUnstructuredContent(data)
	return u, nil
}

func GetUnstructured(b []byte) (*unstructured.Unstructured, error) {
	u, err := GetUnstructuredContent(b, ContentTypeJSON)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid object, err: %s", err.Error())
	}
	return u, nil
}

func GetUnstructuredList(b []byte) (*unstructured.UnstructuredList, error) {
	u, err := GetUnstructuredListContent(b, ContentTypeJSON)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid object, err: %s", err.Error())
	}
	return u, nil
}

func Marshal(u *unstructured.Unstructured) ([]byte, error) {
	return json.Marshal(u)
}

func GetUnstructuredFromFile(filename string) (*unstructured.Unstructured, error) {
	b, err := GetFileName(filename)
	if err != nil {
		return nil, err
	}
	u := &unstructured.Unstructured{}
	switch filepath.Ext(filename) {
	case ".yaml":
		if err := json.Unmarshal(b, u); err != nil {
			return nil, err
		}
	case ".json":
		if err := yaml.Unmarshal(b, u); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported file format, filename: %s, format: %s", filename, filepath.Ext(filename))
	}
	return u, nil

}

func GetFileName(filename string) ([]byte, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("file cannot be a directory")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return convertFile(filename, data)
}

func convertFile(filename string, inputData []byte) ([]byte, error) {
	var data any
	switch filepath.Ext(filename) {
	case ".yaml":
		// Validates the yaml file
		err := yaml.Unmarshal(inputData, &data)
		if err != nil {
			return nil, err
		}
		// Marshal the data to JSON
		return json.Marshal(data)
	case ".json":
		// validates the json file
		err := json.Unmarshal(inputData, &data)
		if err != nil {
			return nil, err
		}
		// Marshal the data to JSON
		return json.Marshal(data)
	default:
		return nil, fmt.Errorf("unsupported file format, filename: %s, format: %s", filename, filepath.Ext(filename))
	}
}

func GetGeneration(ru runtime.Unstructured) int64 {
	u := &unstructured.Unstructured{
		Object: ru.UnstructuredContent(),
	}
	metadata, ok := u.Object["metadata"]
	if ok {
		metadata, ok := metadata.(map[string]any)
		if ok {
			generation, ok := metadata["generation"]
			if ok {
				switch g := generation.(type) {
				case float64:
					return int64(g)
				case float32:
					return int64(g)
				case int:
					return int64(g)
				case int8:
					return int64(g)
				case int16:
					return int64(g)
				case int32:
					return int64(g)
				case int64:
					return int64(g)
				default:
					return int64(0)
				}
			}
		}
	}
	return int64(0)
}

func Empty(obj runtime.Unstructured) runtime.Unstructured {
	empty := &unstructured.Unstructured{
		Object: map[string]interface{}{},
	}
	empty.SetAPIVersion(obj.GetObjectKind().GroupVersionKind().GroupVersion().Identifier())
	empty.SetKind(obj.GetObjectKind().GroupVersionKind().Kind)
	return empty
}

func IsConditionTypeReady(obj map[string]interface{}, conditionType string) bool {
	if obj == nil {
		return false
	}
	status := GetStatus(obj)
	conditions := GetConditions(status)
	for _, condition := range conditions {
		condition, ok := condition.(map[string]any)
		if !ok {
			continue
		}
		if condition["type"] == conditionType {
			v, ok := condition["status"]
			if !ok {
				return false
			}
			status, ok := v.(string)
			if !ok {
				return false
			}
			return status == "True"
		}

	}
	return false
}

func GetCondition(obj map[string]interface{}, conditionType string) map[string]any {
	if obj == nil {
		obj = map[string]any{}
	}

	status := GetStatus(obj)
	conditions := GetConditions(status)

	for i := len(conditions) - 1; i >= 0; i-- {
		condition, ok := conditions[i].(map[string]any)
		if !ok {
			continue
		}
		if condition["type"] == conditionType {
			return condition
		}
	}
	return EmptyCondition(conditionType)

}

func SetCondition(obj map[string]interface{}, conditionType, msg string) {
	if obj == nil {
		obj = map[string]any{}
	}

	status := GetStatus(obj)
	conditions := GetConditions(status)
	newcondition := NewCondition(conditionType, msg)

	found := false
	for i := len(conditions) - 1; i >= 0; i-- {
		condition, ok := conditions[i].(map[string]any)
		if !ok {
			continue
		}
		if condition["type"] == conditionType {
			found = true
			if !Equal(condition, newcondition) {
				// update new conditions
				conditions[i] = newcondition
			}
		} else {
			conditions = append(conditions[:i], conditions[i+1:]...)
		}
	}
	if !found {
		conditions = append(conditions, newcondition)
	}
	status["conditions"] = conditions
}

// Equal returns true if the condition is identical to the supplied condition,
// ignoring the LastTransitionTime.
func Equal(cond1, cond2 map[string]any) bool {
	return cond1["type"] == cond2["type"] &&
		cond1["status"] == cond2["status"] &&
		cond1["reason"] == cond2["reason"] &&
		cond1["message"] == cond2["message"]
}

func EmptyCondition(conditionType string) map[string]any {
	return map[string]any{
		"type":    conditionType,
		"message": "unknown",
		"reason":  "unknown",
		"status":  "unknown",
	}
}

func NewCondition(conditionType, msg string) map[string]any {
	reason := "Ready"
	condStatus := "True" // don't use conditionStatus
	if msg != "" {
		reason = "Failed"
		condStatus = "False" // don't use conditionStatus
	}
	return map[string]any{
		"type":               conditionType,
		"message":            msg,
		"reason":             reason,
		"status":             condStatus,
		"lastTransitionTime": metav1.Now(),
	}
}

func GetStatus(obj map[string]any) map[string]any {
	// Check if the status exists
	statusObj, exists := obj["status"]
	if !exists {
		// Initialize status if it does not exist
		statusObj = make(map[string]any)
		obj["status"] = statusObj
	}

	// The status is expected to be a map[string]interface{}, safe cast as it's either the one we just created or already validated
	status, _ := statusObj.(map[string]any)
	return status
}

func GetConditions(status map[string]any) []any {
	conditionsObj, exists := status["conditions"]
	if !exists {
		conditionsObj = make([]any, 0)
		status["conditions"] = conditionsObj
	}
	conditions, _ := conditionsObj.([]any)
	return conditions
}

func GetObjectRefFromUnstructured(u *unstructured.Unstructured) corev1.ObjectReference {
	return corev1.ObjectReference{
		APIVersion: u.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       u.GetObjectKind().GroupVersionKind().Kind,
		Namespace:  u.GetNamespace(),
		Name:       u.GetName(),
	}
}

func GetObjectRefFromOwnerRef(namespace string, ownref metav1.OwnerReference) corev1.ObjectReference {
	return corev1.ObjectReference{
		APIVersion: ownref.APIVersion,
		Kind:       ownref.Kind,
		Namespace:  namespace,
		Name:       ownref.Name,
	}
}

func GetUnstructructered(obj runtime.Object) (runtime.Unstructured, error) {
	b, err := yaml.Marshal(obj)
	if err != nil {
		return nil, err
	}
	object := map[string]any{}
	yaml.Unmarshal(b, &object)

	return &unstructured.Unstructured{
		Object: object,
	}, nil
}

func IsManagedBy(mfields []metav1.ManagedFieldsEntry, fieldManager string) bool {
	for _, mfentry := range mfields {
		if mfentry.Manager == fieldManager {
			return true
		}
	}
	return false
}

func Exists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

func EnsureDir(dirname string) error {
	if !Exists(dirname) {
		return os.MkdirAll(dirname, 0755)
	}
	return nil
}

func SetFinalizer(u *unstructured.Unstructured, finalizerStr string) {
	finalizers := u.GetFinalizers()
	if len(finalizers) == 0 {
		finalizers = []string{}
	}
	for _, finalizer := range finalizers {
		if finalizer == finalizerStr {
			return
		}
	}
	finalizers = append(finalizers, finalizerStr)
	u.SetFinalizers(finalizers)
}

func DeleteFinalizer(u *unstructured.Unstructured, finalizerStr string) {
	finalizers := u.GetFinalizers()
	for i, finalizer := range finalizers {
		if finalizer == finalizerStr {
			finalizers = append(finalizers[:i], finalizers[i+1:]...)
			u.SetFinalizers(finalizers)
		}
	}
}
