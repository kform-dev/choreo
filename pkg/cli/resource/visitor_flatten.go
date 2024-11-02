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

package resource

import (
	"errors"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewFlattenListVisitor creates a visitor that will expand list style runtime.Objects
// into individual items and then visit them individually.
func NewFlattenListVisitor(v Visitor, typer runtime.ObjectTyper, mapper *mapper) Visitor {
	return FlattenListVisitor{v, typer, mapper}
}

// FlattenListVisitor flattens any objects that runtime.ExtractList recognizes as a list
// - has an "Items" public field that is a slice of runtime.Objects or objects satisfying
// that interface - into multiple Infos. Returns nil in the case of no errors.
// When an error is hit on sub items (for instance, if a List contains an object that does
// not have a registered client or resource), returns an aggregate error.
type FlattenListVisitor struct {
	visitor Visitor
	typer   runtime.ObjectTyper
	mapper  *mapper
}

func (v FlattenListVisitor) Visit(fn VisitorFunc) error {
	return v.visitor.Visit(func(info *Info, err error) error {
		if err != nil {
			return err
		}
		if info.Object == nil {
			return fn(info, nil)
		}
		if !meta.IsListType(info.Object) {
			return fn(info, nil)
		}

		items := []runtime.Object{}
		itemsToProcess := []runtime.Object{info.Object}

		var errm error
		for i := 0; i < len(itemsToProcess); i++ {
			currObj := itemsToProcess[i]
			if !meta.IsListType(currObj) {
				items = append(items, currObj)
				continue
			}

			currItems, err := meta.ExtractList(currObj)
			if err != nil {
				errm = errors.Join(errm, err)
				continue
			}
			if errs := runtime.DecodeList(currItems, v.mapper.decoder); len(errs) > 0 {
				errm = errors.Join(errs...)
				continue
			}
			itemsToProcess = append(itemsToProcess, currItems...)
		}

		var errs error
		for i := range items {
			item, err := v.mapper.infoForObject(items[i], v.typer)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			// propagate list source to items source
			if len(info.Source) != 0 {
				item.Source = info.Source
			}
			if err := fn(item, nil); err != nil {
				errs = errors.Join(errs, err)
			}
		}
		return errs
	})
}
