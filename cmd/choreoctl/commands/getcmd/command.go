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

func NewCmdGet(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewGetFlags()

	cmd := &cobra.Command{
		Use:   "get <RESOURCE> <NAME>",
		Short: "get resource",
		Args:  cobra.MinimumNArgs(1),
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

type GetFlags struct {
	ResourceOuput *genericclioptions.ResourceOutputFlags
}

// NewGetFlags determines which flags will be added to the command
// The defaults are determined here
func NewGetFlags() *GetFlags {
	return &GetFlags{
		ResourceOuput: genericclioptions.NewResourceOutputFlags(),
	}
}

// AddFlags add flags tp the command
func (r *GetFlags) AddFlags(cmd *cobra.Command) {
	r.ResourceOuput.AddFlags(cmd.Flags())
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *GetFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*GetOptions, error) {
	options := &GetOptions{
		Factory:           f,
		Streams:           streams,
		OutputFormat:      *r.ResourceOuput.Output,
		Namespace:         cmd.Flags().Lookup(genericclioptions.FlagNamespace).Value.String(),
		ShowManagedFields: *r.ResourceOuput.ShowManagedFields,
	}
	return options, nil
}

type GetOptions struct {
	Factory           util.Factory
	Streams           *genericclioptions.IOStreams
	OutputFormat      string
	Namespace         string
	ShowManagedFields bool
}

func (r *GetOptions) Validate(args []string) error {
	return nil
}

func (r *GetOptions) Run(ctx context.Context, args []string) error {
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
			ShowManagedFields: r.ShowManagedFields,
			Origin:            "choreoctl",
			Branch:            branch,
			Proxy:             proxy,
		}); err != nil {
			return err
		}
		return r.parseOutput(ul)

	}
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	if err := r.Factory.GetResourceClient().Get(ctx, types.NamespacedName{
		Namespace: r.Namespace,
		Name:      args[1],
	}, u, &resourceclient.GetOptions{
		ShowManagedFields: r.ShowManagedFields,
		Origin:            "choreoctl",
		Branch:            branch,
		Proxy:             proxy,
	}); err != nil {
		return err
	}

	return r.parseOutput(u)
}

func (r *GetOptions) parseOutput(obj runtime.Unstructured) error {
	w := r.Streams.Out
	switch r.OutputFormat {
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
		return fmt.Errorf("invalid output, supported json or yaml, got: %s", r.OutputFormat)
	}

}
