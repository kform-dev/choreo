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
	"fmt"
	"sync"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

type inFormers struct {
	m         sync.RWMutex
	informers map[schema.GroupVersionKind]Informer

	cancel func()
}

func newInformers(
	client resourceclient.Client,
	gvks sets.Set[schema.GroupVersionKind],
	branchName string,
) *inFormers {
	informers := &inFormers{
		informers: make(map[schema.GroupVersionKind]Informer, 0),
	}
	for _, gvk := range gvks.UnsortedList() {
		informers.add(gvk, NewInformer(client, gvk, branchName))
	}
	return informers
}

func (r *inFormers) add(gvk schema.GroupVersionKind, i Informer) {
	r.m.Lock()
	defer r.m.Unlock()
	r.informers[gvk] = i
}

func (r *inFormers) get(gvk schema.GroupVersionKind) (Informer, error) {
	r.m.Lock()
	defer r.m.Unlock()
	i, ok := r.informers[gvk]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return i, nil
}

func (r *inFormers) list() []Informer {
	r.m.RLock()
	defer r.m.RUnlock()
	informers := make([]Informer, 0, len(r.informers))
	for _, informer := range r.informers {
		informers = append(informers, informer)
	}
	return informers
}

func (r *inFormers) addEventHandler(gvk schema.GroupVersionKind, reconcilerName string, handler OnChangeFn) error {
	i, err := r.get(gvk)
	if err != nil {
		return fmt.Errorf("gvk %s not initialized", gvk.String())
	}
	i.addEventHandler(reconcilerName, handler)
	return nil
}

func (r *inFormers) start(ctx context.Context) {
	log := log.FromContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	defer cancel()

	for _, informer := range r.list() {
		go func() {
			informer.start(ctx)
		}()
	}

	<-ctx.Done()
	log.Debug("informers stopped...")
}

func (r *inFormers) stop() {
	if r.cancel != nil {
		r.cancel()
	}
}
