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
		/*
			reqname := usrc.GetName()
			match := false
			if r.Name == "infra.kuid.dev_node_if-si-ni" || r.Name == "infra.kuid.dev_node_bgp" {
				match = true
			}
			conditionStatus := "unknown"
			conditionMessage := "unknown"

			if match {
				c := object.GetCondition(usrc.Object, "IPClaimReady")
				conditionStatus = fmt.Sprintf("%v", c["status"])
				conditionMessage = fmt.Sprintf("%v", c["message"])
				//fmt.Println(r.Name, reqname, "eventhandler for", "ipclaimready condition", conditionStatus, conditionMessage)
			}
		*/

		if !r.Selector.Matches(usrc.Object) {
			/*
				if match {
					fmt.Println(r.Name, reqname, "eventhandler for", "selectornomatch", conditionStatus, conditionMessage)
				}
			*/

			log.Debug("for reconcile event filtered", "src", obj.GetObjectKind().GroupVersionKind().String(), "name", usrc.GetName())
			return true
		}

		/*
			if match {
				if conditionStatus == "unknown" {
					fmt.Println("should never happens\n", usrc)
				}
				ul := &unstructured.UnstructuredList{}
				ul.SetAPIVersion("ipam.be.kuid.dev/v1alpha1")
				ul.SetKind("IPClaim")
				r.Client.List(ctx, ul, resourceclient.ListOptions{ExprSelector: &resourcepb.ExpressionSelector{}, Origin: "eventhandler"})

				for _, u := range ul.Items {
					fmt.Println(r.Name, reqname, "ipclam", u.GetName())
				}

				fmt.Println(r.Name, reqname, "eventhandler for", "selectormatch", conditionStatus, conditionMessage)
			}
		*/

	}
	req := types.NamespacedName{
		Name:      usrc.GetName(),
		Namespace: usrc.GetNamespace(),
	}
	log.Debug("watch reconcile event", "src", obj.GetObjectKind().GroupVersionKind().String(), "name", usrc.GetName())
	r.Queue.Add(req)
	return true
}
