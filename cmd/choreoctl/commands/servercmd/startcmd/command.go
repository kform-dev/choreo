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

package startcmd

import (
	"context"
	"fmt"
	"time"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/builder"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/server/health"
	"github.com/kform-dev/kform/pkg/fsys"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func GetCommand(ctx context.Context, flags *genericclioptions.ConfigFlags) *cobra.Command {
	return NewRunner(flags).Command
}

// NewRunner returns a command runner.
func NewRunner(flags *genericclioptions.ConfigFlags) *Runner {
	r := &Runner{
		ConfigFlags: flags,
	}
	cmd := &cobra.Command{
		Use:  "start DIRECTORY [flags]",
		Args: cobra.ExactArgs(1),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: r.runE,
	}

	r.Command = cmd
	return r
}

type Runner struct {
	Command     *cobra.Command
	ConfigFlags *genericclioptions.ConfigFlags
}

func (r *Runner) runE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	log := log.FromContext(ctx)

	path, err := fsys.NormalizeDir(args[0])
	if err != nil {
		return err
	}

	go func() {
		if err := builder.ChoreoServer.Execute(ctx, path, r.ConfigFlags); err != nil {
			log.Info("cannot start choreo-server", "error", err)
			panic(err)
		}
	}()

	time.Sleep(1 * time.Second)
	if !health.IsServerReady(ctx, r.ConfigFlags) {
		return fmt.Errorf("server is not ready")
	}
	<-ctx.Done()
	log.Debug("context concelled")
	return nil
}
