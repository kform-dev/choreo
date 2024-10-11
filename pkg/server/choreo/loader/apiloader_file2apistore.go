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
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/choreo/crdloader"
	uobject "github.com/kform-dev/choreo/pkg/util/object"
	"github.com/kform-dev/kform/pkg/pkgio"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Loader to load files from fs and apply them to api and apistore
type APILoaderFile2APIStoreAndAPI struct {
	Flags        *genericclioptions.ConfigFlags
	Client       resourceclient.Client
	APIStore     *api.APIStore
	Branch       string // not relevant in commit case
	InternalGVKs sets.Set[schema.GroupVersionKind]
	RepoPath     string // not relevant in commit case
	PathInRepo   string
	DBPath       string
}

func (r *APILoaderFile2APIStoreAndAPI) LoadFromCommit(ctx context.Context, commit *object.Commit) error {
	//log := log.FromContext(ctx)
	var errm error
	fmt.Println("load from commit", *r.Flags.CRDPath, commit.Hash.String())
	if err := r.loadAPIs(ctx, crdloader.GetCommitFileAPICRDReader(
		filepath.Join(r.PathInRepo, *r.Flags.CRDPath), commit)); err != nil {
		errm = errors.Join(errm, fmt.Errorf("cannot load api file in repo, err: %v", err))
	}
	return errm
}

func (r *APILoaderFile2APIStoreAndAPI) Load(ctx context.Context) error {
	var errm error
	/*
		embeddedAnnotations := map[string]string{
			choreov1alpha1.ChoreoAPIEmbeddedKey: "true",
		}

		if err := r.loadAPIs(ctx, crdloader.GetAPIExtReader(), embeddedAnnotations); err != nil {
			errm = errors.Join(errm, fmt.Errorf("cannot load api extension apis, err: %v", err))
		}
		if err := r.loadAPIs(ctx, crdloader.GetEmbeddedAPIReader(), embeddedAnnotations); err != nil {
			errm = errors.Join(errm, fmt.Errorf("cannot load embedded apis, err: %v", err))
		}
		if r.Flags.InternalReconcilers != nil && *r.Flags.InternalReconcilers {
			if err := r.loadAPIs(ctx, crdloader.GetInternalAPIReader(), embeddedAnnotations); err != nil {
				errm = errors.Join(errm, fmt.Errorf("cannot load internal apis, err: %v", err))
			}
		}
	*/

	abspath := filepath.Join(r.RepoPath, r.PathInRepo, *r.Flags.CRDPath)
	if err := r.loadAPIs(ctx, crdloader.GetFileAPICRDReader(abspath)); err != nil {
		errm = errors.Join(errm, fmt.Errorf("cannot load api file in repo, err: %v", err))
	}
	return errm
}

func (r *APILoaderFile2APIStoreAndAPI) loadAPIs(ctx context.Context, reader pkgio.Reader[*yaml.RNode]) error {
	log := log.FromContext(ctx)
	if reader == nil {
		// a nil reader means the path does not exist
		return nil
	}
	datastore, err := reader.Read(ctx)
	if err != nil {
		log.Error("cnanot read datastore", "err", err)
		return err
	}
	apistoreloader := &APIStoreLoader{
		APIStore:     r.APIStore,
		InternalGVKs: r.InternalGVKs,
		PathInRepo:   r.PathInRepo,
		DBPath:       r.DBPath,
	}
	var errm error
	datastore.List(func(k store.Key, rn *yaml.RNode) {
		u, err := uobject.GetUnstructuredContent([]byte(rn.MustString()), uobject.ContentTypeYAML)
		if err != nil {
			errm = errors.Join(errm, err)
			return
		}

		if err := apistoreloader.Load(ctx, u); err != nil {
			errm = errors.Join(errm, err)
			return
		}

		// load data to apiserver
		setAnnotations(u, map[string]string{
			choreov1alpha1.ChoreoLoaderOriginKey: choreov1alpha1.FileLoaderAnnotation.String(),
		})

		if err := r.Client.Apply(ctx, u, &resourceclient.ApplyOptions{
			Branch:       r.Branch,
			FieldManager: ManagedFieldManagerInput,
		}); err != nil {
			errm = errors.Join(errm, err)
			return
		}
	})

	return errm
}

func setAnnotations(u *unstructured.Unstructured, annotations map[string]string) {
	a := u.GetAnnotations()
	if len(a) == 0 {
		a = map[string]string{}
	}
	for k, v := range annotations {
		a[k] = v
	}
	u.SetAnnotations(a)
}
