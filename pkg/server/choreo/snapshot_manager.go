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

package choreo

import (
	"encoding/json"
	"sync"
	"time"

	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/util/inventory"
	"github.com/kform-dev/choreo/pkg/util/object"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

func NewSnapshotManager() *SnapshotManager {
	return &SnapshotManager{
		head:      nil,
		tail:      nil,
		snapshots: map[string]*SnapshotNode{},
	}
}

type SnapshotManager struct {
	m         sync.RWMutex
	head      *SnapshotNode
	tail      *SnapshotNode
	snapshots map[string]*SnapshotNode
}

type SnapshotNode struct {
	snapshot *Snapshot
	next     *SnapshotNode
	prev     *SnapshotNode
}

type Snapshot struct {
	ID           string
	CreatedAt    time.Time
	APIResources []*discoverypb.APIResource
	//Input
	Inventory inventory.Inventory
	//Result
}

func (r *SnapshotManager) Get(id string) (*Snapshot, bool) {
	r.m.RLock()
	defer r.m.RUnlock()

	snapshotNode, found := r.snapshots[id]
	return snapshotNode.snapshot, found
}

func (r *SnapshotManager) GetLatest() (*SnapshotNode, bool) {
	r.m.RLock()
	defer r.m.RUnlock()

	if r.tail == nil {
		return nil, false
	}
	return r.tail, true
}

func (r *SnapshotManager) GetPrevious(id string) (*SnapshotNode, bool) {
	r.m.RLock()
	defer r.m.RUnlock()

	snapshotNode, found := r.snapshots[id]
	if !found {
		return nil, false
	}
	if snapshotNode.prev == nil {
		return nil, false
	}
	return snapshotNode.prev, true
}

func (r *SnapshotManager) Create(id string, apiResources []*discoverypb.APIResource, inventory inventory.Inventory) {
	r.m.Lock()
	defer r.m.Unlock()

	node := &SnapshotNode{
		snapshot: &Snapshot{
			ID:           id,
			CreatedAt:    time.Now(),
			APIResources: apiResources,
			Inventory:    inventory,
		},
	}

	if r.head == nil {
		// this is a head node
		r.head = node
		r.tail = node
	} else {
		// update current tail next pointer to the nexw nod
		r.tail.next = node
		node.prev = r.tail
		r.tail = node
	}

	r.snapshots[node.snapshot.ID] = node
}

func (r *SnapshotManager) Delete(id string) {
	r.m.Lock()
	defer r.m.Unlock()

	node, ok := r.snapshots[id]
	if !ok {
		// id does not exists
		return
	}

	// Update pointers in the linked list
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		// Node is the head
		r.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		// Node is the tail
		r.tail = node.prev
	}

	// Remove from lookup map
	delete(r.snapshots, id)
}

func (r *SnapshotManager) List(req *choreopb.List_Request) (*choreopb.List_Response, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	ul := &unstructured.UnstructuredList{}
	ul.SetAPIVersion(choreov1alpha1.SchemeGroupVersion.Identifier())
	ul.SetKind(choreov1alpha1.SnapshotListKind)
	v, err := object.GetListPrt(ul)
	if err != nil {
		return nil, err
	}

	current := r.head
	for current != nil {
		u := &unstructured.Unstructured{}
		u.SetAPIVersion(choreov1alpha1.SchemeGroupVersion.Identifier())
		u.SetKind(choreov1alpha1.SnapshotKind)
		u.SetUID(types.UID(current.snapshot.ID))
		u.SetCreationTimestamp(metav1.NewTime(current.snapshot.CreatedAt))
		u.SetName(current.snapshot.ID)
		u.SetResourceVersion("0")

		object.AppendItem(v, u)
		current = current.next // continue
	}

	b, err := json.Marshal(ul)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot marshal err: %s", err.Error())
	}

	return &choreopb.List_Response{
		Object: b,
	}, nil
}

func (r *SnapshotManager) Diff(req *choreopb.Diff_Request) (*choreopb.Diff_Response, error) {
	latestSnapshot, found := r.GetLatest()
	if !found {
		return &choreopb.Diff_Response{}, status.Errorf(codes.NotFound, "latest snapshot not available")
	}
	latestInventory := latestSnapshot.snapshot.Inventory
	previousSnapshotNode, found := r.GetPrevious(latestSnapshot.snapshot.ID)
	previousInventory := inventory.Inventory{} // create an empty inventory
	if found {
		previousInventory = previousSnapshotNode.snapshot.Inventory
	}

	diff := choreov1alpha1.BuildDiff(metav1.ObjectMeta{
		Name:      "diff",
		Namespace: "default",
	}, nil, nil)
	if err := latestInventory.Diff(previousInventory, diff, req.Options); err != nil {
		return &choreopb.Diff_Response{}, err
	}

	b, err := json.Marshal(diff)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot marshal err: %s", err.Error())
	}

	return &choreopb.Diff_Response{
		Object: b,
	}, nil
}
