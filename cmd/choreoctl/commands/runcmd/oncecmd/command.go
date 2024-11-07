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
	"sort"
	"time"

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
	runnerClient := r.Factory.GetRunnerClient()
	rsp, err := runnerClient.Once(ctx, &runnerclient.OnceOptions{
		Proxy: r.Factory.GetProxy(),
	})
	if err != nil {
		return err
	}
	if rsp != nil {
		if !rsp.Success {
			// failed
			fmt.Println("execution failed")
			for _, result := range rsp.Results {
				if !result.Success {
					fmt.Println("  reason", "task", result.TaskId, "message", result.Message)
				}
			}
			return nil
		}
		/*
			for _, result := range rsp.Results {
				if len(result.Results) > 0 {
					fmt.Println("execution success, time(msec)", result.Results[len(result.Results)-1].EventTime.AsTime().Sub(result.Results[0].EventTime.AsTime()))
				}
				switch r.ResultOutputFormat {
				case "reconciler":
					printReconcilerResultSummary(calculateReconcilerSummary(result))
				case "resource":
					printReconcilerResourceResultSummary(calculateReconcilerResourceSummary(result))
				case "raw":
					printResultRaw(result)
				}
			}
		*/
		printResults(rsp, r.ResultOutputFormat)

	}
	return nil
}

func printResults(rsp *runnerpb.Once_Response, resultOutputFormat string) error {
	// Pre-calculate maximum widths for all results
	maxWidths := map[string]int{
		"EventTime":  len("EventTime"),
		"Reconciler": len("Reconciler"),
		"Resource":   len("Resource"),
		"Operation":  len("Operation"),
		"Message":    len("Message"),
	}

	// Accumulate rows for each type of summary
	reconcilerRows := [][]string{}
	resourceRows := [][]string{}
	rawRows := [][]string{}

	for _, result := range rsp.Results {
		if len(result.Results) > 0 {
			fmt.Println("execution success, time(msec)", result.Results[len(result.Results)-1].EventTime.AsTime().Sub(result.Results[0].EventTime.AsTime()))
		}

		switch resultOutputFormat {
		case "reconciler":
			reconcilerSummary := calculateReconcilerSummary(result)
			for name, ops := range reconcilerSummary {
				row := []string{name, fmt.Sprint(ops[runnerpb.Operation_START]), fmt.Sprint(ops[runnerpb.Operation_STOP]), fmt.Sprint(ops[runnerpb.Operation_REQUEUE]), fmt.Sprint(ops[runnerpb.Operation_ERROR])}
				reconcilerRows = append(reconcilerRows, row)
				updateMaxWidths(row, maxWidths)
			}
		case "resource":
			resourceSummary := calculateReconcilerResourceSummary(result)
			for res, ops := range resourceSummary {
				row := []string{res.Reconcilername, res.ResourceNameString(), fmt.Sprint(ops[runnerpb.Operation_START]), fmt.Sprint(ops[runnerpb.Operation_STOP]), fmt.Sprint(ops[runnerpb.Operation_REQUEUE]), fmt.Sprint(ops[runnerpb.Operation_ERROR])}
				resourceRows = append(resourceRows, row)
				updateMaxWidths(row, maxWidths)
			}
		case "raw":
			for _, res := range result.Results {
				row := []string{res.EventTime.AsTime().Format(time.RFC3339), res.ReconcilerName, getReconcilerResource(res).ResourceNameString(), res.Operation.String(), res.Message}
				rawRows = append(rawRows, row)
				updateMaxWidths(row, maxWidths)
			}
		}
	}

	// Print summaries based on the format
	if resultOutputFormat == "reconciler" {
		printSummary("Reconciler Summary", maxWidths, reconcilerRows)
	} else if resultOutputFormat == "resource" {
		printSummary("Resource Summary", maxWidths, resourceRows)
	} else if resultOutputFormat == "raw" {
		printSummary("Raw Events Summary", maxWidths, rawRows)
	}

	return nil
}

// Updates the maximum widths based on the content of each row
func updateMaxWidths(row []string, maxWidths map[string]int) {
	keys := []string{"EventTime", "Reconciler", "Resource", "Operation", "Message"}
	for i, value := range row {
		if len(value) > maxWidths[keys[i]] {
			maxWidths[keys[i]] = len(value)
		}
	}
}

