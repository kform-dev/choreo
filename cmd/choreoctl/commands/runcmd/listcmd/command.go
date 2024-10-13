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
	"github.com/kform-dev/choreo/pkg/client/go/choreoclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
		Use: "list [flags]",
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
	w := r.streams.Out

	ul := &unstructured.UnstructuredList{}
	ul.SetAPIVersion(choreov1alpha1.SchemeGroupVersion.Identifier())
	ul.SetKind(choreov1alpha1.SnapshotListKind)

	choreoClient := r.factory.GetChoreoClient()
	if err := choreoClient.List(ctx, ul, &choreoclient.ListOptions{
		Proxy: r.factory.GetProxy(),
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
