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

package diffcmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/branchclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
	"github.com/spf13/cobra"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func NewCmdDiff(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewDiffFlags()

	cmd := &cobra.Command{
		Use:  "diff SRC_BRANCHNAME DST_BRANCHNAME [flags]",
		Args: cobra.ExactArgs(2),
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

type DiffFlags struct {
}

// The defaults are determined here
func NewDiffFlags() *DiffFlags {
	return &DiffFlags{}
}

// AddFlags add flags tp the command
func (r *DiffFlags) AddFlags(cmd *cobra.Command) {
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *DiffFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*DiffOptions, error) {
	options := &DiffOptions{
		Factory: f,
		Streams: streams,
	}
	return options, nil
}

type DiffOptions struct {
	Factory util.Factory
	Streams *genericclioptions.IOStreams
}

func (r *DiffOptions) Validate(args []string) error {
	return nil
}

func (r *DiffOptions) Run(ctx context.Context, args []string) error {
	branchClient := r.Factory.GetBranchClient()
	srcBranchName := args[0]
	dstBranchName := args[1]
	diffs, err := branchClient.Diff(ctx, srcBranchName, dstBranchName, &branchclient.DiffOptions{
		Proxy: r.Factory.GetProxy(),
	})
	if err != nil {
		return err
	}

	var errm error
	maxLen := 0
	// First, find the maximum length of any filename in the diffs
	for _, diff := range diffs {
		if len(diff.SrcFileName) > maxLen {
			maxLen = len(diff.SrcFileName)
		}
		if len(diff.DstFileName) > maxLen {
			maxLen = len(diff.DstFileName)
		}
	}
	//rebase := false
	for _, diff := range diffs {
		format := fmt.Sprintf(" %%s %%-%ds -> %%-%ds\n", maxLen, maxLen)

		if _, err := fmt.Fprintf(r.Streams.Out, format, getAction(diff.Action), diff.SrcFileName, diff.DstFileName); err != nil {
			errm = errors.Join(errm, err)
		}
	}
	return errm
}

func getAction(a branchpb.Diff_FileAction) string {
	switch a {
	case branchpb.Diff_ADDED:
		return "+"
	case branchpb.Diff_MODIFIED:
		return "~"
	case branchpb.Diff_DELETED:
		return "-"
	default:
		return "error"
	}

}
