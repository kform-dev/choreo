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

	"github.com/kform-dev/choreo/pkg/server/selector"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (r *Reconciler) Validate() error {
	name := r.Name
	var errm error
	if err := r.Spec.For.ResourceGVK.Validate(); err != nil {
		errm = errors.Join(errm, fmt.Errorf("reconciler %s for gvk validation failed: %s", name, err.Error()))
	}
	if _, err := selector.ExprSelectorAsSelector(r.Spec.For.Selector); err != nil {
		errm = errors.Join(errm, fmt.Errorf("reconciler %s for selector validation failed: %s", name, err.Error()))
	}
	for _, own := range r.Spec.Owns {
		if err := own.ResourceGVK.Validate(); err != nil {
			errm = errors.Join(errm, fmt.Errorf("reconciler %s own gvk validation failed: %s", name, err.Error()))
		}
		if own.Selector != nil {
			errm = errors.Join(errm, fmt.Errorf("reconciler %s own resource does not expect a selector", name))
		}
	}
	for _, watch := range r.Spec.Watches {
		if err := watch.ResourceGVK.Validate(); err != nil {
			errm = errors.Join(errm, fmt.Errorf("reconciler %s watch gvk validation failed: %s", name, err.Error()))
		}
		if watch.Selector == nil {
			errm = errors.Join(errm, fmt.Errorf("reconciler %s watch resource expects a selector", name))
		} else {
			if watch.Selector.MatchExpressions != nil {
				errm = errors.Join(errm, fmt.Errorf("reconciler %s watch resource only support matches", name))
			}
		}
		if _, err := selector.SelectorAsSelectorBuilder(watch.Selector); err != nil {
			errm = errors.Join(errm, fmt.Errorf("reconciler %s watch selector validation failed: %s", name, err.Error()))
		}
	}

	// TODO check if the for gvk is not used by the watch and own gvks
	// TODO check if the own/watch are unique
	return errm
}

func (r *Reconciler) GetGVKs() sets.Set[schema.GroupVersionKind] {
	gvks := sets.New[schema.GroupVersionKind]()
	gvks.Insert(r.Spec.For.ResourceGVK.GetGVK())
	for _, own := range r.Spec.Owns {
		gvks.Insert(own.ResourceGVK.GetGVK())
	}
	for _, watch := range r.Spec.Watches {
		gvks.Insert(watch.ResourceGVK.GetGVK())
	}
	return gvks
}

func (r *Reconciler) GetOwnsGVKs() sets.Set[schema.GroupVersionKind] {
	gvks := sets.New[schema.GroupVersionKind]()
	for _, own := range r.Spec.Owns {
		gvks.Insert(own.ResourceGVK.GetGVK())
	}
	return gvks
}

func (r *Reconciler) GetForGVK() schema.GroupVersionKind {
	return r.Spec.For.ResourceGVK.GetGVK()
}

func (r ResourceGVK) Validate() error {
	var errm error
	if r.Group == "" {
		errm = errors.Join(errm, fmt.Errorf("group missing in gvk"))
	}
	if r.Version == "" {
		errm = errors.Join(errm, fmt.Errorf("version missing in gvk"))
	}
	if r.Kind == "" {
		errm = errors.Join(errm, fmt.Errorf("kind missing in gvk"))
	}
	return errm
}

func (r ResourceGVK) GetGVK() schema.GroupVersionKind {
	gv := schema.GroupVersion{
		Group:   r.Group,
		Version: r.Version,
	}
	return gv.WithKind(r.Kind)
}
