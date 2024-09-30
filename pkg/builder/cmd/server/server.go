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

package server

import (
	"context"
	"flag"

	"github.com/kform-dev/choreo/pkg/builder/cmd/server/options"
	"github.com/spf13/cobra"
)

// NewCommand provides a CLI handler for 'launch apiserver' command
// with a default ServerOptions.
func NewCommand(ctx context.Context, opts *options.ChoreoOptions) *cobra.Command {
	o := *opts
	cmd := &cobra.Command{
		Short: "launch Choreo apiserver",
		Long:  "launch Choreo apiserver",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(ctx); err != nil {
				return err
			}
			if err := o.Run(ctx); err != nil {
				return err
			}
			return nil
		},
	}
	flags := cmd.PersistentFlags()
	o.Flags().AddServerControllerFlags(flags)
	flags.AddGoFlagSet(flag.CommandLine)

	// flags are generically added

	return cmd
}
