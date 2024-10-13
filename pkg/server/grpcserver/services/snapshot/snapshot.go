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

package snapshot

import (
	"context"

	"github.com/kform-dev/choreo/pkg/proto/snapshotpb"
	"github.com/kform-dev/choreo/pkg/server/choreo"
)

func New(choreo choreo.Choreo) snapshotpb.SnapshotServer {
	return &srv{
		choreo: choreo,
	}
}

type srv struct {
	snapshotpb.UnimplementedSnapshotServer
	choreo choreo.Choreo
}

func (r *srv) Get(ctx context.Context, req *snapshotpb.Get_Request) (*snapshotpb.Get_Response, error) {
	return r.choreo.SnapshotManager().Get(req)
}

func (r *srv) List(ctx context.Context, req *snapshotpb.List_Request) (*snapshotpb.List_Response, error) {
	return r.choreo.SnapshotManager().List(req)
}

func (r *srv) Delete(ctx context.Context, req *snapshotpb.Delete_Request) (*snapshotpb.Delete_Response, error) {
	return r.choreo.SnapshotManager().Delete(req)
}

func (r *srv) Diff(ctx context.Context, req *snapshotpb.Diff_Request) (*snapshotpb.Diff_Response, error) {
	return r.choreo.SnapshotManager().Diff(req)
}

/*
func (r *srv) Watch(req *snapshotpb.Watch_Request, stream snapshotpb.Snapshot_WatchServer) error {
	ctx := stream.Context()
	log := log.FromContext(ctx)

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	wi, err := r.choreo.SnapshotManager().Watch(ctx)
	if err != nil {
		return err
	}

	go r.watch(ctx, wi, stream, cancel)

	// context got cancelled -. proxy got stopped
	<-ctx.Done()
	log.Debug("grpc watch goroutine stopped")
	return nil
}

func (r *srv) watch(ctx context.Context, wi watch.WatchInterface[*api.ResourceContext], clientStream snapshotpb.Snapshot_WatchServer, cancel func()) {
	log := log.FromContext(ctx)

	resultCh := wi.ResultChan()
	for {
		select {
		case <-ctx.Done():
			log.Debug("grpc watch stopped, stopping storage watch")
			wi.Stop()
			return
		case watchEvent, ok := <-resultCh:
			if !ok {
				log.Debug("result channel closed, stopping storage watch")
				cancel()
				continue
			}

			if watchEvent.Type == watch.Error {
				log.Error("received watch error", "event", watchEvent)
				cancel()
				continue
			}

			if watchEvent.Object == nil {
				log.Error("received nil object in watch event", "event", watchEvent)
				cancel()
				continue
			}

			rctx := watchEvent.Object
			if err := clientStream.Send(&snapshotpb.Watch_Response{
				Object:    rctx.APIResource,
				EventType: GetSnapshotPbEventType(watchEvent.Type),
			}); err != nil {
				p, _ := peer.FromContext(clientStream.Context())
				addr := "unknown"
				if p != nil {
					addr = p.Addr.String()
				}
				log.Error("grpc watch send stream failed", "client", addr)
			}
		}
	}
}

func GetSnapshotPbEventType(evtype watch.EventType) snapshotpb.Watch_EventType {
	switch evtype {
	case watch.Added:
		return snapshotpb.Watch_ADDED
	case watch.Deleted:
		return snapshotpb.Watch_DELETED
	case watch.Modified:
		return snapshotpb.Watch_MODIFIED
	case watch.Bookmark:
		return snapshotpb.Watch_BOOKMARK
	case watch.Error:
		return snapshotpb.Watch_ERROR
	default:
		return snapshotpb.Watch_ERROR
	}
}
*/
