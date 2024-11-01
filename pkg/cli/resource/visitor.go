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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

type Info struct {
	// Namespace will be set if the object is namespaced and has a specified value.
	Namespace string
	Name      string

	// Optional, Source is the filename or URL to template file (.json or .yaml),
	// or stdin to use to handle the resource
	Source string
	// Optional, this is the most recent value of the resource.
	Object runtime.Object
}

// String returns the general purpose string representation
func (i *Info) String() string {
	basicInfo := fmt.Sprintf("Name: %q, Namespace: %q", i.Name, i.Namespace)
	return basicInfo
}

// Visit implements Visitor
func (i *Info) Visit(fn VisitorFunc) error {
	return fn(i, nil)
}

// VisitorList implements Visit for the sub visitors it contains. The first error
// returned from a child Visitor will terminate iteration.
type VisitorList []Visitor

// Visit implements Visitor
func (l VisitorList) Visit(fn VisitorFunc) error {
	for i := range l {
		if err := l[i].Visit(fn); err != nil {
			return err
		}
	}
	return nil
}
