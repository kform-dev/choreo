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

package view

import (
	"sort"
	"sync"

	"github.com/gdamore/tcell/v2"
)

/*
type KeyActionEntry struct {
	key       tcell.Key
	keyaction KeyAction
}
*/

// KeyActions tracks mappings between keystrokes and actions.
type KeyActions struct {
	actions map[string]KeyAction
	m       sync.RWMutex
}

// KeyAction represents a keyboard action.
type KeyAction struct {
	Key         tcell.Key
	Description string
	ShortName   string
	Action      ActionHandler
}

// ActionHandler handles a keyboard command.
type ActionHandler func()

func NewKeyActions() *KeyActions {
	return &KeyActions{
		actions: make(map[string]KeyAction),
	}
}

// Add adds a new key action.
func (a *KeyActions) Add(ka KeyAction) {
	a.m.Lock()
	defer a.m.Unlock()

	a.actions[tcell.KeyNames[ka.Key]] = ka
}

func (a *KeyActions) Get(k string) (KeyAction, bool) {
	a.m.RLock()
	defer a.m.RUnlock()

	v, ok := a.actions[k]
	return v, ok
}

func (a *KeyActions) List(visitorFunc func(string, KeyAction)) {
	a.m.RLock()
	defer a.m.RUnlock()

	for key, action := range a.actions {
		if visitorFunc != nil {
			visitorFunc(key, action)
		}
	}
}

func (r *KeyActions) MapKeys() map[string]int {
	keys := []string{}
	r.List(func(key string, _ KeyAction) {
		keys = append(keys, key)
	})
	sort.Strings(keys)
	mapKeys := map[string]int{}
	for i, key := range keys {
		mapKeys[key] = i
	}
	return mapKeys
}
