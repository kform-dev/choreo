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

	"github.com/kform-dev/choreo/cmd/tui/view2"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func GetCommand(ctx context.Context, f util.Factory) *cobra.Command {
	return NewRunner(ctx, f).Command
}

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, f util.Factory) *Runner {
	r := &Runner{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "tui resource",
		//Args:  cobra.ExactArgs(1),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: r.runE,
	}

	cmd.Flags().Float64VarP(&r.frequency, "frequency", "f", 3.0, "refresh frequency in seconds")

	r.Command = cmd
	return r
}

type Runner struct {
	Command   *cobra.Command
	factory   util.Factory
	frequency float64
	//streams *genericclioptions.IOStreams
}

func (r *Runner) runE(cmd *cobra.Command, args []string) error {
	app := view.NewApp(r.factory)
	if err := app.Init(cmd.Context()); err != nil {
		return err
	}
	return app.Run(cmd.Context())
}
