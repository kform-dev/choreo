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

package apiresourcescmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/kform-dev/choreo/pkg/proto/grpcerrors"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

// NewCmdApply returns a cobra command.
func NewCmdAPIResources(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewAPIResourcesFlags()

	cmd := &cobra.Command{
		Use:   "api-resources",
		Short: "get api resources from the resource",
		//Args:  cobra.ExactArgs(0),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := flags.ToOptions(cmd, f, streams)
			if err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			return o.Run(cmd.Context())
		},
	}
	flags.AddFlags(cmd)
	return cmd
}

type APIResourcesFlags struct{}

func NewAPIResourcesFlags() *APIResourcesFlags {
	return &APIResourcesFlags{}
}

// AddFlags add flags to the command
func (r *APIResourcesFlags) AddFlags(cmd *cobra.Command) {}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *APIResourcesFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*APIResourcesOptions, error) {
	options := &APIResourcesOptions{
		Factory: f,
		Streams: streams,
	}
	return options, nil
}

type APIResourcesOptions struct {
	Factory util.Factory
	Streams *genericclioptions.IOStreams
	Output  string
}

// Complete adapts from the command line args and validates them
func (r *APIResourcesOptions) Complete() error {
	return nil
}

func (r *APIResourcesOptions) Validate() error {
	return nil
}

func (r *APIResourcesOptions) Run(ctx context.Context) error {
	branch := r.Factory.GetBranch()
	proxy := r.Factory.GetProxy()

	apiresources, err := r.Factory.GetDiscoveryClient().APIResources(ctx, proxy, branch)
	if err != nil {
		if grpcerrors.IsNotFound(err) {
			if branch == "" {
				return fmt.Errorf("cannot get apiresources (checkedout)")
			}
			if proxy.Name != "" && proxy.Namespace != "" {
				return fmt.Errorf("cannot get apiresources through proxy %s, branch %s not found", proxy.String(), branch)
			}
			return fmt.Errorf("cannot get apiresources, branch %s not found", branch)
		}
		return err
	}
	w := r.Streams.Out

	var errm error
	for _, apiResource := range apiresources {
		switch r.Output {
		case "name":
			// used for autocompletion
			name := fmt.Sprintf("%s.%s", apiResource.Resource, apiResource.Group)
			if _, err := fmt.Fprintf(w, "%s\n", name); err != nil {
				errm = errors.Join(errm, err)
			}
		default:
			if _, err := fmt.Fprintf(w, "%v\n", apiResource); err != nil {
				errm = errors.Join(errm, err)
			}
		}
	}

	return errm
}
