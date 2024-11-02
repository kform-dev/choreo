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

// NewCmdApply returns a cobra command.
func NewCmdStart(cfg *genericclioptions.ChoreoConfig) *cobra.Command {
	flags := NewStartFlags()

	cmd := &cobra.Command{
		Use:   "start [DIRECTORY] [flags]",
		Short: "start server",
		//Args:  cobra.ExactArgs(1),
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

type StartFlags struct {
}

// The defaults are determined here
func NewStartFlags() *StartFlags {
	return &StartFlags{}
}

// AddFlags add flags to the command
func (r *StartFlags) AddFlags(cmd *cobra.Command) {
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *StartFlags) ToOptions(cmd *cobra.Command, cfg *genericclioptions.ChoreoConfig) (*StartOptions, error) {
	options := &StartOptions{
		cfg:        cfg,
		ServerName: cmd.Flags().Lookup(genericclioptions.FlagServerName).Value.String(),
	}
	return options, nil
}

type StartOptions struct {
	cfg        *genericclioptions.ChoreoConfig
	ServerName string
}

func (r *StartOptions) Validate(args []string) error {
	return nil
}

func (r *StartOptions) Run(ctx context.Context, args []string) error {
	log := log.FromContext(ctx)

	choreo := choreo.New(r.cfg)
	// build grpc server which hosts the apiserver storage
	grpcserver := grpcserver.New(&grpcserver.Config{
		Name:   r.ServerName,
		Cfg:    r.cfg,
		Choreo: choreo,
	})

	go func() {
		err := grpcserver.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()

	time.Sleep(1 * time.Second)
	if !health.IsServerReady(ctx, &health.Config{Address: *r.cfg.ChoreoFlags.Address}) {
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
