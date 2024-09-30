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
	"k8s.io/apimachinery/pkg/runtime"
)

func UnstructuredToStarlarkValue(u runtime.Unstructured) (starlark.Value, error) {
	if u == nil {
		return starlark.None, nil
	}
	data := u.UnstructuredContent() // Get the map[string]interface{} representation
	return mapToStarlarkDict(data)
}

func mapToStarlarkDict(data map[string]interface{}) (starlark.Value, error) {
	dict := new(starlark.Dict)
	for k, v := range data {
		sv, err := toStarlarkValue(v)
		if err != nil {
			return nil, err
		}
		dict.SetKey(starlark.String(k), sv)
	}
	return dict, nil
}

func toStarlarkValue(v interface{}) (starlark.Value, error) {
	switch v := v.(type) {
	case string:
		return starlark.String(v), nil
	case int64:
		return starlark.MakeInt64(v), nil
	case int, int32:
		return starlark.MakeInt(v.(int)), nil
	case float64:
		return starlark.Float(v), nil
	case bool:
		return starlark.Bool(v), nil
	case map[string]interface{}:
		return mapToStarlarkDict(v)
	case []interface{}:
		return sliceToStarlarkList(v)
	case nil:
		return starlark.None, nil // Handle nil explicitly
	default:
		return nil, fmt.Errorf("type %T not supported", v)
	}
}

func sliceToStarlarkList(slice []interface{}) (starlark.Value, error) {
	lst := starlark.NewList(nil)
	for _, item := range slice {
		sv, err := toStarlarkValue(item)
		if err != nil {
			return nil, err
		}
		lst.Append(sv)
	}
	return lst, nil
}
