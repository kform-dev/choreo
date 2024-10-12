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

package depscmd

import (
	"context"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/kform-dev/choreo/pkg/util/inventory"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func GetCommand(ctx context.Context, f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	return NewRunner(ctx, f, streams).Command
}

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, f util.Factory, streams *genericclioptions.IOStreams) *Runner {
	r := &Runner{
		factory: f,
		streams: streams,
	}
	cmd := &cobra.Command{
		Use:   "deps",
		Short: "get dependencies between resources",
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: r.runE,
	}

	// Adding a boolean flag named "show-internal" with a default value of false and description
	cmd.Flags().BoolVarP(&r.showChoreo, "show-choreo", "i", false, "Enable displaying internal choreo resources")

	r.Command = cmd
	return r
}

type Runner struct {
	Command    *cobra.Command
	factory    util.Factory
	streams    *genericclioptions.IOStreams
	showChoreo bool
}

func (r *Runner) runE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	branch := r.factory.GetBranch()
	proxy := r.factory.GetProxy()

	apiResources, err := r.factory.GetDiscoveryClient().APIResources(ctx, proxy, branch)
	if err != nil {
		return err
	}

	inv := inventory.Inventory{}
	if err := inv.Build(ctx, r.factory.GetResourceClient(), apiResources, branch, r.showChoreo); err != nil {
		return err
	}
	inv.Print()
	return nil
}

/*

type Options struct {
	Factory util.Factory
	Streams *genericclioptions.IOStreams
	Output  string
	// derived parameters
	//branch string
}

// Complete adapts from the command line args and validates them
func (r *Options) Complete(ctx context.Context) error {
	return nil
}

func (r *Options) Validate(ctx context.Context) error {
	return nil
}

func (r *Options) Run(ctx context.Context) error {
	branch := r.Factory.GetBranch()
	proxy := r.Factory.GetProxy()
	apiresources, err := r.Factory.GetDiscoveryClient().APIResources(ctx, proxy, branch)
	if err != nil {
		return err
	}
	w := r.Streams.Out

	var errm error
	for _, apiResource := range apiresources {
		switch r.Output {
		case "name":
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
*/
