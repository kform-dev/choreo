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

package createcmd

import (
	"context"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/branchclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func NewCmdCreate(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewCreateFlags()

	cmd := &cobra.Command{
		Use:  "create BRANCHNAME [flags]",
		Args: cobra.ExactArgs(1),
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

type CreateFlags struct {
}

// The defaults are determined here
func NewCreateFlags() *CreateFlags {
	return &CreateFlags{}
}

// AddFlags add flags tp the command
func (r *CreateFlags) AddFlags(cmd *cobra.Command) {
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *CreateFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*CreateOptions, error) {
	options := &CreateOptions{
		Factory: f,
		Streams: streams,
	}
	return options, nil
}

type CreateOptions struct {
	Factory util.Factory
	Streams *genericclioptions.IOStreams
}

func (r *CreateOptions) Validate(args []string) error {
	return nil
}

func (r *CreateOptions) Run(ctx context.Context, args []string) error {
	branchClient := r.Factory.GetBranchClient()
	branchName := args[0]
	if err := branchClient.Create(ctx, branchName, &branchclient.CreateOptions{
		Proxy: r.Factory.GetProxy(),
	}); err != nil {
		return err
	}
	return nil
}
