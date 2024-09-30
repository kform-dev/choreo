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

package servercmd

import (
	"context"

	"github.com/kform-dev/choreo/cmd/choreoctl/commands/servercmd/startcmd"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/spf13/cobra"
)

// NewRunner returns a command runner.
func GetCommand(ctx context.Context, flags *genericclioptions.ConfigFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use: "server",
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			h, err := cmd.Flags().GetBool("help")
			if err != nil {
				return err
			}
			if h {
				return cmd.Help()
			}
			return cmd.Usage()
		},
	}

	cmd.AddCommand(
		startcmd.GetCommand(ctx, flags),
	)
	return cmd
}
