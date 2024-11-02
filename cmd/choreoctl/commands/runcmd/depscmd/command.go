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

func NewCmdDeps(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewDepsFlags()

	cmd := &cobra.Command{
		Use:   "deps",
		Short: "get dependencies between resources",
		//Args:  cobra.ExactArgs(1),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			o, err := flags.ToOptions(cmd, f, streams)
			if err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			return o.Run(ctx, args)
		},
	}
	flags.AddFlags(cmd)
	return cmd
}

type DepsFlags struct {
	RunOuput *genericclioptions.RunOutputFlags
}

// The defaults are determined here
func NewDepsFlags() *DepsFlags {
	return &DepsFlags{
		RunOuput: genericclioptions.NewRunOutputFlags(),
	}
}

// AddFlags add flags tp the command
func (r *DepsFlags) AddFlags(cmd *cobra.Command) {
	r.RunOuput.AddFlags(cmd.Flags())
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *DepsFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*DepsOptions, error) {
	options := &DepsOptions{
		Factory:        f,
		Streams:        streams,
		ShowChoreoAPIs: *r.RunOuput.ShowChoreoAPIs,
	}
	return options, nil
}

type DepsOptions struct {
	Factory        util.Factory
	Streams        *genericclioptions.IOStreams
	ShowChoreoAPIs bool
}

func (r *DepsOptions) Validate(args []string) error {
	return nil
}

func (r *DepsOptions) Run(ctx context.Context, args []string) error {
	branch := r.Factory.GetBranch()
	proxy := r.Factory.GetProxy()

	apiResources, err := r.Factory.GetDiscoveryClient().APIResources(ctx, proxy, branch)
	if err != nil {
		return err
	}

	inv := inventory.Inventory{}
	if err := inv.Build(ctx, r.Factory.GetResourceClient(), apiResources, &inventory.BuildOptions{
		ShowManagedField: true,
		Branch:           branch,
		ShowChoreoAPIs:   r.ShowChoreoAPIs,
	}); err != nil {
		return err
	}
	inv.Print()
	return nil
}
