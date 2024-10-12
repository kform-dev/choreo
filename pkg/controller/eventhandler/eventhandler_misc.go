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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
)

type Custom struct {
	Client      resourceclient.Client
	Name        string
	Queue       workqueue.TypedRateLimitingInterface[types.NamespacedName]
	Selector    selector.SelectorBuilder
	ForGVK      schema.GroupVersionKind
	ForSelector selector.Selector
	BranchName  string
}

// CustomEventHandler builds a new map[string]string as input for a list selector that filters
// resources based on this selector
func (r *Custom) EventHandler(ctx context.Context, _ resourcepb.Watch_EventType, obj runtime.Unstructured) bool {
	usrc := &unstructured.Unstructured{
		Object: obj.UnstructuredContent(),
	}
	ctx = InitContext(ctx, r.Name, usrc)
	log := log.FromContext(ctx)

	matchers := r.Selector.GetSelector(obj.UnstructuredContent())
	// this means the selector did not match
	if len(matchers) == 0 {
		log.Debug("custom eventhandler selector no match found")
		return true
	}
	log.Debug("custom eventhandler", "matchers", matchers)

	ul := &unstructured.UnstructuredList{}
	ul.SetGroupVersionKind(r.ForGVK)
	if err := r.Client.List(ctx, ul, &resourceclient.ListOptions{
		ExprSelector: &resourcepb.ExpressionSelector{Match: matchers},
		Branch:       r.BranchName,
	}); err != nil {
		log.Error("custom eventhandler list failed", "error", err)
		return true
	}

	for _, u := range ul.Items {
		req := types.NamespacedName{
			Name:      u.GetName(),
			Namespace: u.GetNamespace(),
		}
		log.Debug("watch reconcile event", "src", obj.GetObjectKind().GroupVersionKind().String(), "name", usrc.GetName())

		if r.ForSelector != nil {
			if r.ForSelector.Matches(usrc.Object) {
				r.Queue.Add(req)
			}
		}
	}

	return true
}
