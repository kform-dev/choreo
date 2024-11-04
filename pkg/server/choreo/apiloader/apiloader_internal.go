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
	"errors"
	"fmt"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/choreo/crdloader"
	"github.com/kform-dev/kform/pkg/pkgio"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type APILoaderInternal struct {
	APIStore   *api.APIStore
	Cfg        *genericclioptions.ChoreoConfig
	DBPath     string
	PathInRepo string
}

func (r *APILoaderInternal) Load(ctx context.Context) error {
	var errm error
	// load internal choreo apis
	if err := r.loadAPIs(ctx, crdloader.GetAPIExtReader(), nil, true); err != nil {
		errm = errors.Join(errm, fmt.Errorf("cannot load api extension apis, err: %v", err))
	}
	// load internal choreo apis
	if err := r.loadAPIs(ctx, crdloader.GetEmbeddedAPIReader(), nil, true); err != nil {
		errm = errors.Join(errm, fmt.Errorf("cannot load embedded apis, err: %v", err))
	}
	// load internal but not choreoAPIs
	if r.Cfg.ServerFlags.InternalReconcilers != nil && *r.Cfg.ServerFlags.InternalReconcilers {
		backends := crdloader.GetBackendConfig()
		if err := r.loadAPIs(ctx, crdloader.GetInternalAPIReader(), backends, false); err != nil {
			errm = errors.Join(errm, fmt.Errorf("cannot load internal apis, err: %v", err))
		}
		if err := crdloader.AddStorage(backends, r.APIStore); err != nil {
			errm = errors.Join(errm, fmt.Errorf("cannot add storage to internal apis, err: %v", err))
		}
	}
	return errm
}

func (r *APILoaderInternal) loadAPIs(ctx context.Context, reader pkgio.Reader[*yaml.RNode], backends map[string]*crdloader.BackendConfig, choreoAPI bool) error {
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
	var errm error
	datastore.List(func(k store.Key, rn *yaml.RNode) {
		if rn.GetApiVersion() == apiextensionsv1.SchemeGroupVersion.Identifier() {
			crd := &apiextensionsv1.CustomResourceDefinition{}
			if err := yaml.Unmarshal([]byte(rn.MustString()), crd); err != nil {
				errm = errors.Join(errm, err)
				return
			}
			// TODO handle internal apis

			resctx, err := crdloader.LoadCRD(ctx, r.PathInRepo, r.DBPath, crd, backends, choreoAPI)
			if err != nil {
				errm = errors.Join(errm, err)
				return
			}
			if err := r.APIStore.Apply(resctx.GV(), resctx); err != nil {
				errm = errors.Join(errm, err)
				return
			}
		}
	})
	return errm
}
