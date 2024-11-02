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

package pushcmd

import (
	"context"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/choreoclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func NewCmdPush(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewPushFlags()

	cmd := &cobra.Command{
		Use:   "push [flags]",
		Short: "push the resource to the version control",
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

type PushFlags struct {
}

// The defaults are determined here
func NewPushFlags() *PushFlags {
	return &PushFlags{}
}

// AddFlags add flags tp the command
func (r *PushFlags) AddFlags(cmd *cobra.Command) {
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *PushFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*PushOptions, error) {
	options := &PushOptions{
		Factory: f,
		Streams: streams,
	}
	return options, nil
}

type PushOptions struct {
	Factory util.Factory
	Streams *genericclioptions.IOStreams
}

func (r *PushOptions) Validate(args []string) error {
	return nil
}

func (r *PushOptions) Run(ctx context.Context, args []string) error {
	choreoClient := r.Factory.GetChoreoClient()
	if err := choreoClient.Push(ctx, &choreoclient.PushOptions{
		Proxy: r.Factory.GetProxy(),
	}); err != nil {
		return err
	}
	return nil
}
