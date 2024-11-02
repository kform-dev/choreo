/*
Copyright 2018 The Kubernetes Authors.

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
	"k8s.io/utils/ptr"
)

const (
	FlagServerName          = "servername"
	flagAPIs                = "apis"
	flagDB                  = "db"
	flagReconcilers         = "reconcilers"
	flagLibraries           = "libraries"
	flagInput               = "input"
	flagPostprocessing      = "post"
	flagOutput              = "output"
	flagRefs                = "refs"
	flagInternalReconcilers = "internalReconcilers"
)

// ResourceFlags are flags for generic resources.
type ServerFlags struct {
	ServerName          *string
	CRDPath             *string
	DBPath              *string
	ReconcilerPath      *string
	LibraryPath         *string
	PostProcessingPath  *string
	OutputPath          *string
	InputPath           *string
	RefsPath            *string
	InternalReconcilers *bool
}

func NewServerFlags() *ServerFlags {
	return &ServerFlags{
		ServerName:          ptr.To("choreo"),
		CRDPath:             ptr.To("crds"),
		DBPath:              ptr.To("db"),
		ReconcilerPath:      ptr.To("reconcilers"),
		LibraryPath:         ptr.To("libs"),
		PostProcessingPath:  ptr.To("post"),
		OutputPath:          ptr.To("out"),
		InputPath:           ptr.To("in"),
		RefsPath:            ptr.To("refs"),
		InternalReconcilers: ptr.To(false),
	}
}

// AddFlags binds file name flags to a given flagset
func (r *ServerFlags) AddFlags(flags *pflag.FlagSet) {
	if r == nil {
		return
	}
	if r.ServerName != nil {
		flags.StringVar(r.ServerName, FlagServerName, *r.ServerName,
			"the name of the server")
	}
	if r.CRDPath != nil {
		flags.StringVar(r.CRDPath, flagAPIs, *r.CRDPath,
			"the path where the api manifests are located")
	}
	if r.DBPath != nil {
		flags.StringVar(r.DBPath, flagDB, *r.DBPath,
			"the path where the db manifests are located")
	}
	if r.ReconcilerPath != nil {
		flags.StringVar(r.ReconcilerPath, flagReconcilers, *r.ReconcilerPath,
			"the path where the reconciler manifests are located")
	}
	if r.LibraryPath != nil {
		flags.StringVar(r.LibraryPath, flagLibraries, *r.LibraryPath,
			"the path where the libraries are located")
	}
	if r.InputPath != nil {
		flags.StringVar(r.InputPath, flagInput, *r.InputPath,
			"the path where the input is located")
	}
	if r.RefsPath != nil {
		flags.StringVar(r.RefsPath, flagRefs, *r.RefsPath,
			"the path where the refs is located")
	}
	if r.PostProcessingPath != nil {
		flags.StringVar(r.PostProcessingPath, flagPostprocessing, *r.PostProcessingPath,
			"the path where the postprocessing files are located")
	}
	if r.InternalReconcilers != nil {
		flags.BoolVarP(r.InternalReconcilers, flagInternalReconcilers, "r", *r.InternalReconcilers,
			"enable internal reconciler")
	}
}
