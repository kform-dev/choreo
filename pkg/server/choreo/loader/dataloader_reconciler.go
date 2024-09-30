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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

func (r *DataLoader) loadReconcilers(ctx context.Context) error {
	loader := &ReconcilerLoader{
		Client:         r.Client,
		Branch:         r.Branch,
		NewReconcilers: sets.New[string](),
	}
	if err := loader.Load(ctx, r.getReconcilerReader(r.RepoPth, r.PathInRepo, r.Flags)); err != nil {
		return err
	}
	if err := loader.Clean(ctx); err != nil {
		return err
	}
	return nil
}

func (r *DataLoader) getReconcilerReader(repoPath, pathInRepo string, flags *genericclioptions.ConfigFlags) pkgio.Reader[[]byte] {
	abspath := filepath.Join(repoPath, pathInRepo, *flags.ReconcilerPath)

	if !fsys.PathExists(abspath) {
		return nil
	}
	return GetFSReader(abspath)
}

type ReconcilerLoader struct {
	Client         resourceclient.Client
	Branch         string
	NewReconcilers sets.Set[string]
}

func (r *ReconcilerLoader) Load(ctx context.Context, reader pkgio.Reader[[]byte]) error {
	if reader == nil {
		// reader nil, mean the path does not exist, whcich is ok
		return nil
	}
	datastore, err := reader.Read(ctx)
	if err != nil {
		return err
	}

	// For starlark we have 2 files fro which the name before the extension match
	// As such we keep a structure with the name, extension, byte
	files := map[string]map[string][]byte{}
	datastore.List(func(k store.Key, b []byte) {
		path := k.Name
		base := filepath.Base(path)
		ext := filepath.Ext(path)
		name := base[:len(base)-len(ext)]

		if _, exists := files[name]; !exists {
			files[name] = map[string][]byte{}
		}
		files[name][ext] = b
	})

	var errm error
	for name, data := range files {
		reconcilerType := getReconcilerType(data)
		switch reconcilerType {
		case choreov1alpha1.SoftwardTechnologyType_Starlark:
			reconcilerConfig := &choreov1alpha1.Reconciler{}
			if err := yaml.Unmarshal(data[".yaml"], reconcilerConfig); err != nil {
				errm = errors.Join(errm, fmt.Errorf("invalid reconciler %s, err: %v", name, err))
				continue
			}
			reconcilerConfig.SetName(name)
			reconcilerConfig.Spec.Type = ptr.To(choreov1alpha1.SoftwardTechnologyType_Starlark)
			reconcilerConfig.Spec.Code = map[string]string{
				"reconciler.star": string(data[".star"]),
			}

			reconcilerConfig.SetAnnotations(map[string]string{
				choreov1alpha1.ChoreoLoaderOriginKey: choreov1alpha1.FileLoaderAnnotation.String(),
			})

			obj, err := object.GetUnstructructered(reconcilerConfig)
			if err != nil {
				errm = errors.Join(errm, fmt.Errorf("cannot unmarshal %s, err: %v", name, err))
				continue
			}

			r.NewReconcilers.Insert(name)
			if err := r.Client.Apply(ctx, obj, &resourceclient.ApplyOptions{
				Branch:       r.Branch,
				FieldManager: ManagedFieldManagerInput,
				Origin:       ManagedFieldManagerInput,
			}); err != nil {
				errm = errors.Join(errm, fmt.Errorf("invalid reconciler %s, err: %v", name, err))
				continue
			}

		default:
			errm = errors.Join(errm, fmt.Errorf("invalid reconciler %s", name))
		}
	}
	return errm
}

func getReconcilerType(data map[string][]byte) choreov1alpha1.SoftwardTechnologyType {
	if _, exists := data[".star"]; exists {
		return choreov1alpha1.SoftwardTechnologyType_Starlark
	}
	return choreov1alpha1.SoftwardTechnologyType_Kform
}

func (r *ReconcilerLoader) Clean(ctx context.Context) error {
	ul := &unstructured.UnstructuredList{}
	ul.SetGroupVersionKind(choreov1alpha1.SchemeGroupVersion.WithKind(choreov1alpha1.ReconcilerKind))
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

			if !r.NewReconcilers.Has(u.GetName()) {
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
