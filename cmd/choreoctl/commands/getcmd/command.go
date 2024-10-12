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
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

func GetCommand(ctx context.Context, f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	return NewRunner(ctx, f, streams).Command
}

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, f util.Factory, streams *genericclioptions.IOStreams) *Runner {
	r := &Runner{
		factory: f,
		streams: streams,
	}
	cmd := &cobra.Command{
		Use:   "get <RESOURCE> <NAME> <NAMESPACE>",
		Short: "get resource",
		Args:  cobra.MinimumNArgs(1),
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

	o := &Options{
		Factory: r.factory,
		Streams: r.streams,
		Output:  cmd.Flags().Lookup(genericclioptions.FlagOutputFormat).Value.String(),
	}
	if err := o.Complete(ctx); err != nil {
		return err
	}
	if err := o.Validate(ctx); err != nil {
		return err
	}
	return o.Run(ctx, args)
}

type Options struct {
	Factory util.Factory
	Streams *genericclioptions.IOStreams
	Output  string
	Branch  string
}

// Complete adapts from the command line args and validates them
func (r *Options) Complete(ctx context.Context) error {
	return nil
}

func (r *Options) Validate(ctx context.Context) error {
	return nil
}

func (r *Options) Run(ctx context.Context, args []string) error {
	// args is always len == 1
	parts := strings.SplitN(args[0], ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expecting <resource>.<group>, got: %v", parts)
	}
	proxy := r.Factory.GetProxy()
	branch := r.Factory.GetBranch()
	gvk, err := r.Factory.GetResourceMapper().KindFor(ctx, schema.GroupResource{Group: parts[1], Resource: parts[0]}, proxy, branch)
	if err != nil {
		return err
	}
	if len(args) == 1 {
		ul := &unstructured.UnstructuredList{}
		ul.SetGroupVersionKind(gvk)
		if err := r.Factory.GetResourceClient().List(ctx, ul, &resourceclient.ListOptions{
			ExprSelector:      &resourcepb.ExpressionSelector{},
			ShowManagedFields: true,
			Origin:            "choreoctl",
			Branch:            r.Branch,
		}); err != nil {
			return err
		}
		return r.parseOutput(ul)

	}
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	if err := r.Factory.GetResourceClient().Get(ctx, types.NamespacedName{
		Namespace: "default",
		Name:      args[1],
	}, u, &resourceclient.GetOptions{
		ShowManagedFields: true,
		Origin:            "choreoctl",
		Branch:            r.Branch,
	}); err != nil {
		return err
	}

	return r.parseOutput(u)
}

func (r *Options) parseOutput(obj runtime.Unstructured) error {
	w := r.Streams.Out
	switch r.Output {
	case "completion":
		if obj.IsList() {
			us := []*unstructured.Unstructured{}
			obj.EachListItem(func(o runtime.Object) error {
				u, _ := o.(*unstructured.Unstructured)
				us = append(us, u)
				return nil
			})
			sort.Slice(us, func(i, j int) bool {
				return us[i].GetName() < us[j].GetName()
			})

			var errm error
			for _, u := range us {
				if _, err := fmt.Fprintf(w, "%s\n", u.GetName()); err != nil {
					errm = errors.Join(errm, err)
				}
			}
			return errm
		}
		return fmt.Errorf("expection a list for completions")
	case "": // no output
		if obj.IsList() {
			us := []*unstructured.Unstructured{}
			obj.EachListItem(func(o runtime.Object) error {
				u, _ := o.(*unstructured.Unstructured)
				us = append(us, u)
				return nil
			})

			sort.Slice(us, func(i, j int) bool {
				return us[i].GetName() < us[j].GetName()
			})

			var errm error
			for _, u := range us {
				if _, err := fmt.Fprintf(w, "%s.%s %s\n", u.GetKind(), u.GetAPIVersion(), u.GetName()); err != nil {
					errm = errors.Join(errm, err)
				}
			}

			return errm
		}
		u := &unstructured.Unstructured{
			Object: obj.UnstructuredContent(),
		}
		if _, err := fmt.Fprintf(w, "%s.%s %s\n", u.GetKind(), u.GetAPIVersion(), u.GetName()); err != nil {
			return err
		}
		return nil
	case "yaml":
		b, err := yaml.Marshal(obj.UnstructuredContent())
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%s\n", string(b)); err != nil {
			return err
		}
		return nil
	case "json":
		b, err := json.MarshalIndent(obj.UnstructuredContent(), "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%s\n", string(b)); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("invalid output, supported json or yaml, got: %s", r.Output)
	}

}
