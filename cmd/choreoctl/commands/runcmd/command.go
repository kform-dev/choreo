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

package runcmd

import (
	"context"

	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/commitcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/diffcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/listcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/loadcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/oncecmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/pushcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/startcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/stopcmd"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
)

// NewRunner returns a command runner.
func GetCommand(ctx context.Context, f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use: "run",
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
		startcmd.GetCommand(ctx, f, streams),
		stopcmd.GetCommand(ctx, f, streams),
		oncecmd.GetCommand(ctx, f, streams),
		loadcmd.GetCommand(ctx, f, streams),
		listcmd.GetCommand(ctx, f, streams),
		diffcmd.GetCommand(ctx, f, streams),
		commitcmd.GetCommand(ctx, f, streams),
		pushcmd.GetCommand(ctx, f, streams),
	)
	return cmd
}
