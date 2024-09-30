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

package informers

import (
	"context"

	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

type InformerFactory interface {
	AddEventHandler(gvk schema.GroupVersionKind, reconcilerName string, handler OnChangeFn) error
	Start(ctx context.Context)
}

func NewInformerFactory(client resourceclient.Client, gvks sets.Set[schema.GroupVersionKind], branchName string) InformerFactory {
	return &informerFactory{
		client:    client,
		informers: newInformers(client, gvks, branchName),
	}
}

type informerFactory struct {
	client    resourceclient.Client
	informers *inFormers
}

func (r *informerFactory) AddEventHandler(gvk schema.GroupVersionKind, reconcilerName string, handler OnChangeFn) error {
	return r.informers.addEventHandler(gvk, reconcilerName, handler)
}

func (r *informerFactory) Start(ctx context.Context) {
	r.informers.start(ctx)
}

func (r *informerFactory) Stop(ctx context.Context) {
	r.informers.stop()
}
