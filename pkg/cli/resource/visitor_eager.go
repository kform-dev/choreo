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

import "errors"

// EagerVisitorList implements Visit for the sub visitors it contains. All errors
// will be captured and returned at the end of iteration.
type EagerVisitorList []Visitor

// Visit implements Visitor, and gathers errors that occur during processing until
// all sub visitors have been visited.
func (r EagerVisitorList) Visit(fn VisitorFunc) error {
	var errm error
	for i := range r {
		err := r[i].Visit(func(info *Info, err error) error {
			if err != nil {
				errm = errors.Join(errm, err)
				return nil
			}
			if err := fn(info, nil); err != nil {
				errm = errors.Join(errm, err)
			}
			return nil
		})
		if err != nil {
			errm = errors.Join(errm, err)
		}
	}
	return errm
}
