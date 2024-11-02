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

func NewCmdDiff(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewDiffFlags()

	cmd := &cobra.Command{
		Use:   "diff [flags]",
		Short: "show diff between snapshots",
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

type DiffFlags struct {
	RunOuput *genericclioptions.RunOutputFlags
}

// The defaults are determined here
func NewDiffFlags() *DiffFlags {
	return &DiffFlags{
		RunOuput: genericclioptions.NewRunOutputFlags(),
	}
}

// AddFlags add flags tp the command
func (r *DiffFlags) AddFlags(cmd *cobra.Command) {
	r.RunOuput.AddFlags(cmd.Flags())
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *DiffFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*DiffOptions, error) {
	options := &DiffOptions{
		Factory:           f,
		Streams:           streams,
		ShowChoreoAPIs:    *r.RunOuput.ShowChoreoAPIs,
		ShowManagedFields: *r.RunOuput.ShowManagedFields,
		ShowDiffDetails:   *r.RunOuput.ShowDiffDetails,
	}
	return options, nil
}

type DiffOptions struct {
	Factory           util.Factory
	Streams           *genericclioptions.IOStreams
	ShowChoreoAPIs    bool
	ShowManagedFields bool
	ShowDiffDetails   bool
}

func (r *DiffOptions) Validate(args []string) error {
	return nil
}

func (r *DiffOptions) Run(ctx context.Context, args []string) error {
	w := r.Streams.Out

	snapshotClient := r.Factory.GetSnapshotClient()
	b, err := snapshotClient.Diff(ctx, &snapshotclient.DiffOptions{
		Proxy:             r.Factory.GetProxy(),
		ShowManagedFields: r.ShowManagedFields,
		ShowChoreoAPIs:    r.ShowChoreoAPIs,
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
		if r.ShowDiffDetails && diffItem.Diff != nil {
			if _, err := fmt.Fprintf(w, "%s\n", *diffItem.Diff); err != nil {
				errm = errors.Join(errm, err)
			}
		}
	}

	return errm
}
