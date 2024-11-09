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
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/kform/pkg/fsys"
	"github.com/kform-dev/kform/pkg/pkgio"
	"github.com/kform-dev/kform/pkg/pkgio/ignore"
)

// UpstreamLoader
type RunningConfigLoader struct {
	Cfg *genericclioptions.ChoreoConfig
	//Client     resourceclient.Client
	//Branch     string
	RepoPath       string
	PathInRepo     string
	RunningConfigs map[string]any
}

func GetRunningConfigReader(path string) pkgio.Reader[[]byte] {
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
		MatchFilesGlob: pkgio.MatchFilesGlob([]string{"*.json"}),
		IgnoreRules:    ignoreRules,
		SkipDir:        true,
	}
}

func (r *RunningConfigLoader) Load(ctx context.Context) error {
	abspath := filepath.Join(r.RepoPath, r.PathInRepo, *r.Cfg.ServerFlags.RunningConfigsPath)

	if !fsys.PathExists(abspath) {
		return nil
	}
	reader := GetRunningConfigReader(abspath)
	datastore, err := reader.Read(ctx)
	if err != nil {
		return err
	}

	var errs error
	if r.RunningConfigs == nil {
		r.RunningConfigs = map[string]any{}
	}
	datastore.List(func(k store.Key, jsonBytes []byte) {
		var j any
		err = json.Unmarshal(jsonBytes, &j)
		if err != nil {
			errs = errors.Join(errs, err)
		}
		// k.Name == filename
		nodeName := k.Name[:len(k.Name)-len(filepath.Ext(k.Name))]
		r.RunningConfigs[nodeName] = j
	})

	return errs
}
