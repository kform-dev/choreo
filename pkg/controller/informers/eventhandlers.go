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

package informers

import "sync"

type eventhandlers struct {
	m        sync.RWMutex
	handlers map[string]OnChangeFn
}

func newEventHandlers() *eventhandlers {
	return &eventhandlers{
		handlers: make(map[string]OnChangeFn, 0),
	}
}

func (r *eventhandlers) add(key string, h OnChangeFn) {
	r.m.Lock()
	defer r.m.Unlock()
	r.handlers[key] = h
}

/*
	func (r *eventhandlers) len() int {
		r.m.RLock()
		defer r.m.RUnlock()
		return len(r.handlers)
	}
*/
func (r *eventhandlers) list() []OnChangeFn {
	r.m.RLock()
	defer r.m.RUnlock()
	handlers := make([]OnChangeFn, 0, len(r.handlers))
	for _, handler := range r.handlers {
		handlers = append(handlers, handler)
	}
	return handlers
}
