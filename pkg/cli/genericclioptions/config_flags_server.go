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

package genericclioptions

import (
	"github.com/spf13/pflag"
)

func (r *ConfigFlags) AddServerControllerFlags(fs *pflag.FlagSet) {
	if r.CRDPath != nil {
		fs.StringVar(r.CRDPath, flagAPIs, *r.CRDPath,
			"the path where the api manifests are located")
	}
	if r.DBPath != nil {
		fs.StringVar(r.DBPath, flagDB, *r.DBPath,
			"the path where the db manifests are located")
	}
	if r.ReconcilerPath != nil {
		fs.StringVar(r.ReconcilerPath, flagReconcilers, *r.ReconcilerPath,
			"the path where the reconciler manifests are located")
	}
	if r.LibraryPath != nil {
		fs.StringVar(r.LibraryPath, flagLibraries, *r.LibraryPath,
			"the path where the libraries are located")
	}
	if r.InputPath != nil {
		fs.StringVar(r.InputPath, flagInput, *r.InputPath,
			"the path where the input is located")
	}
	if r.RefsPath != nil {
		fs.StringVar(r.RefsPath, flagRefs, *r.RefsPath,
			"the path where the refs is located")
	}
	if r.PostProcessingPath != nil {
		fs.StringVar(r.PostProcessingPath, flagPostprocessing, *r.PostProcessingPath,
			"the path where the postprocessing files are located")
	}
	if r.InternalReconcilers != nil {
		fs.BoolVarP(r.InternalReconcilers, flagInternalReconcilers, "r", *r.InternalReconcilers,
			"enable internal reconciler")
	}
}
