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

package rest

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kform-dev/choreo/pkg/server/apiserver/watch"
	"github.com/kform-dev/choreo/pkg/server/selector"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ListOptions is a subset of "k8s.io/apimachinery/pkg/apis/meta/internalversio

type Storage interface {
	Get(ctx context.Context, key string, opts ...GetOption) (runtime.Unstructured, error)
	List(ctx context.Context, opts ...ListOption) (runtime.Unstructured, error)
	Apply(ctx context.Context, obj runtime.Unstructured, opts ...ApplyOption) (runtime.Unstructured, error)
	Create(ctx context.Context, obj runtime.Unstructured, opts ...CreateOption) (runtime.Unstructured, error)
	Update(ctx context.Context, obj runtime.Unstructured, opts ...UpdateOption) (runtime.Unstructured, error)
	Delete(ctx context.Context, key string, opts ...DeleteOption) (runtime.Unstructured, error)
	Watch(ctx context.Context, opts ...ListOption) (watch.Interface, error)
}

type Strategy interface {
	CreateStrategy
	UpdateStrategy
	DeleteStrategy
}

type CreateStrategy interface {
	ValidateCreate(ctx context.Context, obj runtime.Unstructured) field.ErrorList
	PrepareForCreate(ctx context.Context, obj runtime.Unstructured) error
}

type UpdateStrategy interface {
	ValidateUpdate(ctx context.Context, obj, old runtime.Unstructured) field.ErrorList
	PrepareForUpdate(ctx context.Context, obj, old runtime.Unstructured) error
}

type DeleteStrategy interface {
	PrepareForDelete(ctx context.Context, obj runtime.Unstructured) error
}

type GetOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToGet(*GetOptions)
}

var _ GetOption = &GetOptions{}

type GetOptions struct {
	Commit            *object.Commit
	ShowManagedFields bool
	Trace             string
	Origin            string
}

func (o *GetOptions) ApplyToGet(lo *GetOptions) {
	if o.Commit != nil {
		lo.Commit = o.Commit
	}
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
	Commit            *object.Commit
	ShowManagedFields bool
	Trace             string
	Origin            string
	// A selector based on expressions
	Selector selector.Selector
	Watch    bool
}

func (o *ListOptions) ApplyToList(lo *ListOptions) {
	if o.Commit != nil {
		lo.Commit = o.Commit
	}
	lo.ShowManagedFields = o.ShowManagedFields
	lo.Trace = o.Trace
	lo.Origin = o.Origin
	lo.Selector = o.Selector
	lo.Watch = o.Watch
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
	DryRun bool
}

func (o *CreateOptions) ApplyToCreate(lo *CreateOptions) {
	lo.Trace = o.Trace
	lo.Origin = o.Origin
	lo.DryRun = o.DryRun
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
	DryRun bool
}

func (o *UpdateOptions) ApplyToUpdate(lo *UpdateOptions) {
	lo.Trace = o.Trace
	lo.Origin = o.Origin
	lo.DryRun = o.DryRun
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
	DryRun bool
}

func (o *DeleteOptions) ApplyToDelete(lo *DeleteOptions) {
	lo.Trace = o.Trace
	lo.Origin = o.Origin
	lo.DryRun = o.DryRun
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
	DryRun       bool
	FieldManager string
	Force        bool
}

func (o *ApplyOptions) ApplyToApply(lo *ApplyOptions) {
	lo.Trace = o.Trace
	lo.Origin = o.Origin
	lo.DryRun = o.DryRun
	lo.FieldManager = o.FieldManager
	lo.Force = o.Force
}

// ApplyOptions applies the given get options on these options,
// and then returns itself (for convenient chaining).
func (o *ApplyOptions) ApplyOptions(opts []ApplyOption) *ApplyOptions {
	for _, opt := range opts {
		opt.ApplyToApply(o)
	}
	return o
}
