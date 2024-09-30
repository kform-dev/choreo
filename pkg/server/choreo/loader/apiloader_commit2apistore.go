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

package loader

import (
	"context"
	"errors"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/henderiw/logger/log"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/api"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

// This loader is reading committed apis from the commit
type APILoaderCommit2APIStore struct {
	Client       resourceclient.Client // client to read from the db repo since this reader reads the api from commit
	APIStore     *api.APIStore
	InternalGVKs sets.Set[schema.GroupVersionKind]
	PathInRepo   string
	DBPath       string
}

func (r *APILoaderCommit2APIStore) Load(ctx context.Context, commit *object.Commit) error {
	log := log.FromContext(ctx)

	var errm error
	ul := &unstructured.UnstructuredList{}
	ul.SetAPIVersion(apiextensionsv1.SchemeGroupVersion.Identifier())
	ul.SetKind("CustomResourceDefinition")
	if err := r.Client.List(ctx, ul, &resourceclient.ListOptions{
		Commit:            commit,
		ShowManagedFields: false,
		ExprSelector:      &resourcepb.ExpressionSelector{},
	}); err != nil {
		log.Error("cannot load ")
		return nil
	}

	for _, u := range ul.Items {
		// don't add internal apis

		apistoreloader := &APIStoreLoader{
			APIStore:     r.APIStore,
			InternalGVKs: r.InternalGVKs,
			PathInRepo:   r.PathInRepo,
			DBPath:       r.DBPath,
		}

		setAnnotations(&u, map[string]string{
			choreov1alpha1.ChoreoLoaderOriginKey: choreov1alpha1.FileLoaderAnnotation.String(),
		})

		if err := apistoreloader.Load(ctx, &u); err != nil {
			errm = errors.Join(errm, err)
			continue
		}
	}

	return errm
}
