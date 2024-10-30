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

package resourceclient

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type Client interface {
	Get(ctx context.Context, key types.NamespacedName, u runtime.Unstructured, opts ...GetOption) error
	List(ctx context.Context, ul runtime.Unstructured, opts ...ListOption) error
	Apply(ctx context.Context, u runtime.Unstructured, opts ...ApplyOption) error
	Create(ctx context.Context, u runtime.Unstructured, opts ...CreateOption) error
	Update(ctx context.Context, u runtime.Unstructured, opts ...UpdateOption) error
	Delete(ctx context.Context, u runtime.Unstructured, opts ...DeleteOption) error
	Watch(ctx context.Context, u runtime.Unstructured, opts ...ListOption) chan *resourcepb.Watch_Response
	Close() error
}

type GetOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToGet(*GetOptions)
}

var _ GetOption = &GetOptions{}

type GetOptions struct {
	Proxy             types.NamespacedName
	Branch            string
	Ref               string
	Commit            *object.Commit
	ShowManagedFields bool
	Trace             string
	Origin            string
}

func (o *GetOptions) ApplyToGet(lo *GetOptions) {
	if o.Commit != nil {
		lo.Commit = o.Commit
	}
	lo.Proxy = o.Proxy
	lo.Branch = o.Branch
	lo.Ref = o.Ref
	lo.ShowManagedFields = o.ShowManagedFields
	lo.Trace = o.Trace
	lo.Origin = o.Origin

}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *GetOptions) ApplyOptions(opts []GetOption) *GetOptions {
	for _, opt := range opts {
		opt.ApplyToGet(o)
	}
	return o
}

type ListOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToList(*ListOptions)
}

var _ ListOption = &ListOptions{}

type ListOptions struct {
	Proxy             types.NamespacedName
	Branch            string
	Ref               string
	Commit            *object.Commit
	ShowManagedFields bool
	Trace             string
	Origin            string
	// A selector based on expressions
	ExprSelector *resourcepb.ExpressionSelector
	Watch        bool
}

func (o *ListOptions) ApplyToList(lo *ListOptions) {
	if o.ExprSelector != nil {
		lo.ExprSelector = o.ExprSelector
	}
	if o.Commit != nil {
		lo.Commit = o.Commit
	}
	lo.Branch = o.Branch
	lo.Proxy = o.Proxy
	lo.ShowManagedFields = o.ShowManagedFields
	lo.Trace = o.Trace
	lo.Origin = o.Origin
	lo.Watch = o.Watch
	lo.Ref = o.Ref
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *ListOptions) ApplyOptions(opts []ListOption) *ListOptions {
	for _, opt := range opts {
		opt.ApplyToList(o)
	}
	return o
}

type CreateOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToCreate(*CreateOptions)
}

var _ CreateOption = &CreateOptions{}

type CreateOptions struct {
	Trace  string
	Origin string
	DryRun []string
	Branch string
	Proxy  types.NamespacedName
}

func (o *CreateOptions) ApplyToCreate(lo *CreateOptions) {
	lo.Trace = o.Trace
	lo.Origin = o.Origin
	lo.DryRun = o.DryRun
	lo.Branch = o.Branch
	lo.Branch = o.Branch
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *CreateOptions) ApplyOptions(opts []CreateOption) *CreateOptions {
	for _, opt := range opts {
		opt.ApplyToCreate(o)
	}
	return o
}

type UpdateOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToUpdate(*UpdateOptions)
}

var _ UpdateOption = &UpdateOptions{}

type UpdateOptions struct {
	Trace  string
	Origin string
	DryRun []string
	Branch string
	Proxy  types.NamespacedName
}

func (o *UpdateOptions) ApplyToUpdate(lo *UpdateOptions) {
	lo.Trace = o.Trace
	lo.Origin = o.Origin
	lo.DryRun = o.DryRun
	lo.Branch = o.Branch
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *UpdateOptions) ApplyOptions(opts []UpdateOption) *UpdateOptions {
	for _, opt := range opts {
		opt.ApplyToUpdate(o)
	}
	return o
}

type DeleteOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToDelete(*DeleteOptions)
}

var _ DeleteOption = &DeleteOptions{}

type DeleteOptions struct {
	Trace  string
	Origin string
	DryRun []string
	Branch string
	Proxy  types.NamespacedName
}

func (o *DeleteOptions) ApplyToDelete(lo *DeleteOptions) {
	lo.Trace = o.Trace
	lo.Origin = o.Origin
	lo.DryRun = o.DryRun
	lo.Branch = o.Branch
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *DeleteOptions) ApplyOptions(opts []DeleteOption) *DeleteOptions {
	for _, opt := range opts {
		opt.ApplyToDelete(o)
	}
	return o
}

type ApplyOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToApply(*ApplyOptions)
}

var _ ApplyOption = &ApplyOptions{}

type ApplyOptions struct {
	Trace        string
	Origin       string
	DryRun       []string
	FieldManager string
	Force        bool
	Branch       string
	Proxy        types.NamespacedName
}

func (o *ApplyOptions) ApplyToApply(lo *ApplyOptions) {
	lo.Trace = o.Trace
	lo.Origin = o.Origin
	lo.DryRun = o.DryRun
	lo.FieldManager = o.FieldManager
	lo.Force = o.Force
	lo.Branch = o.Branch
	lo.Proxy = o.Proxy
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *ApplyOptions) ApplyOptions(opts []ApplyOption) *ApplyOptions {
	for _, opt := range opts {
		opt.ApplyToApply(o)
	}
	return o
}
