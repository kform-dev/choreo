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

package discovery

import (
	"context"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/henderiw/store/watch"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/choreo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func New(choreo choreo.Choreo) discoverypb.DiscoveryServer {
	return &srv{
		choreo: choreo,
	}
}

type srv struct {
	discoverypb.UnimplementedDiscoveryServer
	choreo choreo.Choreo
}

func (r *srv) getBranchContext(branch string) (*choreo.BranchCtx, error) {
	if branch == "" {
		bctx, err := r.choreo.GetBranchStore().GetCheckedOut()
		if bctx == nil {
			return nil, status.Errorf(codes.NotFound, "no checkedout branch found %v", err)
		}
		return bctx, nil
	}
	bctx, err := r.choreo.GetBranchStore().GetStore().Get(store.ToKey(branch))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "err: %s", err.Error())
	}
	return bctx, nil
}

func (r *srv) Get(ctx context.Context, req *discoverypb.Get_Request) (*discoverypb.Get_Response, error) {
	bctx, err := r.getBranchContext(req.Branch)
	if err != nil {
		return &discoverypb.Get_Response{}, err
	}

	return &discoverypb.Get_Response{Apiresources: bctx.APIStore.GetAPIResources()}, nil
}

func (r *srv) Watch(req *discoverypb.Watch_Request, stream discoverypb.Discovery_WatchServer) error {
	ctx := stream.Context()
	log := log.FromContext(ctx)

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	bctx, err := r.getBranchContext(req.Branch)
	if err != nil {
		return err
	}

	wi, err := r.choreo.GetBranchStore().WatchAPIResources(ctx, &resourceclient.ListOptions{
		Branch: bctx.Branch,
		Watch:  req.Options.Watch,
	})
	if err != nil {
		return err
	}

	go r.watch(ctx, wi, stream, cancel)

	// context got cancelled -. proxy got stopped
	<-ctx.Done()
	log.Debug("grpc watch goroutine stopped")
	return nil
}

func (r *srv) watch(ctx context.Context, wi watch.WatchInterface[*api.ResourceContext], clientStream discoverypb.Discovery_WatchServer, cancel func()) {
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
			if err := clientStream.Send(&discoverypb.Watch_Response{
				ApiResource: rctx.External,
				EventType:   GetDiscoveryPbEventType(watchEvent.Type),
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

func GetDiscoveryPbEventType(evtype watch.EventType) discoverypb.Watch_EventType {
	switch evtype {
	case watch.Added:
		return discoverypb.Watch_ADDED
	case watch.Deleted:
		return discoverypb.Watch_DELETED
	case watch.Modified:
		return discoverypb.Watch_MODIFIED
	case watch.Bookmark:
		return discoverypb.Watch_BOOKMARK
	case watch.Error:
		return discoverypb.Watch_ERROR
	default:
		return discoverypb.Watch_ERROR
	}
}
