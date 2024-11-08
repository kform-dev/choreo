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
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	//"k8s.io/apimachinery/pkg/util/sets"
)

type DevLoader struct {
	Cfg         *genericclioptions.ChoreoConfig
	RepoPath    string
	PathInRepo  string
	Libraries   []*choreov1alpha1.Library
	Reconcilers []*choreov1alpha1.Reconciler
	//DstPath string
	//NewLibraries sets.Set[string] -> TBD if we need a cleaner
	// Experiment to load libraries direct -> since now they are the same as apis
	//Client resourceclient.Client
	//Branch string
}

/*
func (r *DevLoader) Load(ctx context.Context) error {

	//	if err := fsys.EnsureDir(ctx, []string{r.DstPath, *r.Cfg.ServerFlags.InputPath}...); err != nil {
	//		return err
	//	}
	var errm error
	if err := r.loadReconcilers(ctx); err != nil {
		errm = errors.Join(errm, fmt.Errorf("cannot load reconcilers, err: %v", err))
	}
	if err := r.LoadLibraries(ctx); err != nil {
		errm = errors.Join(errm, fmt.Errorf("cannot load libraries, err: %v", err))
	}

	return errm
}
*/
