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
	"github.com/kform-dev/kform/pkg/fsys"
	"github.com/kform-dev/kform/pkg/pkgio"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

func (r *DevLoader) getReconcilerReader() pkgio.Reader[[]byte] {
	//abspath := filepath.Join(r.Path, *r.Cfg.ServerFlags.ReconcilerPath)
	abspath := filepath.Join(r.RepoPath, r.PathInRepo, "reconcilers")

	if !fsys.PathExists(abspath) {
		return nil
	}
	return GetFSReconcilerReader(abspath)
}

func (r *DevLoader) LoadReconcilers(ctx context.Context) error {
	reader := r.getReconcilerReader()
	if reader == nil {
		// reader nil, mean the path does not exist, whcich is ok
		return nil
	}
	datastore, err := reader.Read(ctx)
	if err != nil {
		return err
	}

	// The reconciler reader reads all the directories/files that match config.yaml
	// convetion is that this file contains the directory where the reconcilers
	// are located
	reconcilers := map[string]*choreov1alpha1.Reconciler{}
	var errs error
	datastore.List(func(k store.Key, b []byte) {
		// phase 1 we load all files named config.yaml -> this should give us reconcielr cnfigs
		basepath := filepath.Dir(k.Name)
		abspath := filepath.Join(r.RepoPath, r.PathInRepo, "reconcilers", basepath)

		if !fsys.PathExists(abspath) {
			errs = errors.Join(errs, fmt.Errorf("strange, reconciler path %s does not exists, while we just loaded the config.yaml", abspath))
			return
		}
		reader := GetFSReader(abspath)
		if reader == nil {
			return
		}
		datastore, err := reader.Read(ctx)
		if err != nil {
			errs = errors.Join(errs, err)
			return
		}
		reconcilerConfig := &choreov1alpha1.Reconciler{}
		if err := yaml.Unmarshal(b, reconcilerConfig); err != nil {
			errs = errors.Join(errs, fmt.Errorf("invalid reconciler %s, err: %v", k.Name, err))
			return
		}
		reconcilerConfig.Spec.Code = map[string]string{}
		reconcilerConfig.SetAnnotations(map[string]string{
			choreov1alpha1.ChoreoLoaderOriginKey: choreov1alpha1.FileLoaderAnnotation.String(),
		})
		reconcilers[reconcilerConfig.GetName()] = reconcilerConfig
		datastore.List(func(k store.Key, b []byte) {
			if k.Name == "config.yaml" {
				return
			}
			switch filepath.Ext(k.Name) {
			case ".tpl":
				if reconcilerConfig.Spec.Type != nil && *reconcilerConfig.Spec.Type != choreov1alpha1.SoftwardTechnologyType_GoTemplate {
					errs = errors.Join(errs, fmt.Errorf("a given reconciler %s can only have 1 sw technology type got %s and %s", reconcilerConfig.GetName(), (*reconcilerConfig.Spec.Type).String(), choreov1alpha1.SoftwardTechnologyType_GoTemplate.String()))
					return
				}
				reconcilerConfig.Spec.Type = ptr.To(choreov1alpha1.SoftwardTechnologyType_GoTemplate)
				if reconcilerConfig.Spec.Code == nil {
					reconcilerConfig.Spec.Code = map[string]string{}
				}
				templateName := strings.TrimPrefix(k.Name, reconcilerConfig.GetName()+".")
				reconcilerConfig.Spec.Code[templateName] = string(b)

			case ".jinja2":
				if reconcilerConfig.Spec.Type != nil && *reconcilerConfig.Spec.Type != choreov1alpha1.SoftwardTechnologyType_JinjaTemplate {
					errs = errors.Join(errs, fmt.Errorf("a given reconciler %s can only have 1 sw technology type got %s and %s", reconcilerConfig.GetName(), (*reconcilerConfig.Spec.Type).String(), choreov1alpha1.SoftwardTechnologyType_GoTemplate.String()))
					return
				}
				reconcilerConfig.Spec.Type = ptr.To(choreov1alpha1.SoftwardTechnologyType_JinjaTemplate)
				if reconcilerConfig.Spec.Code == nil {
					reconcilerConfig.Spec.Code = map[string]string{}
				}
				templateName := strings.TrimPrefix(k.Name, reconcilerConfig.GetName()+".")
				reconcilerConfig.Spec.Code[templateName] = string(b)

			case ".star":
				if reconcilerConfig.Spec.Type != nil {
					errs = errors.Join(errs, fmt.Errorf("a starlark reconciler %s can only have 1 code file", reconcilerConfig.GetName()))
					return
				}
				reconcilerConfig.Spec.Type = ptr.To(choreov1alpha1.SoftwardTechnologyType_Starlark)
				if reconcilerConfig.Spec.Code == nil {
					reconcilerConfig.Spec.Code = map[string]string{}
				}
				reconcilerConfig.Spec.Code["reconciler.star"] = string(b)
			}

			reconcilers[reconcilerConfig.GetName()] = reconcilerConfig
		})
	})

	for _, reconcilerConfig := range reconcilers {
		r.Reconcilers = append(r.Reconcilers, reconcilerConfig)
	}

	/*
		for _, reconcilerConfig := range reconcilers {
			if reconcilerConfig.Spec.Type != nil {
				//r.NewReconcilers.Insert(reconcilerConfig.Name)

				b, err := yaml.Marshal(reconcilerConfig)
				if err != nil {
					errs = errors.Join(errs, fmt.Errorf("cannot marshal library %s, err: %v", reconcilerConfig.Name, err))
					continue
				}

				fileName := filepath.Join(
					r.DstPath,
					*r.Cfg.ServerFlags.InputPath,
					fmt.Sprintf("%s.%s.%s.yaml",
						choreov1alpha1.SchemeGroupVersion.Group,
						strings.ToLower(choreov1alpha1.ReconcilerKind),
						reconcilerConfig.Name,
					))
				if err := os.WriteFile(fileName, b, 0644); err != nil {
					errs = errors.Join(errs, fmt.Errorf("cannot marshal library %s, err: %v", reconcilerConfig.Name, err))
				}
			}
		}
	*/
	return errs
}

// TBD if we need to do the clean
