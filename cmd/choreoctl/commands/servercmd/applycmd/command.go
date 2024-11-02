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

package applycmd

import (
	"context"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/choreoclient"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func NewCmdApply(cfg *genericclioptions.ChoreoConfig) *cobra.Command {
	flags := NewApplyFlags()

	cmd := &cobra.Command{
		Use:  "apply URL DIR REF BRANCH [flags]",
		Args: cobra.MinimumNArgs(3),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := flags.ToOptions(cmd, cfg)
			if err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			return o.Run(cmd.Context(), args)
		},
	}
	flags.AddFlags(cmd)
	return cmd
}

type ApplyFlags struct {
}

// NewApplyFlags determines which flags will be added to the command
// The defaults are determined here
func NewApplyFlags() *ApplyFlags {
	return &ApplyFlags{}
}

// AddFlags add flags to the command
func (r *ApplyFlags) AddFlags(cmd *cobra.Command) {
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *ApplyFlags) ToOptions(cmd *cobra.Command, cfg *genericclioptions.ChoreoConfig) (*ApplyOptions, error) {
	options := &ApplyOptions{
		cfg: cfg,
	}
	return options, nil
}

type ApplyOptions struct {
	cfg *genericclioptions.ChoreoConfig
}

func (r *ApplyOptions) Validate(args []string) error {
	return nil
}

func (r *ApplyOptions) Run(ctx context.Context, args []string) error {
	client, err := r.cfg.ToChoreoClient()
	if err != nil {
		return err
	}
	defer client.Close()

	choreoCtx := &choreopb.ChoreoContext{
		Production: true,
		Url:        args[0],
		Directory:  args[1],
		Ref:        args[2],
	}
	if len(args) > 3 {
		choreoCtx.Production = false
		choreoCtx.Branch = args[3]
	}
	if err := client.Apply(ctx, choreoCtx, &choreoclient.ApplyOptions{
		Proxy: r.cfg.ToProxy(),
	}); err != nil {
		return err
	}
	return nil
}