// Prints summary with headers adjusted to maximum widths
func printSummary(title string, maxWidths map[string]int, rows [][]string) {
	fmt.Println(title)
	headers := []string{"EventTime", "Reconciler", "Resource", "Operation", "Message"}
	headerFmt := ""
	for _, h := range headers {
		headerFmt += fmt.Sprintf("%%-%ds ", maxWidths[h])
	}
	headerFmt += "\n"

	fmt.Printf(headerFmt, interfaceSlice(headers)...)
	for _, row := range rows {
		fmt.Printf(headerFmt, interfaceSlice(row)...)
	}
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

func calculateReconcilerSummary(rsp *runnerpb.Once_Result) map[string]Operations {
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

func getReconcilers(rsp *runnerpb.Once_Result) []string {
	reconcilers := []string{}
	for _, result := range rsp.Results {
		reconcilers = append(reconcilers, result.ReconcilerName)
	}
	sort.Strings(reconcilers)
	return reconcilers
}

func (r ReconcilerResource) ResourceNameString() string {
	return fmt.Sprintf("%s.%s.%s.%s", r.Group, r.Kind, r.Namespace, r.Name)
}

func calculateReconcilerResourceSummary(rsp *runnerpb.Once_Result) map[ReconcilerResource]Operations {
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



func printResultRaw(rsp *runnerpb.Once_Result) {
	timeFormat := "2006-01-02 15:04:05.000000 UTC"
	rows := make([][]string, 0)
	for _, result := range rsp.Results {
		row := []string{
			result.EventTime.AsTime().Format(timeFormat),
			result.ReconcilerName,
			getReconcilerResource(result).ResourceNameString(),
			result.Operation.String(),
			result.Message,
		}
		rows = append(rows, row)
	}
	printSummary("Raw Summary", []string{"EventTime", "Reconciler", "Resource", "Operation", "Message"}, rows)
}

// Example usage within your original functions
func printReconcilerResourceResultSummary(reconcilerResourceOperations map[ReconcilerResource]Operations) {
	rows := make([][]string, 0)
	for reconcilerResource, operations := range reconcilerResourceOperations {
		row := []string{
			reconcilerResource.Reconcilername,
			reconcilerResource.ResourceNameString(),
			fmt.Sprint(operations[runnerpb.Operation_START]),
			fmt.Sprint(operations[runnerpb.Operation_STOP]),
			fmt.Sprint(operations[runnerpb.Operation_REQUEUE]),
			fmt.Sprint(operations[runnerpb.Operation_ERROR]),
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(fmt.Sprintf("%q.%q", rows[i][0], rows[i][1])) < strings.ToLower(fmt.Sprintf("%q.%q", rows[j][0], rows[j][1]))
	})
	printSummary("Reconciler Resource Operations Summary", []string{"Reconciler", "resource", "Start", "Stop", "Requeue", "Error"}, rows)
}

func printReconcilerResultSummary(resourceOperations map[string]Operations) {
	rows := make([][]string, 0)
	for name, operations := range resourceOperations {
		row := []string{
			name,
			fmt.Sprint(operations[runnerpb.Operation_START]),
			fmt.Sprint(operations[runnerpb.Operation_STOP]),
			fmt.Sprint(operations[runnerpb.Operation_REQUEUE]),
			fmt.Sprint(operations[runnerpb.Operation_ERROR]),
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i][0]) < strings.ToLower(rows[j][0])
	})
	printSummary("Reconciler Operations Summary", []string{"Reconciler", "Start", "Stop", "Requeue", "Error"}, rows)
}

func printSummary(title string, headers []string, rows [][]string) {
	maxLengths := make([]int, len(headers))
	for i, header := range headers {
		maxLengths[i] = len(header)
	}

	// Determine the maximum length of each column
	for _, row := range rows {
		for i, field := range row {
			if len(field) > maxLengths[i] {
				maxLengths[i] = len(field)
			}
		}
	}

	// Prepare format string for headers and rows
	var formatBuilder strings.Builder
	for _, length := range maxLengths {
		formatBuilder.WriteString(fmt.Sprintf("%%-%ds ", length))
	}
	formatBuilder.WriteString("\n")
	format := formatBuilder.String()

	// Print title
	fmt.Println(title)

	// Print header
	fmt.Printf(format, interfaceSlice(headers)...)

	// Print rows
	for _, row := range rows {
		fmt.Printf(format, interfaceSlice(row)...)
	}
}

*/
