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
	"io"
	"sort"
	"strings"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/runnerclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func NewCmdOnce(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewOnceFlags()

	cmd := &cobra.Command{
		Use:   "once [flags]",
		Short: "run once",
		//Args:  cobra.ExactArgs(1),
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

type OnceFlags struct {
	RunResultsFlags *genericclioptions.RunResultFlags
}

// The defaults are determined here
func NewOnceFlags() *OnceFlags {
	return &OnceFlags{
		RunResultsFlags: genericclioptions.NewRunResultFlags(),
	}
}

// AddFlags add flags tp the command
func (r *OnceFlags) AddFlags(cmd *cobra.Command) {
	r.RunResultsFlags.AddFlags(cmd.Flags())
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *OnceFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*OnceOptions, error) {
	options := &OnceOptions{
		Factory:            f,
		Streams:            streams,
		ResultOutputFormat: *r.RunResultsFlags.ResultOutputFormat,
	}
	return options, nil
}

type OnceOptions struct {
	Factory            util.Factory
	Streams            *genericclioptions.IOStreams
	ResultOutputFormat string
}

func (r *OnceOptions) Validate(args []string) error {
	if !genericclioptions.SupportedResultOutputFormats.Has(r.ResultOutputFormat) {
		return fmt.Errorf("unsupported result output format %s, supported output formats %v", r.ResultOutputFormat, sets.List(genericclioptions.SupportedResultOutputFormats))
	}
	return nil
}

func (r *OnceOptions) Run(ctx context.Context, args []string) error {
	w := r.Streams.Out

	runnerClient := r.Factory.GetRunnerClient()
	stream, err := runnerClient.Once(ctx, &runnerclient.OnceOptions{
		Proxy: r.Factory.GetProxy(),
	})
	if err != nil {
		return err
	}
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			break // Stream is closed by the server
		}
		if err != nil {
			return err
		}
		switch rsp.Type {
		case runnerpb.Once_PROGRESS_UPDATE:
			fmt.Fprintf(w, "%s\n", rsp.GetProgressUpdate().Message)
		case runnerpb.Once_ERROR:
			fmt.Fprintf(w, "%s\n", rsp.GetError().Message)
			stream.CloseSend()
			return nil
		case runnerpb.Once_RUN_RESPONSE:
			r.handleRunResponse(rsp.GetRunResponse())
		case runnerpb.Once_SDC_RESPONSE:
			fmt.Fprintf(w, "%s\n", "to be updated")
		case runnerpb.Once_COMPLETED:
			fmt.Fprintf(w, "%s\n", "completed")
			stream.CloseSend()
			return nil
		}
	}
	return nil
}

func (r *OnceOptions) handleRunResponse(rsp *runnerpb.Once_RunResponse) {
	if !rsp.Success {
		// failed
		fmt.Println("execution failed")
		for _, result := range rsp.Results {
			if !result.Success {
				fmt.Println("  reason", "task", result.TaskId, "message", result.Message)
			}
		}
		return
	}
	var p SummaryPrinter
	switch r.ResultOutputFormat {
	case "reconciler":
		p = NewReconcilerPrinter()
	case "resource":
		p = NewResourcePrinter()
	case "raw":
		p = NewRawPrinter()
	default:
		return
	}
	// first calculate overall max idth
	for _, result := range rsp.Results {
		p.CollectData(result)
	}
	// print the result using the overall max width
	for _, result := range rsp.Results {
		fmt.Printf("Run %s summary\n", result.ReconcilerRunner)
		if len(result.Results) > 0 {
			fmt.Println("execution success, time(msec)", result.Results[len(result.Results)-1].EventTime.AsTime().Sub(result.Results[0].EventTime.AsTime()))
		}
		p.CollectData(result)
		p.PrintSummary()
	}

}

type SummaryPrinter interface {
	CollectData(result *runnerpb.Once_RunResult)
	PrintSummary()
}

