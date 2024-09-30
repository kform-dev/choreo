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

package client

import (
	"context"

	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/client/go/discovery"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type discoveryClient struct {
	client DiscoveryClientInterface
}

func NewDiscoveryClient(config *config.Config) (discovery.DiscoveryInterface, error) {
	client, err := New(config)
	if err != nil {
		return nil, err
	}
	return &discoveryClient{
		client: client,
	}, nil
}

func (r *discoveryClient) Close() error {
	return r.client.Close()
}

func (r *discoveryClient) APIResources(ctx context.Context, branchName string) (*choreov1alpha1.APIResources, error) {
	rsp, err := r.client.Get(ctx, &discoverypb.Get_Request{
		Branch: branchName,
	})
	if err != nil {
		return nil, err
	}
	apiResources := make([]*choreov1alpha1.APIResourceGroup, len(rsp.Apiresources))
	for i, apiResource := range rsp.Apiresources {
		apiResources[i] = &choreov1alpha1.APIResourceGroup{
			Resource:   apiResource.Resource,
			Group:      apiResource.Group,
			Version:    apiResource.Version,
			Kind:       apiResource.Kind,
			ListKind:   apiResource.ListKind,
			Namespaced: apiResource.Namespaced,
			Categories: apiResource.Categories,
		}
	}
	return &choreov1alpha1.APIResources{
		TypeMeta: metav1.TypeMeta{
			APIVersion: choreov1alpha1.SchemeGroupVersion.Identifier(),
			Kind:       choreov1alpha1.APIResourcesKind,
		},
		Spec: choreov1alpha1.APIResourcesSpec{
			Groups: apiResources,
		},
	}, nil
}

func (r *discoveryClient) Watch(ctx context.Context, in *discoverypb.Watch_Request) chan *discoverypb.Watch_Response {
	return r.client.Watch(ctx, in)
}
