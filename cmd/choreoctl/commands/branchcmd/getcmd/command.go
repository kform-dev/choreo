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

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/branchclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func NewCmdGet(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewGetFlags()

	cmd := &cobra.Command{
		Use:  "get BRANCHNAME [flags]",
		Args: cobra.MaximumNArgs(1),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			o, err := flags.ToOptions(cmd, f, streams)
			if err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			return o.Run(ctx, args)
		},
	}
	flags.AddFlags(cmd)
	return cmd
}

type GetFlags struct {
}

// The defaults are determined here
func NewGetFlags() *GetFlags {
	return &GetFlags{}
}

// AddFlags add flags tp the command
func (r *GetFlags) AddFlags(cmd *cobra.Command) {
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *GetFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*GetOptions, error) {
	options := &GetOptions{
		Factory: f,
		Streams: streams,
	}
	return options, nil
}

type GetOptions struct {
	Factory util.Factory
	Streams *genericclioptions.IOStreams
	Files   bool
}

func (r *GetOptions) Validate(args []string) error {
	return nil
}

func (r *GetOptions) Run(ctx context.Context, args []string) error {
	log := log.FromContext(ctx)
	branchClient := r.Factory.GetBranchClient()
	if len(args) == 1 {
		branchName := args[0]
		if r.Files {
			rspch := branchClient.StreamFiles(ctx, branchName)

			for {
				select {
				case <-ctx.Done():
					return nil
				case file, ok := <-rspch:
					if !ok {
						return nil
					}
					if _, err := fmt.Fprintf(r.Streams.Out, "%s:\n%s", file.Name, file.Data); err != nil {
						log.Error("cannot stream to output", "err", err)
					}
				}
			}
		}

		logs, err := branchClient.Get(ctx, branchName, &branchclient.GetOptions{
			Proxy: r.Factory.GetProxy(),
		})
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

			if _, err := fmt.Fprintf(r.Streams.Out, "%s\n", logEntry); err != nil {
				errm = errors.Join(errm, err)
			}
		}
		return errm
	}

	branches, err := branchClient.List(ctx, &branchclient.ListOptions{
		Proxy: r.Factory.GetProxy(),
	})
	if err != nil {
		return err
	}

	var errm error
	for _, branch := range branches {
		if _, err := fmt.Fprintf(r.Streams.Out, "%s\n", branch); err != nil {
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
