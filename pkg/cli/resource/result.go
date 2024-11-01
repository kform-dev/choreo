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

type Result struct {
	err     error
	visitor Visitor

	sources            []Visitor
	singleItemImplied  bool
	targetsSingleItems bool

	mapper *mapper
	//ignoreErrors []utilerrors.Matcher

	// populated by a call to Infos
	info []*Info
}

// withError allows a fluent style for internal result code.
func (r *Result) withError(err error) *Result {
	r.err = errors.Join(r.err, err)
	return r
}

// Infos returns an array of all of the resource infos retrieved via traversal.
// Will attempt to traverse the entire set of visitors only once, and will return
// a cached list on subsequent calls.
func (r *Result) Infos() ([]*Info, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.info != nil {
		return r.info, nil
	}

	infos := []*Info{}
	err := r.visitor.Visit(func(info *Info, err error) error {
		if err != nil {
			return err
		}
		infos = append(infos, info)
		return nil
	})

	r.info, r.err = infos, err
	return infos, err
}
