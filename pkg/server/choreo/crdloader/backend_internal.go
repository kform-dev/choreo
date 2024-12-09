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

	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kuidio/kuid/apis/backend"
	"github.com/kuidio/kuid/apis/backend/as"
	asregister "github.com/kuidio/kuid/apis/backend/as/register"
	asv1alpha1 "github.com/kuidio/kuid/apis/backend/as/v1alpha1"
	"github.com/kuidio/kuid/apis/backend/genid"
	genidregister "github.com/kuidio/kuid/apis/backend/genid/register"
	genidv1alpha1 "github.com/kuidio/kuid/apis/backend/genid/v1alpha1"
	"github.com/kuidio/kuid/apis/backend/ipam"
	ipamregister "github.com/kuidio/kuid/apis/backend/ipam/register"
	ipamv1alpha1 "github.com/kuidio/kuid/apis/backend/ipam/v1alpha1"
	"github.com/kuidio/kuid/apis/backend/vlan"
	vlanregister "github.com/kuidio/kuid/apis/backend/vlan/register"
	vlanv1alpha1 "github.com/kuidio/kuid/apis/backend/vlan/v1alpha1"
	bebackend "github.com/kuidio/kuid/pkg/backend"
	"github.com/kuidio/kuid/pkg/registry/options"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type BackendConfig struct {
	Backend        bebackend.Backend
	IndexKind      string
	EntryKind      string
	ClaimKind      string
	EntryObjectFn  func(runtime.Unstructured) (backend.EntryObject, error)
	ClaimObjectFn  func(runtime.Unstructured) (backend.ClaimObject, error)
	IndexInvokerFn func(be bebackend.Backend) options.BackendInvoker
	ClaimInvokerFn func(be bebackend.Backend) options.BackendInvoker
	// derived
	IndexInvoker options.BackendInvoker
	ClaimInvoker options.BackendInvoker
}

func GetBackendConfig() map[string]*BackendConfig {
	backends := map[string]*BackendConfig{
		as.GroupName: {
			Backend:        asregister.NewBackend(),
			IndexKind:      asv1alpha1.ASIndexKind,
			EntryKind:      asv1alpha1.ASEntryKind,
			ClaimKind:      asv1alpha1.ASClaimKind,
			EntryObjectFn:  as.ASEntryFromUnstructured,
			ClaimObjectFn:  as.ASClaimFromUnstructured,
			IndexInvokerFn: as.NewChoreoIndexInvoker,
			ClaimInvokerFn: as.NewChoreoClaimInvoker,
		},
		genid.GroupName: {
			Backend:        genidregister.NewBackend(),
			IndexKind:      genidv1alpha1.GENIDIndexKind,
			EntryKind:      genidv1alpha1.GENIDEntryKind,
			ClaimKind:      genidv1alpha1.GENIDClaimKind,
			EntryObjectFn:  genid.GENIDEntryFromUnstructured,
			ClaimObjectFn:  genid.GENIDClaimFromUnstructured,
			IndexInvokerFn: genid.NewChoreoIndexInvoker,
			ClaimInvokerFn: genid.NewChoreoClaimInvoker,
		},
		ipam.GroupName: {
			Backend:   ipamregister.NewBackend(),
			IndexKind: ipamv1alpha1.IPIndexKind,
			EntryKind: ipamv1alpha1.IPEntryKind,
			ClaimKind: ipamv1alpha1.IPClaimKind,
			// not type based conversion functions needed since the type is known
			IndexInvokerFn: ipam.NewChoreoIndexInvoker,
			ClaimInvokerFn: ipam.NewChoreoClaimInvoker,
		},
		vlan.GroupName: {
			Backend:        vlanregister.NewBackend(),
			IndexKind:      vlanv1alpha1.VLANIndexKind,
			EntryKind:      vlanv1alpha1.VLANEntryKind,
			ClaimKind:      vlanv1alpha1.VLANClaimKind,
			EntryObjectFn:  vlan.VLANEntryFromUnstructured,
			ClaimObjectFn:  vlan.VLANClaimFromUnstructured,
			IndexInvokerFn: vlan.NewChoreoIndexInvoker,
			ClaimInvokerFn: vlan.NewChoreoClaimInvoker,
		},
	}
	for _, backendConfig := range backends {
		setupBackend(backendConfig)
	}
	return backends
}

// TODO need to add a conversion
func setupBackend(backendConfig *BackendConfig) {
	backendConfig.IndexInvoker = backendConfig.IndexInvokerFn(backendConfig.Backend)
	backendConfig.ClaimInvoker = backendConfig.ClaimInvokerFn(backendConfig.Backend)
}

func AddStorage(backends map[string]*BackendConfig, apiStore *api.APIStore) error {
	var errm error
	for group, backendConfig := range backends {
		entryStorage, err := apiStore.GetStorage(schema.GroupKind{Group: group, Kind: backendConfig.EntryKind})
		if err != nil {
			errm = errors.Join(errm, err)
		}
		claimStorage, err := apiStore.GetStorage(schema.GroupKind{Group: group, Kind: backendConfig.ClaimKind})
		if err != nil {
			errm = errors.Join(errm, err)
		}

		if group == ipam.GroupName {
			backendConfig.Backend.AddStorageInterfaces(NewChoreoIPAMBackendstorage(backendConfig.IndexKind, entryStorage, claimStorage))
		} else {
			backendConfig.Backend.AddStorageInterfaces(NewChoreoGenericBackendstorage(backendConfig.IndexKind, entryStorage, claimStorage, backendConfig.EntryObjectFn, backendConfig.ClaimObjectFn))
		}
	}
	return errm
}
