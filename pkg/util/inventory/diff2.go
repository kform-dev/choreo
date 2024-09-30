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

import "fmt"

func (inv Inventory) Diff(invBefore Inventory) error {
	invSetsBefore := invBefore.Sets()
	// this the
	for ref, treenode := range inv {
		if invSetsBefore.Has(ref) {
			//diffStr, err := diff(invBefore[ref].Resource, treenode.Resource)
			diffStr, err := diff2(invBefore[ref].Resource, treenode.Resource)
			if err != nil {
				return err
			}
			if diffStr == "" {
				inv[ref].DiffStatus = "="
			} else {
				inv[ref].DiffStatus = "~"
				inv[ref].DiffString = diffStr
				fmt.Println(diffStr)
			}
			invSetsBefore.Delete(ref)
		} else {
			inv[ref].DiffStatus = "="
		}
	}
	for _, ref := range invSetsBefore.UnsortedList() {
		inv[ref].DiffStatus = "-"
	}
	return nil
}
