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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
)

type Own struct {
	Name   string
	Client resourceclient.Client
	Queue  workqueue.TypedRateLimitingInterface[types.NamespacedName]
	GVK    schema.GroupVersionKind
}

func (r *Own) EventHandler(ctx context.Context, eventType resourcepb.Watch_EventType, obj runtime.Unstructured) bool {
	usrc := &unstructured.Unstructured{
		Object: obj.UnstructuredContent(),
	}
	ctx = InitContext(ctx, r.Name, usrc)
	log := log.FromContext(ctx)

	for _, ownerRef := range usrc.GetOwnerReferences() {
		if ownerRef.APIVersion == r.GVK.GroupVersion().Identifier() &&
			ownerRef.Kind == r.GVK.Kind {
			req := types.NamespacedName{
				Name:      ownerRef.Name,
				Namespace: usrc.GetNamespace(),
			}
			// we dont need a for selector since these resources were created by a for resource

			log.Debug("own reconcile event", "src", obj.GetObjectKind().GroupVersionKind().String(), "name", usrc.GetName())
			r.Queue.Add(req)
		}
	}
	return true
}
