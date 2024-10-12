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
		Use:  "diff SRC_BRANCHNAME DST_BRANCHNAME [flags]",
		Args: cobra.ExactArgs(2),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: r.runE,
	}

	r.Command = cmd
	return r
}

type Runner struct {
	Command *cobra.Command
	factory util.Factory
	streams *genericclioptions.IOStreams
}

func (r *Runner) runE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	branchClient := r.factory.GetBranchClient()
	srcBranchName := args[0]
	dstBranchName := args[1]
	diffs, err := branchClient.Diff(ctx, srcBranchName, dstBranchName, &branchclient.DiffOptions{
		Proxy: r.factory.GetProxy(),
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

		if _, err := fmt.Fprintf(r.streams.Out, format, getAction(diff.Action), diff.SrcFileName, diff.DstFileName); err != nil {
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
