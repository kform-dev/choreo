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
)

func StarlarkValueToMap(v starlark.Value) (map[string]any, error) {
	dict, ok := v.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("expected dict value to convert to map, got %T", v)
	}

	return starlarkDictToMap(dict)
}

// StarlarkIntToInt64 converts a starlark.Int to a Go int64 safely.
func StarlarkIntToInt64(starInt starlark.Int) int64 {
	bigInt := starInt.BigInt()
	return bigInt.Int64()
}

func StarlarkValueToGoMap(val starlark.Value) (map[string]string, error) {
	dict, ok := val.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("value is not a dictionary")
	}

	result := make(map[string]string)
	for _, key := range dict.Keys() {
		val, found, err := dict.Get(key)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}

		// Type assert the key and value to strings
		keyStr, ok := key.(starlark.String)
		if !ok {
			return nil, fmt.Errorf("key is not a string: %v", key)
		}
		valStr, ok := val.(starlark.String)
		if !ok {
			return nil, fmt.Errorf("value is not a string: %v", val)
		}

		// Add to the result map
		result[string(keyStr)] = string(valStr)
	}
	return result, nil
}
