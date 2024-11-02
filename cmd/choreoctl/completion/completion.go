package completion

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kform-dev/choreo/cmd/choreoctl/commands/apiresourcescmd"
	"github.com/kform-dev/choreo/cmd/choreoctl/commands/getcmd"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/branchclient"
	"github.com/kform-dev/choreo/pkg/client/go/util"
	"github.com/spf13/cobra"
)

type Completion struct {
	Factory util.Factory
}

func (r *Completion) ResourceTypeAndNameCompletionFunc() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var completions []string
		directive := cobra.ShellCompDirectiveNoFileComp
		//resources := []string{"pods", "services", "deployments"}

		if len(args) > 0 {
			if len(args) == 1 {
				completions = r.compGetResource(cmd, args[0], toComplete)
			}
		} else {
			typeComps := r.compGetResourceList(cmd, toComplete)
			completions = append(completions, typeComps...)

		}
		return completions, directive
	}
}

func (r *Completion) compGetResourceList(cmd *cobra.Command, toComplete string) []string {
	ctx := cmd.Context()
	buf := new(bytes.Buffer)
	streams := &genericclioptions.IOStreams{In: os.Stdin, Out: buf, ErrOut: io.Discard}

	o := apiresourcescmd.APIResourcesOptions{
		Factory: r.Factory,
		Streams: streams,
		Output:  "name",
	}
	if err := o.Complete(); err != nil {
		return []string{}
	}
	o.Output = "name"
	// Ignore errors as the output may still be valid
	o.Run(ctx)

	prefix := ""
	suffix := toComplete
	var comps []string
	resources := strings.Split(buf.String(), "\n")
	for _, res := range resources {
		if res != "" && strings.HasPrefix(res, suffix) {
			comps = append(comps, fmt.Sprintf("%s%s", prefix, res))
		}
	}
	return comps
}

func (r *Completion) compGetResource(cmd *cobra.Command, resourceName string, toComplete string) []string {
	ctx := cmd.Context()
	buf := new(bytes.Buffer)

	o := &getcmd.GetOptions{
		Factory:      r.Factory,
		Streams:      &genericclioptions.IOStreams{In: os.Stdin, Out: buf, ErrOut: io.Discard},
		OutputFormat: "completion",
	}
	// Ignore errors as the output may still be valid
	o.Run(ctx, []string{resourceName})

	prefix := ""
	suffix := toComplete
	var comps []string
	resources := strings.Split(buf.String(), "\n")
	for _, res := range resources {
		if res != "" && strings.HasPrefix(res, suffix) {
			comps = append(comps, fmt.Sprintf("%s%s", prefix, res))
		}
	}
	return comps
}

func (r *Completion) CompBranch(cmd *cobra.Command, toComplete string) []string {
	ctx := cmd.Context()
	branchClient := r.Factory.GetBranchClient()
	branches, err := branchClient.List(ctx, &branchclient.ListOptions{
		Proxy: r.Factory.GetProxy(),
	})
	if err != nil {
		return []string{}
	}
	prefix := ""
	suffix := toComplete
	var comps []string
	for _, branchObj := range branches {
		if branchObj.Name != "" && strings.HasPrefix(branchObj.Name, suffix) {
			comps = append(comps, fmt.Sprintf("%s%s", prefix, branchObj.Name))
		}
	}
	return comps

}
