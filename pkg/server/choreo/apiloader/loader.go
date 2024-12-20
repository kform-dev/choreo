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

package apiloader

import (
	"context"

	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/choreo/crdloader"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

const (
	ManagedFieldManagerInput = "inputfileloader"
)

func GetCRDFromUnstructured(u *unstructured.Unstructured) (*apiextensionsv1.CustomResourceDefinition, error) {
	b, err := yaml.Marshal(u.Object)
	if err != nil {
		return nil, err
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(b, crd); err != nil {
		return nil, err
	}
	return crd, nil
}

type APIStoreLoader struct {
	APIStore     *api.APIStore
	InternalGVKs sets.Set[schema.GroupVersionKind]
	PathInRepo   string
	DBPath       string
}

func (r *APIStoreLoader) Load(ctx context.Context, u *unstructured.Unstructured) error {
	crd, err := GetCRDFromUnstructured(u)
	if err != nil {
		return err
	}

	resctx, err := crdloader.LoadCRD(ctx, r.PathInRepo, r.DBPath, crd, nil, false)
	if err != nil {
		return err
	}

	// dont add internal gvks to the store
	if r.InternalGVKs.Has(resctx.ExternalGVK()) {
		return nil
	}

	if err := r.APIStore.Apply(resctx.GV(), resctx); err != nil {
		return err
	}
	return nil
}
