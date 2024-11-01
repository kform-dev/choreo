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

package applycmd

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

// NewCmdApply returns a cobra command.
func NewCmdApply(f util.Factory, streams *genericclioptions.IOStreams) *cobra.Command {
	flags := NewApplyFlags()

	cmd := &cobra.Command{
		Use:   "apply [flags]",
		Short: "apply resource",
		//Args:  cobra.ExactArgs(0),
		//Short:   docs.InitShort,
		//Long:    docs.InitShort + "\n" + docs.InitLong,
		//Example: docs.InitExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := flags.ToOptions(cmd, f, streams)
			if err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			return o.Run(cmd.Context(), args)
		},
	}
	flags.AddFlags(cmd)
	return cmd
}

type ApplyFlags struct {
	FileNameFlags *genericclioptions.FileNameFlags
	Streams       *genericclioptions.IOStreams
}

// NewApplyFlags determines which flags will be added to the command
// The defaults are determined here
func NewApplyFlags() *ApplyFlags {
	usage := "The files that contain the configuration to apply."

	// setup command defaults
	filenames := []string{}
	recursive := false

	return &ApplyFlags{
		FileNameFlags: &genericclioptions.FileNameFlags{Usage: usage, Filenames: &filenames, Recursive: &recursive},
	}
}

// AddFlags add flags tp the command
func (r *ApplyFlags) AddFlags(cmd *cobra.Command) {
	r.FileNameFlags.AddFlags(cmd.Flags())
}

// ToOptions renders the options based on the flags that were set and will be the base context used to run the command
func (r *ApplyFlags) ToOptions(cmd *cobra.Command, f util.Factory, streams *genericclioptions.IOStreams) (*ApplyOptions, error) {
	options := &ApplyOptions{
		Factory: f,
		Streams: streams,
	}
	options.FileNameOptions = r.FileNameFlags.ToOptions()
	return options, nil
}

type ApplyOptions struct {
	Factory         util.Factory
	Streams         *genericclioptions.IOStreams
	FileNameOptions resource.FilenameOptions
}

func (r *ApplyOptions) Validate(args []string) error {
	return nil
}

func (r *ApplyOptions) Run(ctx context.Context, args []string) error {
	infos, err := r.GetObjects()
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
		fmt.Println("applying", u.GetAPIVersion(), u.GetKind(), u.GetName())
		if err := r.applyOneObject(ctx, u); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (r *ApplyOptions) GetObjects() ([]*resource.Info, error) {
	b := resource.NewBuilder(r.Factory.GetResourceMapper(), r.Factory.GetProxy(), r.Factory.GetBranch()).
		Unstructured().
		ContinueOnError().
		FilenameParam(&r.FileNameOptions).
		Flatten().
		Do()

	return b.Infos()
}

func (r *ApplyOptions) applyOneObject(ctx context.Context, ru runtime.Unstructured) error {
	client := r.Factory.GetResourceClient()
	return client.Apply(ctx, ru, &resourceclient.ApplyOptions{
		Origin:       "choreoctl",
		FieldManager: "inputfileloader",
		Branch:       r.Factory.GetBranch(),
		Proxy:        r.Factory.GetProxy(),
	})
}
