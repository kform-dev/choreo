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

package listcmd

import (
	"context"
	"errors"
	"fmt"
	"sort"

	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/snapshotclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func NewCmdList(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewListFlags()

	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "list snapshots",
		Args:  cobra.ExactArgs(1),
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

type ListFlags struct {
}

// The defaults are determined here
func NewListFlags() *ListFlags {
	return &ListFlags{}
}

// AddFlags add flags tp the command
func (r *ListFlags) AddFlags(cmd *cobra.Command) {
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *ListFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*ListOptions, error) {
	options := &ListOptions{
		Factory: f,
		Streams: streams,
	}
	return options, nil
}

type ListOptions struct {
	Factory util.Factory
	Streams *genericclioptions.IOStreams
}

func (r *ListOptions) Validate(args []string) error {
	return nil
}

func (r *ListOptions) Run(ctx context.Context, args []string) error {
	w := r.Streams.Out

	ul := &unstructured.UnstructuredList{}
	ul.SetAPIVersion(choreov1alpha1.SchemeGroupVersion.Identifier())
	ul.SetKind(choreov1alpha1.SnapshotListKind)

	snapshotClient := r.Factory.GetSnapshotClient()
	if err := snapshotClient.List(ctx, ul, &snapshotclient.ListOptions{
		Proxy: r.Factory.GetProxy(),
	}); err != nil {
		return err
	}
	us := []*unstructured.Unstructured{}
	ul.EachListItem(func(o runtime.Object) error {
		u, _ := o.(*unstructured.Unstructured)
		us = append(us, u)
		return nil
	})

	sort.Slice(us, func(i, j int) bool {
		return us[i].GetCreationTimestamp().Time.After(us[j].GetCreationTimestamp().Time)
	})

	var errm error
	for _, u := range us {
		if _, err := fmt.Fprintf(w, "%s %s\n", u.GetName(), u.GetCreationTimestamp()); err != nil {
			errm = errors.Join(errm, err)
		}
	}
	return errm
}
