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

package eventhandler

import (
	"context"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/selector"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
)

type For struct {
	Name     string
	Queue    workqueue.TypedRateLimitingInterface[types.NamespacedName]
	Selector selector.Selector
	Client   resourceclient.Client
}

func (r *For) EventHandler(ctx context.Context, _ resourcepb.Watch_EventType, obj runtime.Unstructured) bool {
	obj = obj.DeepCopyObject().(runtime.Unstructured)
	usrc := &unstructured.Unstructured{
		Object: obj.UnstructuredContent(),
	}
	ctx = InitContext(ctx, r.Name, usrc)
	log := log.FromContext(ctx)

	if r.Selector != nil {

		if !r.Selector.Matches(usrc.Object) {
			log.Debug("for reconcile event filtered", "src", obj.GetObjectKind().GroupVersionKind().String(), "name", usrc.GetName())
			return true
		}
	}
	req := types.NamespacedName{
		Name:      usrc.GetName(),
		Namespace: usrc.GetNamespace(),
	}
	log.Debug("watch reconcile event", "src", obj.GetObjectKind().GroupVersionKind().String(), "name", usrc.GetName())
	r.Queue.Add(req)
	return true
}
