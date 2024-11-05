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
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	ChoreoAPIEmbeddedKey  = "api.choreo.kform.dev/embedded"
	ChoreoAPIInternalKey  = "api.choreo.kform.dev/internal"
	ChoreoLoaderOriginKey = "api.choreo.kform.dev/origin"
)

func HasChoreoAPIAnnotation(u *unstructured.Unstructured) bool {
	a := u.GetAnnotations()
	for k := range a {
		if k == ChoreoAPIEmbeddedKey || k == ChoreoAPIInternalKey {
			return true
		}
	}
	return false
}

func HasChoreoEmbeddedAPIAnnotation(a map[string]string) bool {
	for k := range a {
		if k == ChoreoAPIEmbeddedKey {
			return true
		}
	}
	return false
}

type LoaderAnnotation struct {
	Kind      string `json:"kind"`
	URL       string `json:"url,omitempty"`
	Directory string `json:"directory,omitempty"`
	Ref       string `json:"ref,omitempty"`
}

func (r LoaderAnnotation) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		fmt.Println("loader annotation stringify failed", err)
		return ""
	}
	return string(b)
}

func ToLoaderAnnotation(s string) LoaderAnnotation {
	a := LoaderAnnotation{}
	if err := json.Unmarshal([]byte(s), &a); err != nil {
		return LoaderAnnotation{}
	}
	return a
}

var FileLoaderAnnotation = LoaderAnnotation{
	Kind: "File",
}

func GetFileInputAnnotation() string {
	return FileLoaderAnnotation.String()
}
