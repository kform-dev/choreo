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
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/snapshotclient"
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
		Use: "diff [flags]",
		//Args: cobra.ExactArgs(1),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: r.runE,
	}

	// Adding a boolean flag named "show-internal" with a default value of false and description
	cmd.Flags().BoolVarP(&r.showChoreoAPIs, "show-choreoAPIs", "i", false, "Enable displaying internal choreo api resources")
	cmd.Flags().BoolVarP(&r.showManagedFields, "show-managedFields", "m", false, "Enable displaying managedFields")
	cmd.Flags().BoolVarP(&r.showDetails, "show-details", "a", false, "Enable showing details on diff")

	r.Command = cmd
	return r
}

type Runner struct {
	Command           *cobra.Command
	factory           util.Factory
	streams           *genericclioptions.IOStreams
	showChoreoAPIs    bool
	showManagedFields bool
	showDetails       bool
}

func (r *Runner) runE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	w := r.streams.Out

	snapshotClient := r.factory.GetSnapshotClient()
	b, err := snapshotClient.Diff(ctx, &snapshotclient.DiffOptions{
		Proxy:             r.factory.GetProxy(),
		ShowManagedFields: r.showManagedFields,
		ShowChoreoAPIs:    r.showChoreoAPIs,
	})
	if err != nil {
		return err
	}

	diff := &choreov1alpha1.Diff{}
	if err := json.Unmarshal(b, diff); err != nil {
		return err
	}

	diffItems := diff.Status.Items
	if len(diffItems) == 0 {
		return nil
	}
	sort.Slice(diffItems, func(i, j int) bool {
		return fmt.Sprintf("%s.%s", diffItems[i].GetGVK().String(), diffItems[i].Name) <
			fmt.Sprintf("%s.%s", diffItems[j].GetGVK().String(), diffItems[j].Name)
	})

	var errm error
	for _, diffItem := range diffItems {
		if _, err := fmt.Fprintf(w, "%s %s %s\n", diffItem.GetStatusSymbol(), diffItem.GetGVK().String(), diffItem.Name); err != nil {
			errm = errors.Join(errm, err)
		}
		if r.showDetails && diffItem.Diff != nil {
			if _, err := fmt.Fprintf(w, "%s\n", *diffItem.Diff); err != nil {
				errm = errors.Join(errm, err)
			}
		}
	}

	return errm
}
