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

package loader

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/henderiw/store"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/util/object"
	"github.com/kform-dev/kform/pkg/fsys"
	"github.com/kform-dev/kform/pkg/pkgio"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	syaml "sigs.k8s.io/yaml"
)

func (r *DataLoader) loadInput(ctx context.Context) error {
	loader := &InputLoader{
		Client:         r.Client,
		Branch:         r.Branch,
		NewInput:       sets.New[corev1.ObjectReference](),
		APIStore:       r.APIStore,
		InternalAPISet: r.InternalAPISet,
	}
	if err := loader.Load(ctx, r.getInputReader(r.RepoPth, r.PathInRepo, r.Flags)); err != nil {
		return err
	}
	if err := loader.Clean(ctx); err != nil {
		return err
	}
	return nil

}

func (r *DataLoader) getInputReader(repoPath, pathInRepo string, flags *genericclioptions.ConfigFlags) pkgio.Reader[*yaml.RNode] {
	abspath := filepath.Join(repoPath, pathInRepo, *flags.InputPath)
	gvks := []schema.GroupVersionKind{}

	if !fsys.PathExists(abspath) {
		return nil
	}
	return GetFSYAMLReader(abspath, gvks)
}

type InputLoader struct {
	Client         resourceclient.Client
	Branch         string
	NewInput       sets.Set[corev1.ObjectReference]
	APIStore       *api.APIStore
	InternalAPISet sets.Set[schema.GroupVersionKind]
}

func (r *InputLoader) Load(ctx context.Context, reader pkgio.Reader[*yaml.RNode]) error {
	fmt.Println("input loader", reader)
	if reader == nil {
		// reader nil, mean the path does not exist, whcich is ok
		return nil
	}
	datastore, err := reader.Read(ctx)
	if err != nil {
		return err
	}

	var errm error
	datastore.List(func(k store.Key, rn *yaml.RNode) {
		fmt.Println("input loader", rn.GetKind(), rn.GetName())

		a := rn.GetAnnotations()
		if len(a) == 0 {
			a = map[string]string{}
		}
		a[choreov1alpha1.ChoreoLoaderOriginKey] = choreov1alpha1.FileLoaderAnnotation.String()
		rn.SetAnnotations(a)

		object := map[string]any{}
		if err := syaml.Unmarshal([]byte(rn.MustString()), &object); err != nil {
			errm = errors.Join(errm, err)
			return
		}
		obj := &unstructured.Unstructured{Object: object}

		r.NewInput.Insert(corev1.ObjectReference{
			APIVersion: obj.GetAPIVersion(),
			Kind:       obj.GetKind(),
			Namespace:  obj.GetNamespace(),
			Name:       obj.GetName(),
		})
		if err := r.Client.Apply(ctx, obj, &resourceclient.ApplyOptions{
			FieldManager: ManagedFieldManagerInput,
			Origin:       ManagedFieldManagerInput,
			Branch:       r.Branch,
		}); err != nil {
			errm = errors.Join(errm, err)
			return
		}
	})
	return errm
}

func (r *InputLoader) Clean(ctx context.Context) error {
	var errm error
	for _, gvk := range r.APIStore.GetGVKSet().UnsortedList() {
		// dont look at internal apis
		if r.InternalAPISet.Has(gvk) {
			continue
		}
		ul := &unstructured.UnstructuredList{}
		ul.SetGroupVersionKind(gvk)
		if err := r.Client.List(ctx, ul, &resourceclient.ListOptions{
			ExprSelector:      &resourcepb.ExpressionSelector{},
			ShowManagedFields: true,
			Branch:            r.Branch,
		}); err != nil {
			errm = errors.Join(errm, err)
			continue
		}

		for _, u := range ul.Items {
			if len(u.GetAnnotations()) != 0 &&
				u.GetAnnotations()[choreov1alpha1.ChoreoLoaderOriginKey] == choreov1alpha1.FileLoaderAnnotation.String() &&
				object.IsManagedBy(u.GetManagedFields(), ManagedFieldManagerInput) {
				if !r.NewInput.Has(corev1.ObjectReference{
					APIVersion: u.GetAPIVersion(),
					Kind:       u.GetKind(),
					Namespace:  u.GetNamespace(),
					Name:       u.GetName(),
				}) {
					if err := r.Client.Delete(ctx, &u, &resourceclient.DeleteOptions{
						Branch: r.Branch,
					}); err != nil {
						errm = errors.Join(errm, err)
					}
				}
			}
		}
	}
	return nil
}
