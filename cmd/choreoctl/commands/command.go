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

package commands

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/kform-dev/choreo/cmd/choreoctl/commands/apiresourcescmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/applycmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/branchcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/deletecmd.go"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/depscmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/devcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/getcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/runcmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/servercmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/tuicmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/completion"
	"github.com/kform-dev/choreo/cmd/choreoctl/globals"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
)

func GetMain(ctx context.Context) (*cobra.Command, util.Factory) {
	//showVersion := false
	cmd := &cobra.Command{
		Use:          "choreoctl",
		Short:        "choreoctl is client tool to intercat with choreo",
		Long:         "choreoctl is client tool to intercat with choreo",
		SilenceUsage: true,
		// We handle all errors in main after return from cobra so we can
		// adjust the error message coming from libraries
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				return err
			}
			if debug {
				globals.LogLevel.Set(slog.LevelDebug)
			}

			return genericclioptions.InitConfig(cmd)
		},
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
	cmd.SetContext(ctx)
	// choreo flags
	flags := cmd.PersistentFlags()
	choreoFlags := genericclioptions.NewConfigFlags()
	// server
	choreoFlags.AddServerControllerFlags(flags)
	// client
	choreoFlags.AddClientFlags(flags)

	flags.AddGoFlagSet(flag.CommandLine)

	f, err := util.NewFactory(choreoFlags)
	if err != nil {
		panic(err)
	}
	streams := &genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	registerCompletionFuncForGlobalFlags(cmd, f)

	subCmds := map[string]*cobra.Command{
		"apiresources": apiresourcescmd.GetCommand(ctx, f, streams),
		"get":          getcmd.NewCmdGet(f, streams),
		"apply":        applycmd.NewCmdApply(f, streams),
		"delete":       deletecmd.NewCmdDelete(f, streams),
		"deps":         depscmd.GetCommand(ctx, f, streams),
		"branch":       branchcmd.GetCommand(ctx, f, streams),
		"run":          runcmd.GetCommand(ctx, f, streams),
		"tui":          tuicmd.GetCommand(ctx, f),
		"server":       servercmd.GetCommand(ctx, choreoFlags),
		"dev":          devcmd.GetCommand(ctx, choreoFlags),
	}

	for cmdName, subCmd := range subCmds {
		if cmdName == "get" || cmdName == "delete" {
			// required to avoid import cycle
			subCmd.ValidArgsFunction = (&completion.Completion{Factory: f}).ResourceTypeAndNameCompletionFunc()
		}
		cmd.AddCommand(subCmd)
	}
	cmd.AddCommand(GetVersionCommand(ctx))
	return cmd, f
}

type Runner struct {
	Command *cobra.Command
}

func registerCompletionFuncForGlobalFlags(cmd *cobra.Command, f util.Factory) error {
	if err := cmd.RegisterFlagCompletionFunc(
		"branch",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return (&completion.Completion{Factory: f}).CompBranch(cmd, toComplete), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		return err
	}
	return nil
}
