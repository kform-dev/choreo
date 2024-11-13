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
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/commitcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/depscmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/diffcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/listcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/loadcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/oncecmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/pushcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/resultcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/startcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd/stopcmd"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
)

func NewCmdRun(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
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
		commitcmd.NewCmdCommit(f, streams),
		depscmd.NewCmdDeps(f, streams),
		diffcmd.NewCmdDiff(f, streams),
		resultcmd.NewCmdResult(f, streams),
		listcmd.NewCmdList(f, streams),
		loadcmd.NewCmdLoad(f, streams),
		oncecmd.NewCmdOnce(f, streams),
		pushcmd.NewCmdPush(f, streams),
		startcmd.NewCmdStart(f, streams),
		stopcmd.NewCmdStop(f, streams),
	)
	return cmd
}
