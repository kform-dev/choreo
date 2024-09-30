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
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/util/object"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *storage) Apply(ctx context.Context, new runtime.Unstructured, opts ...rest.ApplyOption) (runtime.Unstructured, error) {
	o := rest.ApplyOptions{}
	o.ApplyOptions(opts)

	log := log.FromContext(ctx)
	log.Debug("apply")

	if o.FieldManager == "" {
		return nil, status.Errorf(codes.InvalidArgument, "cannot apply a resource with an empty fieldmanager")
	}

	if r.fieldManager == nil {
		return nil, status.Errorf(codes.Internal, "cannot apply a resource without a fieldmanager")
	}

	newObjectMeta, err := meta.Accessor(new)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot access objectMeta err: %s", err.Error())
	}

	old, err := r.Get(ctx, newObjectMeta.GetName(), &rest.GetOptions{
		ShowManagedFields: true,
		Trace:             "apply",
	})
	if err != nil {
		// Create an empty object we supply to the Apply fieldmanager as the liveObject
		// this ensures the fieldmanager does not add the before-first-apply manager as the operation

		empty := object.Empty(new)
		obj, err := r.fieldManager.Apply(empty, new, o.FieldManager, o.Force)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "fieldmanager apply failed err: %s", err.Error())
		}
		uobj, ok := obj.(runtime.Unstructured)
		if !ok {
			return nil, status.Errorf(codes.Internal, "fieldmanager does not return an unstructured object")
		}
		return r.Create(ctx, uobj, &rest.CreateOptions{DryRun: o.DryRun})
	}
	oldu := &unstructured.Unstructured{
		Object: old.UnstructuredContent(),
	}
	newu := &unstructured.Unstructured{
		Object: new.UnstructuredContent(),
	}
	// need to copy the creationtimestamp otherwise
	newu.SetCreationTimestamp(oldu.GetCreationTimestamp())
	//newu.SetResourceVersion(oldu.GetResourceVersion())
	//newu.SetGeneration(oldu.GetGeneration())

	newobj, err := r.fieldManager.Apply(old, new, o.FieldManager, o.Force)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "fieldmanager apply failed err: %s", err.Error())
	}
	new, ok := newobj.(runtime.Unstructured)
	if !ok {
		return nil, status.Errorf(codes.Internal, "fieldmanager does not return an unstructured object")
	}
	//fmt.Println("update apply", new)

	return r.update(ctx, new, old, &rest.UpdateOptions{DryRun: o.DryRun})
}
