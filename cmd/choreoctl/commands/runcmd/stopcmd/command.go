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

package stopcmd

import (
	"context"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/runnerclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func NewCmdStop(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewStopFlags()

	cmd := &cobra.Command{
		Use:   "stop [flags]",
		Short: "stop the reconcilers",
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

type StopFlags struct {
}

// The defaults are determined here
func NewStopFlags() *StopFlags {
	return &StopFlags{}
}

// AddFlags add flags tp the command
func (r *StopFlags) AddFlags(cmd *cobra.Command) {
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *StopFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*StopOptions, error) {
	options := &StopOptions{
		Factory: f,
		Streams: streams,
	}
	return options, nil
}

type StopOptions struct {
	Factory util.Factory
	Streams *genericclioptions.IOStreams
}

func (r *StopOptions) Validate(args []string) error {
	return nil
}

func (r *StopOptions) Run(ctx context.Context, args []string) error {
	runnerClient := r.Factory.GetRunnerClient()
	if err := runnerClient.Stop(ctx, &runnerclient.StopOptions{
		Proxy: r.Factory.GetProxy(),
	}); err != nil {
		return err
	}
	return nil
}
