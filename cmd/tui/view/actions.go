package view

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

// KeyActions tracks mappings between keystrokes and actions.
type KeyActions struct {
	actions map[tcell.Key]ActionHandler
	m       sync.RWMutex
}

// ActionHandler handles a keyboard command.
type ActionHandler func()

func NewKeyActions() *KeyActions {
	return &KeyActions{
		actions: make(map[tcell.Key]ActionHandler),
	}
}

// Add adds a new key action.
func (a *KeyActions) Add(k tcell.Key, h ActionHandler) {
	a.m.Lock()
	defer a.m.Unlock()

	a.actions[k] = h
}

func (a *KeyActions) Get(key tcell.Key) (ActionHandler, bool) {
	a.m.RLock()
	defer a.m.RUnlock()

	v, ok := a.actions[key]
	return v, ok
}
