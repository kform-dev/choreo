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
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	selectorv1alpha1 "github.com/kform-dev/choreo/apis/selector/v1alpha1"
	"github.com/kform-dev/choreo/pkg/server/selector"
	"github.com/kform-dev/kform/pkg/fsys"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r *ConfigGenerator) Validate(path string) error {
	name := r.Name
	var errm error
	if err := r.Spec.ProviderSelector.Validate(); err != nil {
		errm = errors.Join(errm, fmt.Errorf("configgenerator %s for gvk validation failed: %s", name, err.Error()))
	}
	if _, err := selector.SelectorAsSelectorBuilder(&selectorv1alpha1.ExpressionSelector{
		Match: r.Spec.ProviderSelector.Match,
	}); err != nil {
		errm = errors.Join(errm, fmt.Errorf("configgenerator %s for selector validation failed: %s", name, err.Error()))
	}

	if r.Spec.Dir == nil {
		errm = errors.Join(errm, fmt.Errorf("configgenerator %s baspath is required", name))
	} else {
		if !fsys.IsDir(filepath.Join(path, *r.Spec.Dir)) {
			errm = errors.Join(errm, fmt.Errorf("configgenerator %s baspath does not exist", name))
		}
	}
	// TODO check if the for gvk is not used by the watch and own gvks
	// TODO check if the own/watch are unique
	return errm
}

func (r *ConfigGenerator) GetDir() string {
	if r.Spec.Dir == nil {
		return ""
	}
	return *r.Spec.Dir
}

func (r *ConfigGenerator) GetProviderSelectorGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   r.Spec.ProviderSelector.Group,
		Version: r.Spec.ProviderSelector.Version,
		Kind:    r.Spec.ProviderSelector.Kind,
	}
}

func (r *ConfigGenerator) GetMatchKeys() []string {
	keys := make([]string, len(r.Spec.ProviderSelector.Match))
	i := 0
	for k := range r.Spec.ProviderSelector.Match {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func (r *ConfigGenerator) GetProviderMatch() map[string]string {
	return r.Spec.ProviderSelector.Match
}
