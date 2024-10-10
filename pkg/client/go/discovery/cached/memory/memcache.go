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

package memory

import (
	"context"
	"sync"

	"github.com/kform-dev/choreo/pkg/client/go/discovery"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
)

type memCacheClient struct {
	delegate discovery.DiscoveryInterface

	m            sync.RWMutex
	apiResources []*discoverypb.APIResource
	cacheValid   bool
}

// NewMemCacheClient creates a new CachedDiscoveryInterface which caches
// discovery information in memory and will stay up-to-date if Invalidate is
// called with regularity.
//
// NOTE: The client will NOT resort to live lookups on cache misses.
func NewMemCacheDiscoveryClient(delegate discovery.DiscoveryInterface) discovery.CachedDiscoveryInterface {
	return &memCacheClient{
		delegate: delegate,
	}
}

func (r *memCacheClient) Close() error {
	return r.delegate.Close()
}

func (r *memCacheClient) Invalidate() {
	r.m.Lock()
	defer r.m.Unlock()
	r.cacheValid = false
	r.apiResources = nil
	if d, ok := r.delegate.(discovery.CachedDiscoveryInterface); ok {
		d.Invalidate()
	}
}

func (r *memCacheClient) APIResources(ctx context.Context, branch string) ([]*discoverypb.APIResource, error) {
	r.m.Lock()
	defer r.m.Unlock()

	if !r.cacheValid {
		if err := r.refreshLocked(ctx, branch); err != nil {
			return nil, err
		}
	}
	return r.apiResources, nil
}

func (r *memCacheClient) Watch(ctx context.Context, req *discoverypb.Watch_Request) chan *discoverypb.Watch_Response {
	return r.delegate.Watch(ctx, req)
}

// refreshLocked refreshes the state of cache. The caller must hold d.lock for
// writing.
func (r *memCacheClient) refreshLocked(ctx context.Context, branch string) error {
	apiresources, err := r.delegate.APIResources(ctx, branch)
	if err != nil {
		return err
	}
	r.apiResources = apiresources
	r.cacheValid = true
	return nil
}
