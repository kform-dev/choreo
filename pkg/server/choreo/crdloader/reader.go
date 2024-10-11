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

package crdloader

import (
	"embed"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kform-dev/kform/pkg/fsys"
	"github.com/kform-dev/kform/pkg/pkgio"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Embed specific files
//
//go:embed embedded/* internal/* apiextension/*
var EmbeddedAPIFS embed.FS

func GetAPIExtReader() pkgio.Reader[*yaml.RNode] {
	return &pkgio.YAMLDirReader{
		RelFsysPath: "apiextension",
		Fsys:        EmbeddedAPIFS,
		SkipDir:     false,
	}
}

func GetEmbeddedAPIReader() pkgio.Reader[*yaml.RNode] {
	return &pkgio.YAMLDirReader{
		RelFsysPath: "embedded",
		Fsys:        EmbeddedAPIFS,
		SkipDir:     false,
	}
}

func GetInternalAPIReader() pkgio.Reader[*yaml.RNode] {
	return &pkgio.YAMLDirReader{
		RelFsysPath: "internal",
		Fsys:        EmbeddedAPIFS,
		SkipDir:     false,
	}
}

func GetFileAPICRDReader(abspath string) pkgio.Reader[*yaml.RNode] {
	gvks := []schema.GroupVersionKind{
		apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"),
	}
	// for filesystem read we need to provide the full path to the root of the repo
	if !fsys.PathExists(abspath) {
		return nil
	}
	return GetFSYAMLReader(abspath, gvks)
}

func GetFSYAMLReader(path string, matchgvks []schema.GroupVersionKind) pkgio.Reader[*yaml.RNode] {
	return &pkgio.YAMLDirReader{
		FsysPath:  path,
		SkipDir:   true,
		MatchGVKs: matchgvks,
	}
}

func GetCommitFileAPICRDReader(crdPath string, commit *object.Commit) pkgio.Reader[*yaml.RNode] {
	gvks := []schema.GroupVersionKind{
		apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"),
	}
	return &pkgio.CommitYAMLReader{
		Commit:    commit,
		Path:      crdPath,
		SkipDir:   true,
		MatchGVKs: gvks,
	}
}
