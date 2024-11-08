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

package starlark

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func newCache(libraries []*choreov1alpha1.Library) *cache {
	modules := map[string]string{}
	for _, library := range libraries {
		if library.Spec.Type == choreov1alpha1.SoftwardTechnologyType_Starlark {
			modules[library.GetName()] = library.Spec.Code
		}
	}

	return &cache{
		cache:   make(map[string]*entry),
		modules: modules,
	}
}

type cache struct {
	m       sync.Mutex
	cache   map[string]*entry
	modules map[string]string // key = filename, value is starlark file
}

type entry struct {
	owner   unsafe.Pointer // a *cycleChecker; see cycleCheck
	globals starlark.StringDict
	err     error
	ready   chan struct{}
}

// get loads and returns an entry (if not already loaded).
func (c *cache) get(cc *cycleChecker, module string, builtins starlark.StringDict) (starlark.StringDict, error) {
	c.m.Lock()
	e := c.cache[module]
	if e != nil {
		c.m.Unlock()
		// Some other goroutine is getting this module.
		// Wait for it to become ready.
		// Detect load cycles to avoid deadlocks.
		if err := cycleCheck(e, cc); err != nil {
			return nil, err
		}
		cc.setWaitsFor(e)
		<-e.ready
		cc.setWaitsFor(nil)
	} else {
		// First request for this module.
		e = &entry{ready: make(chan struct{})}
		c.cache[module] = e
		c.m.Unlock()
		e.setOwner(cc)
		e.globals, e.err = c.doLoad(cc, module, builtins)
		e.setOwner(nil)
		// Broadcast that the entry is now ready.
		close(e.ready)
	}
	return e.globals, e.err
}

func (c *cache) doLoad(cc *cycleChecker, module string, builtins starlark.StringDict) (starlark.StringDict, error) {
	thread := &starlark.Thread{
		Name:  "exec " + module,
		Print: func(_ *starlark.Thread, msg string) { fmt.Println(msg) },
		Load: func(_ *starlark.Thread, module string) (starlark.StringDict, error) {
			// Tunnel the cycle-checker state for this "thread of loading".
			return c.get(cc, module, builtins)
		},
	}
	data, ok := c.modules[module]
	if !ok {
		return nil, fmt.Errorf("module %s not found in library", module)
	}
	return starlark.ExecFileOptions(&syntax.FileOptions{}, thread, module, data, builtins)
}

// -- concurrent cycle checking --
// A cycleChecker is used for concurrent deadlock detection.
// Each top-level call to Load creates its own cycleChecker,
// which is passed to all recursive calls it makes.
// It corresponds to a logical thread in the deadlock detection literature.
type cycleChecker struct {
	waitsFor unsafe.Pointer // an *entry; see cycleCheck
}

func (cc *cycleChecker) setWaitsFor(e *entry) {
	atomic.StorePointer(&cc.waitsFor, unsafe.Pointer(e))
}
func (e *entry) setOwner(cc *cycleChecker) {
	atomic.StorePointer(&e.owner, unsafe.Pointer(cc))
}

// cycleCheck reports whether there is a path in the waits-for graph
// from resource 'e' to thread 'me'.
//
// The waits-for graph (WFG) is a bipartite graph whose nodes are
// alternately of type entry and cycleChecker.  Each node has at most
// one outgoing edge.  An entry has an "owner" edge to a cycleChecker
// while it is being readied by that cycleChecker, and a cycleChecker
// has a "waits-for" edge to an entry while it is waiting for that entry
// to become ready.
//
// Before adding a waits-for edge, the cache checks whether the new edge
// would form a cycle.  If so, this indicates that the load graph is
// cyclic and that the following wait operation would deadlock.
func cycleCheck(e *entry, me *cycleChecker) error {
	for e != nil {
		cc := (*cycleChecker)(atomic.LoadPointer(&e.owner))
		if cc == nil {
			break
		}
		if cc == me {
			return fmt.Errorf("cycle in load graph")
		}
		e = (*entry)(atomic.LoadPointer(&cc.waitsFor))
	}
	return nil
}
