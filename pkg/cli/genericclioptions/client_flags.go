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
	flagProxy     = "proxy"
	flagBranch    = "branch"
	flagMaxRcvMsg = "maxRcvMsg"
	flagTimeout   = "timeout"
	flagCacheDir  = "cacheDir"
)

// ResourceFlags are flags for generic resources.
type ClientFlags struct {
	Branch    *string
	Proxy     *string
	MaxRcvMsg *int
	Timeout   *int
	CacheDir  *string
}

func NewClientFlags() *ClientFlags {
	return &ClientFlags{
		Branch:    ptr.To(""),
		Proxy:     ptr.To(""),
		MaxRcvMsg: ptr.To(25165824),
		Timeout:   ptr.To(10),
		CacheDir:  ptr.To(getDefaultCacheDir(getConfigPath())),
	}
}

// AddFlags binds file name flags to a given flagset
func (r *ClientFlags) AddFlags(flags *pflag.FlagSet) {
	if r == nil {
		return
	}

	if r.Branch != nil {
		flags.StringVarP(r.Branch, flagBranch, "b", *r.Branch,
			"branch from which the client wants to retrieve the info")
	}
	if r.Proxy != nil {
		flags.StringVarP(r.Proxy, flagProxy, "p", *r.Proxy,
			"proxy context from which the client wants to retrieve the info")
	}
	if r.MaxRcvMsg != nil {
		flags.IntVar(r.MaxRcvMsg, flagMaxRcvMsg, *r.MaxRcvMsg,
			"the maximum message size in bytes the client can receive")
	}
	if r.Timeout != nil {
		flags.IntVar(r.Timeout, flagTimeout, *r.Timeout,
			"gRPC dial timeout in seconds")
	}
	if r.CacheDir != nil {
		flags.StringVar(r.CacheDir, flagCacheDir, *r.CacheDir,
			"cache directory where the api resource information is stored")
	}
}
