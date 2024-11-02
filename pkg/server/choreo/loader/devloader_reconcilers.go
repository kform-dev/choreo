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
	"os"
	"path/filepath"
	"strings"

	"github.com/henderiw/store"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/kform/pkg/fsys"
	"github.com/kform-dev/kform/pkg/pkgio"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

func (r *DevLoader) getReconcilerReader() pkgio.Reader[[]byte] {
	abspath := filepath.Join(r.Path, *r.Cfg.ServerFlags.ReconcilerPath)

	if !fsys.PathExists(abspath) {
		return nil
	}
	return GetFSReader(abspath)
}

func (r *DevLoader) loadReconcilers(ctx context.Context) error {
	reader := r.getReconcilerReader()
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

		case ".jinja2":
			if reconcilerConfig.Spec.Type != nil && *reconcilerConfig.Spec.Type != choreov1alpha1.SoftwardTechnologyType_JinjaTemplate {
				errm = errors.Join(errm, fmt.Errorf("a given reconciler %s can only have 1 sw technology type got %s and %s", reconcilerName, (*reconcilerConfig.Spec.Type).String(), choreov1alpha1.SoftwardTechnologyType_GoTemplate.String()))
				return
			}
			reconcilerConfig.Spec.Type = ptr.To(choreov1alpha1.SoftwardTechnologyType_JinjaTemplate)
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
			//r.NewReconcilers.Insert(reconcilerConfig.Name)

			b, err := yaml.Marshal(reconcilerConfig)
			if err != nil {
				errm = errors.Join(errm, fmt.Errorf("cannot marshal library %s, err: %v", reconcilerConfig.Name, err))
			}

			fileName := filepath.Join(
				r.Path,
				*r.Cfg.ServerFlags.InputPath,
				fmt.Sprintf("%s.%s.%s.yaml",
					choreov1alpha1.SchemeGroupVersion.Group,
					strings.ToLower(choreov1alpha1.ReconcilerKind),
					reconcilerConfig.Name,
				))
			if err := os.WriteFile(fileName, b, 0644); err != nil {
				errm = errors.Join(errm, fmt.Errorf("cannot marshal library %s, err: %v", reconcilerConfig.Name, err))
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

// TBD if we need to do the clean
