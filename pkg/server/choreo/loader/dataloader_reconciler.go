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
	"strings"

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
	reconcilers := map[string]*choreov1alpha1.Reconciler{}
	var errm error
	datastore.List(func(k store.Key, b []byte) {
		if filepath.Ext(k.Name) == ".yaml" {
			reconcilerName := strings.TrimSuffix(k.Name, ".yaml")

			reconcilerConfig := &choreov1alpha1.Reconciler{}
			if err := yaml.Unmarshal(b, reconcilerConfig); err != nil {
				errm = errors.Join(errm, fmt.Errorf("invalid reconciler %s, err: %v", k.Name, err))
				return
			}
			reconcilerConfig.SetName(reconcilerName)
			reconcilerConfig.Spec.Code = map[string]string{}
			reconcilerConfig.SetAnnotations(map[string]string{
				choreov1alpha1.ChoreoLoaderOriginKey: choreov1alpha1.FileLoaderAnnotation.String(),
			})

			reconcilers[reconcilerName] = reconcilerConfig
		}
	})

	datastore.List(func(k store.Key, b []byte) {
		if filepath.Ext(k.Name) == ".yaml" {
			return // we already loaded the yaml files
		}
		reconcilerName, err := getReconcilerName(reconcilers, k.Name)
		if err != nil {
			errm = errors.Join(errm, err)
			return
		}
		reconcilerConfig := reconcilers[reconcilerName]

		switch filepath.Ext(k.Name) {
		case ".tpl":
			if reconcilerConfig.Spec.Type != nil && *reconcilerConfig.Spec.Type != choreov1alpha1.SoftwardTechnologyType_GoTemplate {
				errm = errors.Join(errm, fmt.Errorf("a given reconciler %s can only have 1 sw technology type got %s and %s", reconcilerName, (*reconcilerConfig.Spec.Type).String(), choreov1alpha1.SoftwardTechnologyType_GoTemplate.String()))
				return
			}
			reconcilerConfig.Spec.Type = ptr.To(choreov1alpha1.SoftwardTechnologyType_GoTemplate)
			if reconcilerConfig.Spec.Code == nil {
				reconcilerConfig.Spec.Code = map[string]string{}
			}
			templateName := strings.TrimPrefix(k.Name, reconcilerName+".")
			reconcilerConfig.Spec.Code[templateName] = string(b)

		case ".star":
			if reconcilerConfig.Spec.Type != nil {
				errm = errors.Join(errm, fmt.Errorf("a starlark reconciler %s can only have 1 code file", reconcilerName))
				return
			}
			reconcilerConfig.Spec.Type = ptr.To(choreov1alpha1.SoftwardTechnologyType_Starlark)
			if reconcilerConfig.Spec.Code == nil {
				reconcilerConfig.Spec.Code = map[string]string{}
			}
			reconcilerConfig.Spec.Code["reconciler.star"] = string(b)
		}

		reconcilers[reconcilerName] = reconcilerConfig
	})

	for _, reconcilerConfig := range reconcilers {
		if reconcilerConfig.Spec.Type != nil {
			r.NewReconcilers.Insert(reconcilerConfig.Name)

			obj, err := object.GetUnstructructered(reconcilerConfig)
			if err != nil {
				errm = errors.Join(errm, fmt.Errorf("cannot unmarshal %s, err: %v", reconcilerConfig.Name, err))
				continue
			}

			r.NewReconcilers.Insert(reconcilerConfig.Name)
			if err := r.Client.Apply(ctx, obj, &resourceclient.ApplyOptions{
				Branch:       r.Branch,
				FieldManager: ManagedFieldManagerInput,
				Origin:       ManagedFieldManagerInput,
			}); err != nil {
				errm = errors.Join(errm, fmt.Errorf("invalid reconciler %s, err: %v", reconcilerConfig.Name, err))
				continue
			}

		}
	}

	return errm
}

func getReconcilerName(reconcilers map[string]*choreov1alpha1.Reconciler, filename string) (string, error) {
	for reconcilerName := range reconcilers {
		if strings.HasPrefix(filename, reconcilerName) {
			return reconcilerName, nil
		}
	}
	return "", fmt.Errorf("no reconciler config found for %s", filename)
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
