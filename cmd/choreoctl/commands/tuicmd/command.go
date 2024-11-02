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

package tuicmd

import (
	"context"

	view "github.com/kform-dev/choreo/cmd/tui/view2"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func NewCmdTUI(f util.Factory) *cobra.Command {
	flags := NewTUIFlags()

	cmd := &cobra.Command{
		Use:   "tui",
		Short: "tui resource",
		Args:  cobra.ExactArgs(0),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := flags.ToOptions(cmd, f)
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

type TUIFlags struct {
	tuiflags *tuiFlags
}

func NewTUIFlags() *TUIFlags {
	return &TUIFlags{
		tuiflags: newtuiflags(),
	}
}

func (r *TUIFlags) AddFlags(cmd *cobra.Command) {
	r.tuiflags.AddFlags(cmd.Flags())
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *TUIFlags) ToOptions(cmd *cobra.Command, f util.Factory) (*TUIOptions, error) {
	options := &TUIOptions{
		Factory:   f,
		Frequency: *r.tuiflags.frequency,
	}
	return options, nil
}

type TUIOptions struct {
	Factory   util.Factory
	Frequency float64
}

func (r *TUIOptions) Validate(args []string) error {
	return nil
}

func (r *TUIOptions) Run(ctx context.Context, args []string) error {
	app := view.NewApp(r.Factory)
	if err := app.Init(ctx); err != nil {
		return err
	}
	return app.Run(ctx)
}
