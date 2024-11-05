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
	"strings"

	"github.com/kform-dev/choreo/pkg/repository/git"
)

func (r *UpstreamRef) LoaderAnnotation() LoaderAnnotation {
	a := LoaderAnnotation{
		Kind: "Upstream",
		URL:  r.Spec.URL,
		Ref:  r.Spec.Ref.Name,
	}
	if r.Spec.Directory == nil {
		return a
	}
	a.Directory = *r.Spec.Directory
	return a
}

func (r *UpstreamRef) GetPlumbingReference() string {
	refName := r.Spec.Ref.Name
	if r.Spec.Ref.Type == RefType_Tag {
		refName = git.TagName(refName).TagInLocal().String()
	}
	return refName
}

func (r *UpstreamRef) GetPathInRepo() string {
	pathInRepo := "."
	if r.Spec.Directory != nil {
		pathInRepo = *r.Spec.Directory
	}
	return pathInRepo
}

func (r *UpstreamRef) GetURLPath() string {
	url := r.Spec.URL
	replace := strings.NewReplacer("/", "-", ":", "-")
	return replace.Replace(url)
}
