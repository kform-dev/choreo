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

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/kform/pkg/fsys"
	"github.com/kform-dev/kform/pkg/pkgio"
	"github.com/kform-dev/kform/pkg/pkgio/ignore"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type DataLoader struct {
	Cfg        *genericclioptions.ChoreoConfig
	Client     resourceclient.Client
	Branch     string
	GVKs       []schema.GroupVersionKind
	RepoPth    string
	PathInRepo string
	// used to clean reasources
	//APIStore       *api.APIStore
	InternalAPISet sets.Set[schema.GroupVersionKind]
}

func (r *DataLoader) Load(ctx context.Context) error {
	var errm error

	if err := fsys.EnsureDir(ctx, r.RepoPth, r.PathInRepo, *r.Cfg.ServerFlags.InputPath); err != nil {
		return err
	}

	if err := r.loadInput(ctx); err != nil {
		errm = errors.Join(errm, fmt.Errorf("cannot load input, err: %v", err))
	}

	return errm
}

func GetFSReader(path string) pkgio.Reader[[]byte] {
	fsys := fsys.NewDiskFS(path)
	ignoreRules := ignore.Empty(pkgio.IgnoreFileMatch[0])
	f, err := fsys.Open(pkgio.IgnoreFileMatch[0])
	if err == nil {
		// if an error is return the rules is empty, so we dont have to worry about the error
		ignoreRules, _ = ignore.Parse(f)
	}
	return &pkgio.DirReader{
		RelFsysPath:    ".",
		Fsys:           fsys,
		MatchFilesGlob: pkgio.MatchAll,
		IgnoreRules:    ignoreRules,
		SkipDir:        true,
	}
}

func GetFSStarReader(path string) pkgio.Reader[[]byte] {
	fsys := fsys.NewDiskFS(path)
	ignoreRules := ignore.Empty(pkgio.IgnoreFileMatch[0])
	f, err := fsys.Open(pkgio.IgnoreFileMatch[0])
	if err == nil {
		// if an error is return the rules is empty, so we dont have to worry about the error
		ignoreRules, _ = ignore.Parse(f)
	}
	return &pkgio.DirReader{
		RelFsysPath:    ".",
		Fsys:           fsys,
		MatchFilesGlob: pkgio.MatchFilesGlob([]string{"*.star"}),
		IgnoreRules:    ignoreRules,
		SkipDir:        true,
	}
}

func GetFSReconcilerReader(path string) pkgio.Reader[[]byte] {
	fsys := fsys.NewDiskFS(path)
	ignoreRules := ignore.Empty(pkgio.IgnoreFileMatch[0])
	f, err := fsys.Open(pkgio.IgnoreFileMatch[0])
	if err == nil {
		// if an error is return the rules is empty, so we dont have to worry about the error
		ignoreRules, _ = ignore.Parse(f)
	}
	return &pkgio.DirReader{
		RelFsysPath:    ".",
		Fsys:           fsys,
		MatchFilesGlob: pkgio.MatchFilesGlob([]string{"config.yaml"}),
		IgnoreRules:    ignoreRules,
		SkipDir:        false,
	}
}

func GetFSYAMLReader(path string, matchgvks []schema.GroupVersionKind) pkgio.Reader[*yaml.RNode] {
	return &pkgio.YAMLDirReader{
		FsysPath:  path,
		SkipDir:   true,
		MatchGVKs: matchgvks,
	}
}
