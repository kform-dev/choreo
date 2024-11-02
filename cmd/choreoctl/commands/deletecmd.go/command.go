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
	"errors"
	"fmt"

	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/cli/resource"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/kform-dev/choreo/pkg/util/object"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	//docs "github.com/kform-dev/kform/internal/docs/generated/applydocs"
)

// NewCmdDelete returns a cobra command.
func NewCmdDelete(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewDeleteFlags()

	cmd := &cobra.Command{
		Use:   "delete <RESOURCE> <NAME> <NAMESPACE> [flags]",
		Short: "delete resource",
		Args:  cobra.MaximumNArgs(2), // TODO BULK delete, right now this is a protection
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

type DeleteFlags struct {
	FileNameFlags *genericclioptions.FileNameFlags
	Streams       *genericclioptions.IOStreams
}

// NewDeleteFlags determines which flags will be added to the command
// The defaults are determined here
func NewDeleteFlags() *DeleteFlags {
	usage := "The files that contain the resource to delete."

	return &DeleteFlags{
		FileNameFlags: genericclioptions.NewFileNameFlags(usage),
	}
}

// AddFlags add flags tp the command
func (r *DeleteFlags) AddFlags(cmd *cobra.Command) {
	r.FileNameFlags.AddFlags(cmd.Flags())
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *DeleteFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*DeleteOptions, error) {
	options := &DeleteOptions{
		Factory: f,
		Streams: streams,
	}
	options.FileNameFlags = r.FileNameFlags
	return options, nil
}

type DeleteOptions struct {
	Factory       util.Factory
	Streams       *genericclioptions.IOStreams
	FileNameFlags *genericclioptions.FileNameFlags
}

func (r *DeleteOptions) Validate(args []string) error {
	if len(args) == 1 && len(*r.FileNameFlags.Filenames) == 0 {
		return fmt.Errorf("nothing to delete, missing a resourcename or filename input (-f)")
	}
	return nil
}

func (r *DeleteOptions) Run(ctx context.Context, args []string) error {
	infos, err := r.GetObjects(args)
	if err != nil {
		return err
	}
	if len(infos) == 0 {
		return fmt.Errorf("no object passed to apply")
	}
	var errs error
	for _, info := range infos {
		ru, err := object.GetUnstructructered(info.Object)
		if err != nil {
			return err
		}
		u := &unstructured.Unstructured{
			Object: ru.UnstructuredContent(),
		}
		fmt.Println("deleting", u.GetAPIVersion(), u.GetKind(), u.GetName())
		if err := r.deleteOneObject(ctx, u); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (r *DeleteOptions) GetObjects(args []string) ([]*resource.Info, error) {
	b := resource.NewBuilder(r.Factory.GetResourceMapper(), r.Factory.GetProxy(), r.Factory.GetBranch()).
		Unstructured().
		ContinueOnError().
		FilenameParam(&resource.FilenameOptions{Filenames: *r.FileNameFlags.Filenames, Recursive: *r.FileNameFlags.Recursive}).
		ResourceTypeOrNameArgs(args...).
		Flatten().
		Do()

	return b.Infos()
}

func (r *DeleteOptions) deleteOneObject(ctx context.Context, ru runtime.Unstructured) error {
	client := r.Factory.GetResourceClient()
	return client.Delete(ctx, ru, &resourceclient.DeleteOptions{
		Origin: "choreoctl",
		Branch: r.Factory.GetBranch(),
		Proxy:  r.Factory.GetProxy(),
	})
}
