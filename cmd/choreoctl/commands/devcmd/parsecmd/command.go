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

package parsecmd

import (
	"context"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

// NewCmdApply returns a cobra command.
func NewCmdParse(cfg *genericclioptions.ChoreoConfig) *cobra.Command {
	flags := NewParseFlags()

	cmd := &cobra.Command{
		Use:   "parse PATH [flags]",
		Short: "parse path",
		Args:  cobra.ExactArgs(1),
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

type ParseFlags struct{}

func NewParseFlags() *ParseFlags { return &ParseFlags{} }

func (r *ParseFlags) AddFlags(cmd *cobra.Command) {}

func (r *ParseFlags) ToOptions(cmd *cobra.Command, cfg *genericclioptions.ChoreoConfig) (*ParseOptions, error) {
	options := &ParseOptions{
		cfg: cfg,
	}
	return options, nil
}

type ParseOptions struct {
	cfg *genericclioptions.ChoreoConfig
}

func (r *ParseOptions) Validate(args []string) error {
	return nil
}

func (r *ParseOptions) Run(ctx context.Context, args []string) error {
	/*
		path, err := fsys.NormalizeDir(args[0])
		if err != nil {
			return err
		}

		loader := loader.DevLoader{
			SrcPath: path,
			DstPath: path,
			Cfg:     r.cfg,
		}
		return loader.Load(ctx)
	*/
	return nil
}
