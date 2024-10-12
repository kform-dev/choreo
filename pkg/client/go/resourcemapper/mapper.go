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

package resourcemapper

import (
	"context"
	"fmt"

	"github.com/kform-dev/choreo/pkg/client/go/discovery"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type Mapper interface {
	KindFor(ctx context.Context, resource schema.GroupResource, proxy types.NamespacedName, branch string) (schema.GroupVersionKind, error)
}

// NewDiscoveryRESTMapper returns a PriorityRESTMapper based on the discovered
// groups and resources passed in.
func NewMapper(d discovery.DiscoveryInterface) Mapper {
	return &mapper{
		d: d,
	}
}

type mapper struct {
	d discovery.DiscoveryInterface
}

func (r *mapper) KindFor(ctx context.Context, resource schema.GroupResource, proxy types.NamespacedName, branch string) (schema.GroupVersionKind, error) {
	apiResources, err := r.d.APIResources(ctx, proxy, branch)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	if apiResources == nil {
		return schema.GroupVersionKind{}, fmt.Errorf("no apiResource received")
	}

	for _, apiresource := range apiResources {
		if apiresource.Resource == resource.Resource && apiresource.Group == resource.Group {
			return schema.GroupVersionKind{
				Group:   apiresource.Group,
				Version: apiresource.Version,
				Kind:    apiresource.Kind,
			}, nil
		}
	}
	return schema.GroupVersionKind{}, fmt.Errorf("cannot find a resource mapping for %s", resource.String())
}
