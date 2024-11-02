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

package getcmd

import (
	"context"
	"fmt"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/choreoclient"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

// NewCmdApply returns a cobra command.
func NewCmdGet(cfg *genericclioptions.ChoreoConfig) *cobra.Command {
	flags := NewGetFlags()

	cmd := &cobra.Command{
		Use:   "get [flags]",
		Short: "get server status",
		//Args:  cobra.ExactArgs(0),
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

type GetFlags struct {
}

// NewApplyFlags determines which flags will be added to the command
// The defaults are determined here
func NewGetFlags() *GetFlags {
	return &GetFlags{}
}

// AddFlags add flags to the command
func (r *GetFlags) AddFlags(cmd *cobra.Command) {
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *GetFlags) ToOptions(cmd *cobra.Command, cfg *genericclioptions.ChoreoConfig) (*GetOptions, error) {
	options := &GetOptions{
		cfg: cfg,
	}
	return options, nil
}

type GetOptions struct {
	cfg *genericclioptions.ChoreoConfig
}

func (r *GetOptions) Validate(args []string) error {
	return nil
}

func (r *GetOptions) Run(ctx context.Context, args []string) error {
	client, err := r.cfg.ToChoreoClient()
	if err != nil {
		return err
	}
	defer client.Close()

	rsp, err := client.Get(ctx, &choreoclient.GetOptions{
		Proxy: r.cfg.ToProxy(),
	})
	if err != nil {
		return err
	}
	fmt.Println("choreoctx", rsp.ChoreoContext)
	fmt.Println("status", rsp.Status)
	fmt.Println("reason", rsp.Reason)
	return nil
}
