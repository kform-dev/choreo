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
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/henderiw/store"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	uobject "github.com/kform-dev/choreo/pkg/util/object"
	"github.com/kform-dev/kform/pkg/pkgio"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type DataLoaderUpstream struct {
	Cfg        *genericclioptions.ChoreoConfig
	PathInRepo string
	//UpstreamClient          resourceclient.Client // used to read data from commit from upstream repo
	Client                  resourceclient.Client
	Branch                  string
	ChildGVKSet             sets.Set[schema.GroupVersionKind]
	UpstreamAnnotationValue string
	NewUpstreamData         sets.Set[corev1.ObjectReference]
}

func (r *DataLoaderUpstream) Load(ctx context.Context, commit *object.Commit) error {
	r.NewUpstreamData = sets.New(corev1.ObjectReference{})
	if err := r.load(ctx, commit); err != nil {
		return err
	}

	if err := r.clean(ctx); err != nil {
		return err
	}
	return nil
}

func (r *DataLoaderUpstream) load(ctx context.Context, commit *object.Commit) error {
	//log := log.FromContext(ctx)

	reader := GetCommitFileReader(filepath.Join(r.PathInRepo, *r.Cfg.ServerFlags.InputPath), commit)

	datastore, err := reader.Read(ctx)
	if err != nil {
		return err
	}

	var errm error
	datastore.List(func(k store.Key, rn *yaml.RNode) {
		u, err := uobject.GetUnstructuredContent([]byte(rn.MustString()), uobject.ContentTypeYAML)
		if err != nil {
			errm = errors.Join(errm, err)
			return
		}

		setAnnotations(u, map[string]string{
			choreov1alpha1.ChoreoLoaderOriginKey: r.UpstreamAnnotationValue,
		})
		r.NewUpstreamData.Insert(
			corev1.ObjectReference{
				APIVersion: u.GetAPIVersion(),
				Kind:       u.GetKind(),
				Namespace:  u.GetNamespace(),
				Name:       u.GetName(),
			},
		)
		if err := r.Client.Apply(ctx, u, &resourceclient.ApplyOptions{
			FieldManager: ManagedFieldManagerInput,
			Branch:       r.Branch,
		}); err != nil {
			errm = errors.Join(errm, err)
			return
		}
	})
	return errm
}

func (r *DataLoaderUpstream) clean(ctx context.Context) error {
	var errm error
	for _, gvk := range r.ChildGVKSet.UnsortedList() {
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
				u.GetAnnotations()[choreov1alpha1.ChoreoLoaderOriginKey] == r.UpstreamAnnotationValue &&
				uobject.IsManagedBy(u.GetManagedFields(), ManagedFieldManagerInput) {
				if !r.NewUpstreamData.Has(corev1.ObjectReference{
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

func GetCommitFileReader(path string, commit *object.Commit) pkgio.Reader[*yaml.RNode] {
	return &pkgio.CommitYAMLReader{
		Commit:  commit,
		Path:    path,
		SkipDir: true,
	}
}
