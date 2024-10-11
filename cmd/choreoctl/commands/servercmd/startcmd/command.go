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
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/server/choreo"
	"github.com/kform-dev/choreo/pkg/server/grpcserver"
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
		Use: "start [DIRECTORY] [flags]",
		//Args: cobra.ExactArgs(1),
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

	choreo := choreo.New(r.ConfigFlags)
	// build grpc server which hosts the apiserver storage
	grpcserver := grpcserver.New(&grpcserver.Config{
		Name:   "choreoServer",
		Flags:  r.ConfigFlags,
		Choreo: choreo,
	})

	go func() {
		err := grpcserver.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()

	time.Sleep(1 * time.Second)
	if !health.IsServerReady(ctx, &health.Config{Address: *r.ConfigFlags.Address}) {
		return fmt.Errorf("server is not ready")
	}

	if len(args) > 0 {
		path, err := fsys.NormalizeDir(args[0])
		if err != nil {
			return err
		}

		if _, err := choreo.Apply(ctx, &choreopb.Apply_Request{
			ChoreoContext: &choreopb.ChoreoContext{Path: path},
		}); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	go choreo.Start(ctx)
	defer cancel()

	<-ctx.Done()
	log.Debug("context concelled")
	return nil
}

/*

choreo, err := choreo.New(ctx, r.path, r.flags)
	if err != nil {
		return err
	}
	// build grpc server which hosts the apiserver storage
	grpcserver := grpcserver.New(&grpcserver.Config{
		Name:   r.serverName,
		Flags:  r.flags,
		Choreo: choreo,
	})
	// start the server
	go func() {
		err := grpcserver.Run(ctx)
		if err != nil {
			log.Error("grpc server failed", "err", err)
		}
	}()
	if !health.IsServerReady(ctx, r.flags) {
		return fmt.Errorf("server is not ready")
	}

	//r.repo.AddResourceClient(client)
	ctx, cancel := context.WithCancel(ctx)
	go choreo.Start(ctx)
	defer cancel()

	<-ctx.Done()
	log.Debug("context concelled")
	return nil

*/