type BasePrinter struct {
	maxWidths []int
	header    []string
	rows      [][]string
}

func (bp *BasePrinter) updateMaxWidths() {
	for _, row := range bp.rows {
		for i, value := range row {
			if len(value) > bp.maxWidths[i] {
				bp.maxWidths[i] = len(value)
			}
		}
	}
}

func (bp *BasePrinter) PrintSummary() {
	headerFormat := ""
	for i := range bp.header {
		headerFormat += fmt.Sprintf("%%-%ds ", bp.maxWidths[i])
	}
	headerFormat = strings.TrimSpace(headerFormat) + "\n"

	fmt.Printf(headerFormat, interfaceSlice(bp.header)...)
	for _, row := range bp.rows {
		fmt.Printf(headerFormat, interfaceSlice(row)...)
	}
}

type ReconcilerPrinter struct {
	BasePrinter
}

func NewReconcilerPrinter() SummaryPrinter {
	header := []string{"Reconciler", "Start", "Stop", "Requeue", "Error"}
	maxWidths := make([]int, len(header))
	for i, head := range header {
		maxWidths[i] = len(head)
	}
	return &ReconcilerPrinter{BasePrinter{header: header, maxWidths: maxWidths}}
}

func (rp *ReconcilerPrinter) CollectData(result *runnerpb.Once_RunResult) {
	rp.rows = [][]string{}
	reconcilerOperations := calculateReconcilerSummary(result)
	for name, operations := range reconcilerOperations {
		row := []string{
			name,
			fmt.Sprint(operations[runnerpb.Operation_START]),
			fmt.Sprint(operations[runnerpb.Operation_STOP]),
			fmt.Sprint(operations[runnerpb.Operation_REQUEUE]),
			fmt.Sprint(operations[runnerpb.Operation_ERROR]),
		}
		rp.rows = append(rp.rows, row)
	}
	rp.updateMaxWidths()

	sort.Slice(rp.rows, func(i, j int) bool {
		return strings.ToLower(fmt.Sprintf("%q.%q", rp.rows[i][0], rp.rows[i][1])) < strings.ToLower(fmt.Sprintf("%q.%q", rp.rows[j][0], rp.rows[j][1]))
	})
}

type ResourcePrinter struct {
	BasePrinter
}

func NewResourcePrinter() SummaryPrinter {
	header := []string{"Reconciler", "Resource", "Start", "Stop", "Requeue", "Error"}
	maxWidths := make([]int, len(header))
	for i, head := range header {
		maxWidths[i] = len(head)
	}
	return &ResourcePrinter{BasePrinter{header: header, maxWidths: maxWidths}}
}

func (rp *ResourcePrinter) CollectData(result *runnerpb.Once_RunResult) {
	resourceOperations := calculateReconcilerResourceSummary(result)
	for res, ops := range resourceOperations {
		row := []string{
			res.Reconcilername,
			res.ResourceNameString(),
			fmt.Sprint(ops[runnerpb.Operation_START]),
			fmt.Sprint(ops[runnerpb.Operation_STOP]),
			fmt.Sprint(ops[runnerpb.Operation_REQUEUE]),
			fmt.Sprint(ops[runnerpb.Operation_ERROR]),
		}
		rp.rows = append(rp.rows, row)
	}
	rp.updateMaxWidths()
	sort.Slice(rp.rows, func(i, j int) bool {
		return strings.ToLower(fmt.Sprintf("%q.%q", rp.rows[i][0], rp.rows[i][1])) < strings.ToLower(fmt.Sprintf("%q.%q", rp.rows[j][0], rp.rows[j][1]))
	})
}

type RawPrinter struct {
	BasePrinter
}

func NewRawPrinter() SummaryPrinter {
	header := []string{"EventTime", "Reconciler", "Resource", "Operation", "Message"}
	maxWidths := make([]int, len(header))
	for i, head := range header {
		maxWidths[i] = len(head)
	}
	return &RawPrinter{BasePrinter{header: header, maxWidths: maxWidths}}
}

