/*
Copyright 2023 The Nephio Authors.

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

package resources

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/grpcerrors"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
)

func New(name string, client resourceclient.Client, cr *unstructured.Unstructured, owns sets.Set[schema.GroupVersionKind], branch string) *Resources {
	return &Resources{
		name:              name,
		client:            client,
		branch:            branch,
		cr:                cr,
		owns:              owns,
		newResources:      []*unstructured.Unstructured{},
		existingResources: map[corev1.ObjectReference]*unstructured.Unstructured{},
	}
}

type Resources struct {
	name              string
	client            resourceclient.Client
	branch            string
	cr                *unstructured.Unstructured
	owns              sets.Set[schema.GroupVersionKind]
	m                 sync.RWMutex
	newResources      []*unstructured.Unstructured // we use a list to preserve the order -> important for claims
	existingResources map[corev1.ObjectReference]*unstructured.Unstructured
}

func GetObjectReference(u *unstructured.Unstructured) corev1.ObjectReference {
	return corev1.ObjectReference{
		APIVersion: u.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       u.GetObjectKind().GroupVersionKind().Kind,
		Namespace:  u.GetNamespace(),
		Name:       u.GetName(),
	}
}

// AddNewResource adds a new resource to the inventoru
func (r *Resources) AddNewResource(ctx context.Context, u *unstructured.Unstructured) {
	// TODO check if this is part of the owns list

	r.m.Lock()
	defer r.m.Unlock()

	log := log.FromContext(ctx)

	u.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: r.cr.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			Kind:       r.cr.GetObjectKind().GroupVersionKind().Kind,
			Name:       r.cr.GetName(),
			UID:        r.cr.GetUID(),
			Controller: ptr.To(true),
		},
	})

	ref := GetObjectReference(u)

	log.Debug("add newresource", "ref", ref.String())

	r.newResources = append(r.newResources, u)
}

// GetExistingResources retrieves the exisiting resource that match the label selector and the owner reference
// and puts the results in the resource inventory
func (r *Resources) getExistingResources(ctx context.Context) error {
	log := log.FromContext(ctx)

	var errm error
	for _, gvk := range r.owns.UnsortedList() {
		gvk := gvk

		// ideally we should do a list with ownerreferences- selector

		//ownObjList := ownObj.NewObjList()
		ul := &unstructured.UnstructuredList{}
		ul.SetGroupVersionKind(gvk)
		if err := r.client.List(ctx, ul, &resourceclient.ListOptions{
			ExprSelector: &resourcepb.ExpressionSelector{},
			Branch:       r.branch,
		}); err != nil {
			log.Error("getExistingResources list failed", "err", err.Error())
			errm = errors.Join(errm, err)
			continue
		}
		log.Debug("getExistingResources items", "gvk", gvk.String(), "items", len(ul.Items))
		for _, o := range ul.Items {
			log.Debug("getExistingResources item", "gvk", o.GetObjectKind().GroupVersionKind().String(), "name", o.GetName())
			o := o
			for _, ref := range o.GetOwnerReferences() {
				log.Debug("ownerref", "refs", fmt.Sprintf("%s/%s", ref.UID, r.cr.GetUID()))
				if ref.UID == r.cr.GetUID() {
					r.existingResources[corev1.ObjectReference{
						APIVersion: gvk.GroupVersion().String(),
						Kind:       gvk.Kind,
						Name:       o.GetName(),
						Namespace:  o.GetNamespace()}] = &o
				}
			}
		}
	}
	return errm
}

// APIDelete is used to delete the existing resources that are owned by this cr
// the implementation retrieves the existing resources and deletes them
func (r *Resources) Delete(ctx context.Context) error {
	r.m.Lock()
	defer r.m.Unlock()

	// step 0: get existing resources
	if err := r.getExistingResources(ctx); err != nil {
		return err
	}
	return r.apiDelete(ctx)
}

func (r *Resources) apiDelete(ctx context.Context) error {
	// delete in priority
	var errm error
	for ref, o := range r.existingResources {
		ref := ref
		o := o
		if ref.Kind == "Namespace" {
			continue
		}
		if err := r.delete(ctx, ref, o); err != nil {
			errm = errors.Join(errm, err)
			continue
		}
	}
	for ref, o := range r.existingResources {
		ref := ref
		o := o
		if err := r.delete(ctx, ref, o); err != nil {
			errm = errors.Join(errm, err)
			continue
		}
	}
	return nil
}

func (r *Resources) delete(ctx context.Context, ref corev1.ObjectReference, u *unstructured.Unstructured) error {
	log := log.FromContext(ctx)
	log.Info("api delete existing resource", "referernce", ref.String())
	if err := r.client.Delete(ctx, u, &resourceclient.DeleteOptions{
		Branch: r.branch,
	}); err != nil {
		if !grpcerrors.IsNotFound(err) {
			log.Error("api delete", "error", err, "object", u)
			return err
		}
		delete(r.existingResources, ref)
	}
	return nil
}

// APIApply
// step 0: get existing resources
// step 1: remove the exisiting resources from the internal resource list that overlap with the new resources
// step 2: delete the exisiting resources that are no longer needed
// step 3: apply the new resources to the api server
func (r *Resources) Apply(ctx context.Context) error {
	r.m.Lock()
	defer r.m.Unlock()

	log := log.FromContext(ctx)
	// step 0: get existing resources
	if err := r.getExistingResources(ctx); err != nil {
		return err
	}

	// step 1: remove the exisiting resources that overlap with the new resources
	// since apply will change the content.
	for _, u := range r.newResources {
		delete(r.existingResources, GetObjectReference(u))
	}

	log.Debug("api apply existing resources to be deleted", "existing resources", r.getExistingRefs())
	// step2b delete the exisiting resource that are no longer needed
	if err := r.apiDelete(ctx); err != nil {
		return err
	}

	// step3b apply the new resources to the api server
	return r.apiApply(ctx)
}

func (r *Resources) apiApply(ctx context.Context) error {
	// apply in priority
	var errm error
	for _, u := range r.newResources {
		u := u
		if u.GetKind() == "Namespace" { // apply in priority
			if err := r.apply(ctx, u); err != nil {
				errm = errors.Join(errm, err)
				continue
			}
		} else {
			continue
		}
	}
	for _, u := range r.newResources {
		u := u
		if err := r.apply(ctx, u); err != nil {
			errm = errors.Join(errm, err)
			continue
		}
	}
	return errm
}

func (r *Resources) apply(ctx context.Context, u *unstructured.Unstructured) error {
	log := log.FromContext(ctx)
	key := types.NamespacedName{Namespace: u.GetNamespace(), Name: u.GetName()}
	log.Debug("api apply object", "key", key.String(), "gvk", u.GetObjectKind().GroupVersionKind().String())

	return r.client.Apply(ctx, u, &resourceclient.ApplyOptions{
		FieldManager: r.name,
		Branch:       r.branch,
	})
}

func (r *Resources) GetNewResources() map[corev1.ObjectReference]*unstructured.Unstructured {
	r.m.RLock()
	defer r.m.RUnlock()

	res := make(map[corev1.ObjectReference]*unstructured.Unstructured, len(r.newResources))
	for _, u := range r.newResources {
		u := u
		res[GetObjectReference(u)] = u
	}
	return res
}

func (r *Resources) getExistingRefs() []corev1.ObjectReference {
	l := []corev1.ObjectReference{}
	for ref := range r.existingResources {
		ref := ref
		l = append(l, ref)
	}
	return l
}
