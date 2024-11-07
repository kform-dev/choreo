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

package util

import (
	"fmt"

	"go.starlark.net/starlark"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Example usage: Create an unstructured.Unstructured from a starlark.Dict.
func StarlarkValueToUnstructured(v starlark.Value) (runtime.Unstructured, error) {
	dict, ok := v.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("expected dict value to convert to unstructured, got %T", v)
	}

	data, err := starlarkDictToMap(dict)
	if err != nil {
		return nil, err
	}
	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(data)
	return u, nil
}

// Convert starlark.Dict to a map[string]interface{}.
func starlarkDictToMap(dict *starlark.Dict) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, item := range dict.Items() {
		key, ok := item[0].(starlark.String)
		if !ok {
			return nil, fmt.Errorf("dict key is not a string: %s", item[0].Type())
		}
		value, err := starlarkValueToInterface(item[1])
		if err != nil {
			return nil, err
		}
		result[string(key)] = value
	}
	return result, nil
}

// Convert a starlark.Value to an interface{} that can be used by unstructured.Unstructured.
func starlarkValueToInterface(value starlark.Value) (interface{}, error) {
	switch v := value.(type) {
	case starlark.String:
		return string(v), nil
	case starlark.Bool:
		return bool(v), nil
	case starlark.Int:
		intVal, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("integer out of range")
		}
		return intVal, nil
	case starlark.Float:
		return float64(v), nil
	case *starlark.Dict:
		return starlarkDictToMap(v)
	case *starlark.List:
		return starlarkListToSlice(v)
	case starlark.NoneType:
		return "", nil
	default:
		if value == nil {
			return nil, fmt.Errorf("unsupported starlark type: nil")
		}
		return nil, fmt.Errorf("unsupported starlark type: %s", value.Type())
	}
}

// Convert starlark.List to a []interface{}.
func starlarkListToSlice(list *starlark.List) ([]interface{}, error) {
	var slice []interface{}
	for i := 0; i < list.Len(); i++ {
		item := list.Index(i)
		value, err := starlarkValueToInterface(item)
		if err != nil {
			return nil, err
		}
		slice = append(slice, value)
	}
	return slice, nil
}
