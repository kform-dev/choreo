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

package crdloader

/*

import (
	"context"

	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/server/api"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func GetResourceCtxAPIExt(ctx context.Context, pathInRepo, dbpath string) (*api.ResourceContext, error) {
	datastore, err := GetAPIExtReader().Read(ctx)
	if err != nil {
		return nil, err
	}
	var b []byte
	datastore.List(func(k store.Key, r *yaml.RNode) {
		b = []byte(r.MustString())
	})

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(b, crd); err != nil {
		return nil, err
	}
	return LoadCRD(ctx, pathInRepo, dbpath, crd, nil)
}
*/
