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
)

// IdentifierSet manages a set of identifiers.
type IdentifierSet struct {
	m           sync.RWMutex
	identifiers map[string]struct{} // Map to check for existence quickly
	sortedIDs   []string            // Sorted slice to maintain order
}

// NewIdentifierSet creates a new IdentifierSet.
func NewIdentifierSet() *IdentifierSet {
	return &IdentifierSet{
		identifiers: make(map[string]struct{}),
		sortedIDs:   []string{},
	}
}

// AddIdentifier adds a new identifier to the set and returns the position it was added at,
// or the existing position if it was already present.
func (s *IdentifierSet) AddIdentifier(id string) (int, bool) {
	s.m.Lock()
	defer s.m.Unlock()
	if _, exists := s.identifiers[id]; exists {
		// Return the position of the existing identifier
		pos := sort.SearchStrings(s.sortedIDs, id)
		return pos, true
	}

	// Add the new identifier
	s.identifiers[id] = struct{}{}
	s.sortedIDs = append(s.sortedIDs, id)
	sort.Strings(s.sortedIDs) // Re-sort the slice to maintain order

	// Find the position where the new identifier was added
	pos := sort.SearchStrings(s.sortedIDs, id)
	return pos, false
}

func (s *IdentifierSet) DeleteIdentifier(id string) int {
	s.m.Lock()
	defer s.m.Unlock()
	if _, exists := s.identifiers[id]; !exists {
		return -1 // Identifier not found
	}

	// Remove the identifier from the map
	delete(s.identifiers, id)

	// Find the position and remove it from the sorted slice
	pos := sort.SearchStrings(s.sortedIDs, id)
	s.sortedIDs = append(s.sortedIDs[:pos], s.sortedIDs[pos+1:]...)

	return pos
}
