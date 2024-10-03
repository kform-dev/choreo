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

import (
	"errors"

	asv1alpha1 "github.com/kform-dev/choreo/apis/kuid/backend/as/v1alpha1"
	genidv1alpha1 "github.com/kform-dev/choreo/apis/kuid/backend/genid/v1alpha1"
	ipamv1alpha1 "github.com/kform-dev/choreo/apis/kuid/backend/ipam/v1alpha1"
	vlanv1alpha1 "github.com/kform-dev/choreo/apis/kuid/backend/vlan/v1alpha1"
	"github.com/kform-dev/choreo/pkg/backend"
	"github.com/kform-dev/choreo/pkg/backend/ipam"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/apiserver/registry"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type BackendConfig struct {
	Backend         backend.Backend
	ClaimPreparator registry.APIPrepator
	IndexPreparator registry.APIPrepator
	IndexKind       string
	EntryKind       string
	ClaimKind       string
}

func GetBackendConfig() map[schema.GroupVersion]*BackendConfig {
	backends := map[schema.GroupVersion]*BackendConfig{
		asv1alpha1.SchemeGroupVersion:    {Backend: asv1alpha1.NewBackend(), IndexKind: asv1alpha1.ASIndexKind, EntryKind: asv1alpha1.ASEntryKind, ClaimKind: asv1alpha1.ASClaimKind},
		vlanv1alpha1.SchemeGroupVersion:  {Backend: vlanv1alpha1.NewBackend(), IndexKind: vlanv1alpha1.VLANIndexKind, EntryKind: vlanv1alpha1.VLANEntryKind, ClaimKind: vlanv1alpha1.VLANClaimKind},
		ipamv1alpha1.SchemeGroupVersion:  {Backend: ipam.New(), IndexKind: ipamv1alpha1.IPIndexKind, EntryKind: ipamv1alpha1.IPEntryKind, ClaimKind: ipamv1alpha1.IPClaimKind},
		genidv1alpha1.SchemeGroupVersion: {Backend: genidv1alpha1.NewBackend(), IndexKind: genidv1alpha1.GENIDIndexKind, EntryKind: genidv1alpha1.GENIDEntryKind, ClaimKind: genidv1alpha1.GENIDClaimKind},
	}
	for _, backendConfig := range backends {
		setupBackend(backendConfig)
	}
	return backends
}

func setupBackend(backendConfig *BackendConfig) {
	backendConfig.IndexPreparator = backend.NewIndexPreparator(backendConfig.Backend)
	backendConfig.ClaimPreparator = backend.NewClaimPreparator(backendConfig.Backend)
}

func AddStorage(backends map[schema.GroupVersion]*BackendConfig, apiStore *api.APIStore) error {
	var errm error
	for gv, backendConfig := range backends {
		entryStorage, err := apiStore.GetStorage(gv.WithKind(backendConfig.EntryKind))
		if err != nil {
			errm = errors.Join(errm, err)
		}
		claimStorage, err := apiStore.GetStorage(gv.WithKind(backendConfig.ClaimKind))
		if err != nil {
			errm = errors.Join(errm, err)
		}
		backendConfig.Backend.AddStorage(entryStorage, claimStorage)
	}
	return errm
}
