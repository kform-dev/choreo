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

package eventhandler

import (
	"fmt"

	"github.com/henderiw/store"
	"github.com/henderiw/store/memory"
)

type OnChangeFn func()

// Informers (handle the path) -> eventhandler handle the name
type Informers struct {
	Store store.Storer[*EventHandlers]
}

type EventHandlers struct {
	Store store.Storer[OnChangeFn]
}

func NewInformers() *Informers {
	return &Informers{
		Store: memory.NewStore[*EventHandlers](nil),
	}
}

func (r *Informers) AddEventHandler(path, name string, handler OnChangeFn) error {
	exists := true
	ehs, err := r.Store.Get(store.ToKey(path))
	if err != nil {
		exists = false
		ehs = &EventHandlers{
			Store: memory.NewStore[OnChangeFn](nil),
		}
	}
	if err := ehs.Store.Create(store.ToKey(name), handler); err != nil {
		return fmt.Errorf("duplicate eventhandler: %v", err)
	}
	if !exists {
		if err := r.Store.Create(store.ToKey(path), ehs); err != nil {
			return err
		}
	}
	if err := r.Store.Update(store.ToKey(path), ehs); err != nil {
		return err
	}
	return nil
}
