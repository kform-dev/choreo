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

	"github.com/henderiw/store"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/util/object"
	"github.com/kform-dev/kform/pkg/fsys"
	"github.com/kform-dev/kform/pkg/pkgio"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (r *DataLoader) loadLibraries(ctx context.Context) error {
	loader := &LibraryLoader{
		Client:       r.Client,
		Branch:       r.Branch,
		NewLibraries: sets.New[string](),
	}
	if err := loader.Load(ctx, r.getLibraryReader(r.RepoPth, r.PathInRepo, r.Flags)); err != nil {
		return err
	}
	if err := loader.Clean(ctx); err != nil {
		return err
	}
	return nil
}

func (r *DataLoader) getLibraryReader(repoPath, pathInRepo string, flags *genericclioptions.ConfigFlags) pkgio.Reader[[]byte] {
	abspath := filepath.Join(repoPath, pathInRepo, *flags.LibraryPath)

	if !fsys.PathExists(abspath) {
		return nil
	}
	return GetFSReader(abspath)
}

type LibraryLoader struct {
	Client       resourceclient.Client
	Branch       string
	NewLibraries sets.Set[string]
}

func (r *LibraryLoader) Load(ctx context.Context, reader pkgio.Reader[[]byte]) error {
	if reader == nil {
		// reader nil, mean the path does not exist, whcich is ok
		return nil
	}
	datastore, err := reader.Read(ctx)
	if err != nil {
		return err
	}

	var errm error
	datastore.List(func(k store.Key, b []byte) {
		library := &choreov1alpha1.Library{
			TypeMeta: metav1.TypeMeta{
				APIVersion: choreov1alpha1.SchemeGroupVersion.Identifier(),
				Kind:       choreov1alpha1.LibraryKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: k.Name,
			},
			Spec: choreov1alpha1.LibrarySpec{
				Type: choreov1alpha1.SoftwardTechnologyType_Starlark,
				Code: string(b),
			},
		}

		library.SetAnnotations(map[string]string{
			choreov1alpha1.ChoreoLoaderOriginKey: choreov1alpha1.FileLoaderAnnotation.String(),
		})

		r.NewLibraries.Insert(k.Name)
		obj, err := object.GetUnstructructered(library)
		if err != nil {
			errm = errors.Join(errm, fmt.Errorf("cannot unmarshal %s, err: %v", k.Name, err))
			return
		}

		if err := r.Client.Apply(ctx, obj, &resourceclient.ApplyOptions{
			Branch:       r.Branch,
			FieldManager: ManagedFieldManagerInput,
		}); err != nil {
			errm = errors.Join(errm, fmt.Errorf("invalid library %s, err: %v", k.Name, err))
			return
		}
	})
	return errm
}

func (r *LibraryLoader) Clean(ctx context.Context) error {
	ul := &unstructured.UnstructuredList{}
	ul.SetGroupVersionKind(choreov1alpha1.SchemeGroupVersion.WithKind(choreov1alpha1.LibraryKind))
	if err := r.Client.List(ctx, ul, &resourceclient.ListOptions{
		ExprSelector:      &resourcepb.ExpressionSelector{},
		ShowManagedFields: true,
		Branch:            r.Branch,
	}); err != nil {
		return err
	}
	var errm error
	for _, u := range ul.Items {

		if len(u.GetAnnotations()) != 0 &&
			u.GetAnnotations()[choreov1alpha1.ChoreoLoaderOriginKey] == choreov1alpha1.FileLoaderAnnotation.String() &&
			object.IsManagedBy(u.GetManagedFields(), ManagedFieldManagerInput) {
			if !r.NewLibraries.Has(u.GetName()) {
				if err := r.Client.Delete(ctx, &u, &resourceclient.DeleteOptions{
					Branch: r.Branch,
				}); err != nil {
					errm = errors.Join(errm, err)
				}
			}
		}
	}
	return errm
}
