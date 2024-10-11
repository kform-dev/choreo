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

package getcmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/branchclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func GetCommand(ctx context.Context, f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	return NewRunner(f, streams).Command
}

// NewRunner returns a command runner.
func NewRunner(f util.Factory, streams *genericclioptions.IOStreams) *Runner {
	r := &Runner{
		factory: f,
		streams: streams,
	}
	cmd := &cobra.Command{
		Use:  "get BRANCHNAME [flags]",
		Args: cobra.MaximumNArgs(1),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: r.runE,
	}

	cmd.Flags().BoolVarP(&r.files, "files", "f", false, "get files")

	r.Command = cmd
	return r
}

type Runner struct {
	Command *cobra.Command
	factory util.Factory
	streams *genericclioptions.IOStreams
	files   bool
}

func (r *Runner) runE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	branchClient := r.factory.GetBranchClient()
	if len(args) == 1 {
		branchName := args[0]
		if r.files {
			rspch := branchClient.StreamFiles(ctx, branchName)

			for {
				select {
				case <-ctx.Done():
					return nil
				case file, ok := <-rspch:
					if !ok {
						return nil
					}
					if _, err := fmt.Fprintf(r.streams.Out, "%s:\n%s", file.Name, file.Data); err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		logs, err := branchClient.Get(ctx, branchName, branchclient.GetOptions{})
		if err != nil {
			return err
		}
		var errm error
		for _, log := range logs {
			//logEntry := fmt.Sprintf("commit %s\nAuthor: %s <%s>\nDate:   %s\n\n\t%s\n",
			//	log.CommitHash, log.AuthorName, log.AuthorEmail, log.Date, log.Message)

			logEntry := fmt.Sprintf("%scommit %s\n%sAuthor: %s <%s>\n%sDate:   %s\n%s\n\t%s\n%s",
				colorGreen, log.CommitHash,
				colorReset, log.AuthorName, log.AuthorEmail,
				colorReset, log.Date,
				colorReset, log.Message,
				colorReset)

			if _, err := fmt.Fprintf(r.streams.Out, "%s\n", logEntry); err != nil {
				errm = errors.Join(errm, err)
			}
		}
		return errm
	}

	branches, err := branchClient.List(ctx, branchclient.ListOptions{
		Choreo: "network",
	})
	if err != nil {
		return err
	}

	var errm error
	for _, branch := range branches {
		if _, err := fmt.Fprintf(r.streams.Out, "%s\n", branch); err != nil {
			errm = errors.Join(errm, err)
		}
	}
	return errm

}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)
