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
	"strconv"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/server/apiserver/watch"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func (r *storage) notifyWatcher(ctx context.Context, event watch.Event) {
	log := log.FromContext(ctx).With("eventType", event.Type)
	log.Debug("notify watcherManager")

	r.watcherManager.WatchChan() <- event
}

func isSpecEqual(old, new runtime.Unstructured) (bool, error) {
	oldSpec, oldfound, err := unstructured.NestedMap(old.UnstructuredContent(), "spec")
	if err != nil {
		return false, status.Errorf(codes.Internal, "obj is not map[string]interface{} err: %s", err.Error())
	}
	newSpec, newfound, err := unstructured.NestedMap(new.UnstructuredContent(), "spec")
	if err != nil {
		return false, status.Errorf(codes.Internal, "obj is not map[string]interface{} err: %s", err.Error())
	}

	// one has a spec and the other doesnt -> this is seen as a change
	if oldfound != newfound {
		return false, nil
	}
	if oldfound && newfound {
		return apiequality.Semantic.DeepEqual(newSpec, oldSpec), nil
	}
	// no spec present in both object so spec is equal
	return true, nil
}

func UpdateResourceVersion(new, old metav1.Object) error {
	resourceVersion, err := strconv.Atoi(old.GetResourceVersion())
	if err != nil {
		return err
	}
	resourceVersion++
	new.SetResourceVersion(strconv.Itoa(resourceVersion))
	return nil
}

func UpdateGeneration(new metav1.Object, generation int64) {
	generation++
	new.SetGeneration(generation)
}
