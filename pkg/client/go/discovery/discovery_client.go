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

package discovery

import (
	"context"

	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
)

type DiscoveryInterface interface {
	APIResources(ctx context.Context, branch string) (*choreov1alpha1.APIResources, error)
	Close() error
	Watch(context.Context, *discoverypb.Watch_Request) chan *discoverypb.Watch_Response
}

// CachedDiscoveryInterface is a DiscoveryInterface with cache invalidation and freshness.
// Note that If the ServerResourcesForGroupVersion method returns a cache miss
// error, the user needs to explicitly call Invalidate to clear the cache,
// otherwise the same cache miss error will be returned next time.
type CachedDiscoveryInterface interface {
	DiscoveryInterface
	// Invalidate enforces that no cached data that is older than the current time
	// is used.
	Invalidate()
}
