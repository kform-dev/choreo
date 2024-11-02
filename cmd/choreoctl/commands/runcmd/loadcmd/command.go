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

package loadcmd

import (
	"context"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/runnerclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func NewCmdLoad(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewLoadFlags()

	cmd := &cobra.Command{
		Use:   "load [flags]",
		Short: "load data",
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

type LoadFlags struct {
}

// The defaults are determined here
func NewLoadFlags() *LoadFlags {
	return &LoadFlags{}
}

// AddFlags add flags tp the command
func (r *LoadFlags) AddFlags(cmd *cobra.Command) {
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *LoadFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*LoadOptions, error) {
	options := &LoadOptions{
		Factory: f,
		Streams: streams,
	}
	return options, nil
}

type LoadOptions struct {
	Factory util.Factory
	Streams *genericclioptions.IOStreams
}

func (r *LoadOptions) Validate(args []string) error {
	return nil
}

func (r *LoadOptions) Run(ctx context.Context, args []string) error {
	runnerClient := r.Factory.GetRunnerClient()
	if err := runnerClient.Load(ctx, &runnerclient.LoadOptions{
		Proxy: r.Factory.GetProxy(),
	}); err != nil {
		return err
	}
	return nil
}
