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

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/henderiw/store"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/repository"
	"github.com/kform-dev/choreo/pkg/repository/repogit"
	"github.com/kform-dev/choreo/pkg/server/choreo/instance"
	uobject "github.com/kform-dev/choreo/pkg/util/object"
	"github.com/kform-dev/kform/pkg/fsys"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	syaml "sigs.k8s.io/yaml"
)

// UpstreamLoader
type UpstreamLoader struct {
	Parent     instance.ChoreoInstance
	Cfg        *genericclioptions.ChoreoConfig
	Client     resourceclient.Client
	Branch     string
	RepoPath   string
	PathInRepo string
	TempDir    string
	CallbackFn UpstreamCallBackFn
}

type UpstreamCallBackFn func(ctx context.Context, parentName string, repo repository.Repository, upstreamRef *choreov1alpha1.UpstreamRef, cfg *genericclioptions.ChoreoConfig, commit *object.Commit, annotationVal string) error

func (r *UpstreamLoader) Load(ctx context.Context) error {
	gvks := []schema.GroupVersionKind{
		choreov1alpha1.SchemeGroupVersion.WithKind(choreov1alpha1.UpstreamRefKind),
	}
	abspath := filepath.Join(r.RepoPath, r.PathInRepo, *r.Cfg.ServerFlags.RefsPath)

	if !fsys.PathExists(abspath) {
		return nil
	}
	reader := GetFSYAMLReader(abspath, gvks)
	datastore, err := reader.Read(ctx)
	if err != nil {
		return err
	}

	var errs error
	datastore.List(func(k store.Key, rn *yaml.RNode) {
		upstreamRef := &choreov1alpha1.UpstreamRef{}
		if err := syaml.Unmarshal([]byte(rn.MustString()), upstreamRef); err != nil {
			errs = errors.Join(errs, fmt.Errorf("invalid upstreamref %s, err: %v", k.Name, err))
			return
		}
		// upload the upstream to the apiserver
		//r.NewChoreoRef.Insert(k.Name)
		obj, err := uobject.GetUnstructructered(upstreamRef)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("cannot unmarshal %s, err: %v", k.Name, err))
			return
		}

		// update the apiserver with the refs
		if err := r.Client.Apply(ctx, obj, &resourceclient.ApplyOptions{
			Branch:       r.Branch,
			FieldManager: ManagedFieldManagerInput,
		}); err != nil {
			errs = errors.Join(errs, fmt.Errorf("cannot apply upstream ref %s, err: %v", k.Name, err))
			return
		}

		childRepoPath := filepath.Join(r.TempDir, upstreamRef.GetURLPath())
		refName := upstreamRef.GetPlumbingReference()
		url := upstreamRef.Spec.URL

		repo, commit, err := repogit.NewUpstreamRepo(ctx, childRepoPath, url, refName)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("cannot open repo %s, err: %v", url, err))
			return
		}

		childInstance, err := instance.NewChildChoreoInstance(ctx, repo, upstreamRef, r.Cfg, commit, upstreamRef.LoaderAnnotation().String())
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("cannot create child choreo instance for %s from repo %s, err: %v", refName, url, err))
			return
		}
		if err := r.Parent.AddChildChoreoInstance(childInstance); err != nil {
			errs = errors.Join(errs, err)
			return
		}
	})
	return errs
}
