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

package registry

import (
	"context"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/server/apiserver/watch"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func (r *storage) Create(ctx context.Context, obj runtime.Unstructured, opts ...rest.CreateOption) (runtime.Unstructured, error) {
	o := rest.CreateOptions{}
	o.ApplyOptions(opts)

	log := log.FromContext(ctx)
	log.Debug("create choreoapiserver")

	objectMeta, err := meta.Accessor(obj)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot access objectMeta err: %s", err.Error())
	}
	objectMeta.SetCreationTimestamp(metav1.Now())
	objectMeta.SetUID(uuid.NewUUID())
	objectMeta.SetResourceVersion("0")

	if errs := r.createStrategy.ValidateCreate(ctx, obj); len(errs) > 0 {
		log.Error("validation create failed", "obj", obj, "error", errs)
		return nil, status.Errorf(codes.InvalidArgument, "validation create failed: %v", errs)
	}

	recursion := false
	if len(o.DryRun) == 1 && o.DryRun[0] == "recursion" {
		recursion = true
		o.DryRun = []string{}
	}

	robj, err := r.createStrategy.InvokeCreate(ctx, obj, recursion)
	if err != nil {
		log.Error("invoke synchronous create failed", "obj", obj, "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "invoke synchronous create failed: %v", err)
	}
	obj = robj.(runtime.Unstructured)

	// namespace is ignore, name is assumed to be always present
	key := objectMeta.GetName()

	if len(o.DryRun) > 0 {
		return obj, nil
	}

	if err := r.storage.Create(store.ToKey(key), obj); err != nil {
		return obj, status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	r.notifyWatcher(ctx, watch.Event{
		Type:   resourcepb.Watch_ADDED,
		Object: obj,
	})
	return obj, nil
}
