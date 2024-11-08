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
	"path/filepath"
	"strings"

	"github.com/henderiw/store"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/kform/pkg/fsys"
	"github.com/kform-dev/kform/pkg/pkgio"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *DevLoader) getLibraryReader() pkgio.Reader[[]byte] {
	//abspath := filepath.Join(r.Path, *r.Cfg.ServerFlags.LibraryPath)
	abspath := filepath.Join(r.RepoPath, r.PathInRepo, *r.Cfg.ServerFlags.CRDPath)

	if !fsys.PathExists(abspath) {
		return nil
	}
	return GetFSStarReader(abspath)
}

func (r *DevLoader) LoadLibraries(ctx context.Context) error {
	reader := r.getLibraryReader()
	if reader == nil {
		// reader nil, mean the path does not exist, whcich is ok
		return nil
	}
	datastore, err := reader.Read(ctx)
	if err != nil {
		return err
	}

	var errs error
	datastore.List(func(k store.Key, b []byte) {

		name := strings.ReplaceAll(k.Name, "_", ".")

		library := &choreov1alpha1.Library{
			TypeMeta: metav1.TypeMeta{
				APIVersion: choreov1alpha1.SchemeGroupVersion.Identifier(),
				Kind:       choreov1alpha1.LibraryKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: choreov1alpha1.LibrarySpec{
				Type: choreov1alpha1.SoftwardTechnologyType_Starlark,
				Code: string(b),
			},
		}

		library.SetAnnotations(map[string]string{
			choreov1alpha1.ChoreoLoaderOriginKey: choreov1alpha1.FileLoaderAnnotation.String(),
		})

		r.Libraries = append(r.Libraries, library)

		// this is to see if we need to do cleanup
		//r.NewLibraries.Insert(k.Name)

		/*
			b, err := yaml.Marshal(library)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("cannot marshal library %s, err: %v", k.Name, err))
				return
			}

			fileName := filepath.Join(
				r.DstPath,
				*r.Cfg.ServerFlags.InputPath,
				fmt.Sprintf("%s.%s.%s.yaml",
					choreov1alpha1.SchemeGroupVersion.Group,
					strings.ToLower(choreov1alpha1.LibraryKind),
					name,
				))
			if err := os.WriteFile(fileName, b, 0644); err != nil {
				errs = errors.Join(errs, fmt.Errorf("cannot marshal library %s, err: %v", k.Name, err))
			}
		*/
		/*

			obj := map[string]any{}
			if err := yaml.Unmarshal(b, &obj); err != nil {
				errs = errors.Join(errs, fmt.Errorf("cannot marshal library %s, err: %v", k.Name, err))
				return
			}
			u := &unstructured.Unstructured{
				Object: obj,
			}

			if err := r.Client.Apply(ctx, u, &resourceclient.ApplyOptions{
				Branch:       r.Branch,
				FieldManager: ManagedFieldManagerInput,
			}); err != nil {
				errs = errors.Join(errs, err)
				return
			}
		*/
	})
	return errs
}

// TBD if we need this
/*
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
*/
