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
	"github.com/kform-dev/choreo/pkg/server/choreo/loader"
	"github.com/kform-dev/kform/pkg/fsys"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func GetCommand(ctx context.Context, flags *genericclioptions.ConfigFlags) *cobra.Command {
	return NewRunner(ctx, flags).Command
}

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, flags *genericclioptions.ConfigFlags) *Runner {
	r := &Runner{
		flags: flags,
	}
	cmd := &cobra.Command{
		Use:   "parse PATH [flags]",
		Short: "parse path",
		Args:  cobra.ExactArgs(1),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: r.runE,
	}

	r.Command = cmd
	return r
}

type Runner struct {
	Command *cobra.Command
	flags   *genericclioptions.ConfigFlags
}

func (r *Runner) runE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	path, err := fsys.NormalizeDir(args[0])
	if err != nil {
		return err
	}

	loader := loader.DevLoader{
		Path:  path,
		Flags: r.flags,
	}
	return loader.Load(ctx)
}
