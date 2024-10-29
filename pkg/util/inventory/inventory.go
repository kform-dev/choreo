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

package inventory

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/util/object"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Inventory map[corev1.ObjectReference]*treeNode

type treeNode struct {
	ChoreoAPI bool
	Resource  *unstructured.Unstructured
	Children  []*treeNode
}

type BuildOption interface {
	// ApplyToBuikd applies this configuration to the given options.
	ApplyToBuild(*BuildOptions)
}

var _ BuildOption = &BuildOptions{}

type BuildOptions struct {
	ShowManagedField bool
	Branch           string
	ShowChoreoAPIs   bool
}

func (o *BuildOptions) ApplyToBuild(lo *BuildOptions) {
	lo.ShowManagedField = o.ShowManagedField
	lo.Branch = o.Branch
	lo.ShowChoreoAPIs = o.ShowChoreoAPIs
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *BuildOptions) ApplyOptions(opts []BuildOption) *BuildOptions {
	for _, opt := range opts {
		opt.ApplyToBuild(o)
	}
	return o
}

func (inv Inventory) Build(ctx context.Context, client resourceclient.Client, apiResources []*discoverypb.APIResource, opts ...BuildOption) error {
	o := BuildOptions{}
	o.ApplyOptions(opts)

	for _, apiResource := range apiResources {
		// skip showing internal choreo resources: reconcilers, libraries
		if !o.ShowChoreoAPIs && apiResource.ChoreoAPI {
			continue
		}
		ul := &unstructured.UnstructuredList{}
		ul.SetAPIVersion(schema.GroupVersion{Group: apiResource.Group, Version: apiResource.Version}.String())
		ul.SetKind(apiResource.Kind)
		if err := client.List(ctx, ul, &resourceclient.ListOptions{
			ExprSelector:      &resourcepb.ExpressionSelector{},
			ShowManagedFields: o.ShowManagedField,
			Branch:            o.Branch,
		}); err != nil {
			return err
		}

		ul.EachListItem(func(o runtime.Object) error {
			u, _ := o.(*unstructured.Unstructured)

			objRef := object.GetObjectRefFromUnstructured(u)
			if _, exists := inv[objRef]; !exists {
				inv[objRef] = &treeNode{
					Children: []*treeNode{},
				}
			}
			inv[objRef].Resource = u
			inv[objRef].ChoreoAPI = apiResource.ChoreoAPI

			for _, ref := range u.GetOwnerReferences() {
				ownerObjRef := object.GetObjectRefFromOwnerRef(ref)
				if _, exists := inv[ownerObjRef]; !exists {
					inv[ownerObjRef] = &treeNode{
						Children: []*treeNode{},
					}
				}
				inv[ownerObjRef].Children = append(inv[ownerObjRef].Children, inv[objRef])
			}

			return nil
		})
	}

	// sort the children
	for _, node := range inv {
		sort.Slice(node.Children, func(i, j int) bool {
			namei := fmt.Sprintf("%s.%s %s", node.Children[i].Resource.GetKind(), node.Children[i].Resource.GetAPIVersion(), node.Children[i].Resource.GetName())
			namej := fmt.Sprintf("%s.%s %s", node.Children[j].Resource.GetKind(), node.Children[j].Resource.GetAPIVersion(), node.Children[j].Resource.GetName())
			return strings.ToLower(namei) < strings.ToLower(namej)
		})
	}
	return nil
}

func (r Inventory) Print() {
	roots := r.getRoots()
	for _, root := range roots {
		printTree(root, 0)
	}
}

func (r Inventory) getRoots() []*treeNode {
	roots := make([]*treeNode, 0)
	for _, node := range r {
		if node.Resource != nil {
			if len(node.Resource.GetOwnerReferences()) == 0 {
				roots = append(roots, node)
			}
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		namei := fmt.Sprintf("%s.%s %s", roots[i].Resource.GetKind(), roots[i].Resource.GetAPIVersion(), roots[i].Resource.GetName())
		namej := fmt.Sprintf("%s.%s %s", roots[j].Resource.GetKind(), roots[j].Resource.GetAPIVersion(), roots[j].Resource.GetName())
		return strings.ToLower(namei) < strings.ToLower(namej)
	})

	return roots
}

func (r Inventory) CollectGarbage() sets.Set[corev1.ObjectReference] {
	g := newGarbageCollector()
	for _, node := range r {
		if node.Resource == nil {
			g.collect(node)
		}
	}
	return g.sets
}

func newGarbageCollector() *garbagecollector {
	return &garbagecollector{
		sets: sets.New[corev1.ObjectReference](),
	}
}

type garbagecollector struct {
	sets sets.Set[corev1.ObjectReference]
}

func (r *garbagecollector) collect(node *treeNode) {
	for _, node := range node.Children {
		if node.Resource != nil {
			r.sets.Insert(corev1.ObjectReference{
				APIVersion: node.Resource.GetAPIVersion(),
				Kind:       node.Resource.GetKind(),
				Namespace:  node.Resource.GetNamespace(),
				Name:       node.Resource.GetName(),
			})
		}
		r.collect(node)
	}
}

func printTree(node *treeNode, indent int) {
	// Create the indent string using the indent value
	prefix := ""
	if indent >= 1 {
		prefix = strings.Repeat(" ", (indent-1)*2) + "+-"
	}

	fmt.Printf("%s%s.%s %s\n", prefix, node.Resource.GetKind(), node.Resource.GetAPIVersion(), node.Resource.GetName())

	// Recurse over children
	for _, child := range node.Children {
		printTree(child, indent+1)
	}
}

func (r Inventory) Sets() sets.Set[corev1.ObjectReference] {
	s := sets.New[corev1.ObjectReference]()
	for ref := range r {
		s.Insert(ref)
	}
	return s
}
