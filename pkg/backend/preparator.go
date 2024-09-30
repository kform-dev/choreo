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

package backend

import (
	"context"

	"github.com/kform-dev/choreo/pkg/server/apiserver/registry"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewClaimPreparator(be Backend) registry.APIPrepator {
	return &claimPreparator{
		be: be,
	}
}

type claimPreparator struct {
	be Backend
}

func (r *claimPreparator) PrepareForCreate(ctx context.Context, obj runtime.Unstructured) error {
	return r.be.Claim(ctx, obj)
}

func (r *claimPreparator) PrepareForUpdate(ctx context.Context, obj, old runtime.Unstructured) error {
	return r.be.Claim(ctx, obj)
}

func (r *claimPreparator) PrepareForDelete(ctx context.Context, obj runtime.Unstructured) error {
	return r.be.Release(ctx, obj)
}

func NewIndexPreparator(be Backend) registry.APIPrepator {
	return &indexPreparator{
		be: be,
	}
}

type indexPreparator struct {
	be Backend
}

func (r *indexPreparator) PrepareForCreate(ctx context.Context, obj runtime.Unstructured) error {
	return r.be.CreateIndex(ctx, obj)
}

func (r *indexPreparator) PrepareForUpdate(ctx context.Context, obj, old runtime.Unstructured) error {
	return r.be.CreateIndex(ctx, obj)
}

func (r *indexPreparator) PrepareForDelete(ctx context.Context, obj runtime.Unstructured) error {
	return r.be.DeleteIndex(ctx, obj)
}
