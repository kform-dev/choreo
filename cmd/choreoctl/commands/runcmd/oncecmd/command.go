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

package oncecmd

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
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
		Use: "once [flags]",
		//Args: cobra.ExactArgs(1),
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

	choreoClient := r.factory.GetChoreoClient()
	rsp, err := choreoClient.Once(ctx)
	if err != nil {
		return err
	}
	if rsp != nil {
		printResult(rsp)
	}
	return nil
}

func printResult(rsp *choreopb.Once_Response) {
	if !rsp.Success {
		// failed
		fmt.Println("execution failed", rsp.ReconcileRef, rsp.Message)
		return
	}
	fmt.Println("execution success, time(sec)", rsp.ExecutionTime)
	reconcilers := []string{}
	for name := range rsp.Results {
		reconcilers = append(reconcilers, name)
	}

	if len(reconcilers) == 0 {
		return
	}
	sort.Strings(reconcilers)

	// Calculate maximum lengths for columns
	maxNameLen := len("Reconciler")
	maxStartLen := len("Start")
	maxStopLen := len("Stop")
	maxRequeueLen := len("Requeue")
	maxErrorLen := len("Error")

	// Calculate maximum lengths
	for name, operations := range rsp.Results {
		maxNameLen = max(maxNameLen, len(name))
		maxStartLen = max(maxStartLen, digitCount(int(operations.OperationCounts[choreopb.Operation_START.String()])))
		maxStopLen = max(maxStopLen, digitCount(int(operations.OperationCounts[choreopb.Operation_STOP.String()])))
		maxRequeueLen = max(maxRequeueLen, digitCount(int(operations.OperationCounts[choreopb.Operation_REQUEUE.String()])))
		maxErrorLen = max(maxErrorLen, digitCount(int(operations.OperationCounts[choreopb.Operation_ERROR.String()])))
	}

	// Print header
	headerFormat := fmt.Sprintf("%%-%ds %%-%ds %%-%ds %%-%ds %%-%ds\n",
		maxNameLen, maxStartLen, maxStopLen, maxRequeueLen, maxErrorLen)
	fmt.Printf(headerFormat, "Reconciler", "Start", "Stop", "Requeue", "Error")

	// Print each row
	rowFormat := fmt.Sprintf("%%-%ds %%%dd %%%dd %%%dd %%%dd\n",
		maxNameLen, maxStartLen, maxStopLen, maxRequeueLen, maxErrorLen)

	for _, name := range reconcilers {
		op := rsp.Results[name]
		fmt.Printf(rowFormat,
			name,
			op.OperationCounts[choreopb.Operation_START.String()],
			op.OperationCounts[choreopb.Operation_STOP.String()],
			op.OperationCounts[choreopb.Operation_REQUEUE.String()],
			op.OperationCounts[choreopb.Operation_ERROR.String()],
		)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func digitCount(n int) int {
	if n == 0 {
		return 1
	}
	return int(math.Log10(float64(n))) + 1
}