func (rp *RawPrinter) CollectData(result *runnerpb.Once_RunResult) {
	timeFormat := "2006-01-02 15:04:05.000000 UTC"
	rp.rows = [][]string{}
	for _, result := range result.Results {
		row := []string{
			result.EventTime.AsTime().Format(timeFormat),
			result.ReconcilerName,
			getReconcilerResource(result).ResourceNameString(),
			result.Operation.String(),
			result.Message,
		}
		rp.rows = append(rp.rows, row)
	}
	rp.updateMaxWidths()
}

type Operations map[runnerpb.Operation]int

func NewOperations() Operations {
	return map[runnerpb.Operation]int{
		runnerpb.Operation_START:   0,
		runnerpb.Operation_STOP:    0,
		runnerpb.Operation_REQUEUE: 0,
		runnerpb.Operation_ERROR:   0,
	}
}

func calculateReconcilerSummary(rsp *runnerpb.Once_RunResult) map[string]Operations {
	reconcilers := getReconcilers(rsp)
	reconcilerOperations := make(map[string]Operations, len(reconcilers))
	for _, result := range rsp.Results {
		if _, exists := reconcilerOperations[result.ReconcilerName]; !exists {
			reconcilerOperations[result.ReconcilerName] = NewOperations()
		}
		reconcilerOperations[result.ReconcilerName][result.Operation]++
	}
	return reconcilerOperations
}

func getReconcilers(rsp *runnerpb.Once_RunResult) []string {
	reconcilers := []string{}
	for _, result := range rsp.Results {
		reconcilers = append(reconcilers, result.ReconcilerName)
	}
	sort.Strings(reconcilers)
	return reconcilers
}

type ReconcilerResource struct {
	Reconcilername string
	Group          string
	Kind           string
	Namespace      string
	Name           string
}

func getReconcilerResource(result *runnerpb.ReconcileResult) ReconcilerResource {
	return ReconcilerResource{
		Reconcilername: result.ReconcilerName,
		Group:          result.Resource.Group,
		Kind:           result.Resource.Kind,
		Name:           result.Resource.Name,
		Namespace:      result.Resource.Namespace,
	}
}

func (r ReconcilerResource) ResourceNameString() string {
	return fmt.Sprintf("%s.%s.%s.%s", r.Group, r.Kind, r.Namespace, r.Name)
}

func calculateReconcilerResourceSummary(rsp *runnerpb.Once_RunResult) map[ReconcilerResource]Operations {
	//reconcilerResourceSet := getReconcilerResourceSet(rsp)
	reconcilerOperations := make(map[ReconcilerResource]Operations, 0)
	for _, result := range rsp.Results {
		reconcilerResource := getReconcilerResource(result)
		if _, exists := reconcilerOperations[reconcilerResource]; !exists {
			reconcilerOperations[reconcilerResource] = NewOperations()
		}
		reconcilerOperations[reconcilerResource][result.Operation]++
	}
	return reconcilerOperations
}

// Converts a slice of strings to a slice of interfaces for formatting purposes
func interfaceSlice(slice []string) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}

/*

stream, err := client.Once(...)
if err != nil {
    // Handle error
}
for {
    response, err := stream.Recv()
    if err == io.EOF {
        break // Stream is closed by the server
    }
    if err != nil {
        // Handle error
    }

    switch response.Type {
    case Once.MessageType.PROGRESS_UPDATE:
        fmt.Println("Progress:", response.GetProgressUpdate().Message)
    case Once.MessageType.ERROR:
        fmt.Println("Error:", response.GetError().Message)
    case Once.MessageType.RUN_RESPONSE:
        handleRunResponse(response.GetRunResponse())
    case Once.MessageType.SDC_RESPONSE:
        fmt.Println("SDC Message:", response.GetSdcResponse().Message)
    case Once.MessageType.COMPLETED:
        fmt.Println("Completed:", response.GetCompleted().Message)
    }
}
*/
