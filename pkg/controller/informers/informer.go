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

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/util/object"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type OnChangeFn func(ctx context.Context, eventType resourcepb.Watch_EventType, obj runtime.Unstructured) bool

type Informer interface {
	start(ctx context.Context)
	addEventHandler(reconcilerName string, handler OnChangeFn)
}

func NewInformer(client resourceclient.Client, gvk schema.GroupVersionKind, branchName string) Informer {
	return &informer{
		client:        client,
		gvk:           gvk,
		eventHandlers: newEventHandlers(),
		branch:        branchName,
	}
}

type informer struct {
	client        resourceclient.Client
	gvk           schema.GroupVersionKind
	eventHandlers *eventhandlers
	branch        string
}

func (r *informer) addEventHandler(reconcilerName string, handler OnChangeFn) {
	r.eventHandlers.add(reconcilerName, handler)
}

func (r *informer) start(ctx context.Context) {
	log := log.FromContext(ctx)
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(r.gvk)

	rspCh := r.client.Watch(ctx, u, &resourceclient.ListOptions{
		Branch:       r.branch,
		ExprSelector: &resourcepb.ExpressionSelector{},
	})

	//fmt.Println("event subscription", u.GetAPIVersion(), u.GetKind())
	for {
		select {
		case <-ctx.Done():
			return
		case rsp, ok := <-rspCh:
			if !ok {
				log.Debug("watch done...")
				return
			}

			u, err := object.GetUnstructured(rsp.Object)
			if err != nil {
				log.Error("unstructured conversion failed")
				continue
			}

			/*
				if err := r.client.Get(ctx, types.NamespacedName{
					Name:      u.GetName(),
					Namespace: u.GetNamespace(),
				}, u, rest.GetOptions{}); err != nil {
					log.Error("cannot get object", "err", err)
					continue
				}
			*/

			//fmt.Println("event received", u.GetAPIVersion(), u.GetKind(), u.GetName(), rsp.EventType.String())

			for _, handler := range r.eventHandlers.list() {
				go handler(ctx, rsp.EventType, u)
			}
		}
	}
}
