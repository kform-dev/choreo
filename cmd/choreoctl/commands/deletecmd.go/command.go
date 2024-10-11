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

package deletecmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
		Use:   "delete <RESOURCE> <NAME> <NAMESPACE>",
		Short: "delete resource",
		Args:  cobra.MinimumNArgs(2),
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
	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return err
	}

	name := args[1]

	parts := strings.SplitN(args[0], ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expecting <resource>.<group>, got: %v", parts)
	}
	gvk, err := r.factory.GetResourceMapper().KindFor(ctx, schema.GroupResource{Group: parts[1], Resource: parts[0]}, branch)
	if err != nil {
		return err
	}

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	u.SetName(name)
	u.SetNamespace("default")

	client := r.factory.GetResourceClient()
	return client.Delete(ctx, u, &resourceclient.DeleteOptions{
		Origin: "choreoctl",
		Branch: branch,
	})
}
