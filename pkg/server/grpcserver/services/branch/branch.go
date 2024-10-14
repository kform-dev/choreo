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

package branch

import (
	"context"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/henderiw/store/watch"
	"github.com/kform-dev/choreo/pkg/proto/branchpb"
	"github.com/kform-dev/choreo/pkg/repository"
	"github.com/kform-dev/choreo/pkg/server/choreo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func New(choreo choreo.Choreo) branchpb.BranchServer {
	return &srv{
		choreo: choreo,
	}
}

type srv struct {
	branchpb.UnimplementedBranchServer
	choreo choreo.Choreo
}

func (r *srv) Get(ctx context.Context, req *branchpb.Get_Request) (*branchpb.Get_Response, error) {
	repo := r.choreo.GetRootChoreoInstance().GetRepo()
	logs, err := repo.GetBranchLog(req.Branch)
	if err != nil {
		return &branchpb.Get_Response{}, err
	}
	return &branchpb.Get_Response{
		Logs: logs,
	}, nil
}
func (r *srv) List(ctx context.Context, req *branchpb.List_Request) (*branchpb.List_Response, error) {
	repo := r.choreo.GetRootChoreoInstance().GetRepo()
	branches := repo.GetBranches()
	return &branchpb.List_Response{BranchObjects: branches}, nil
}
func (r *srv) Create(ctx context.Context, req *branchpb.Create_Request) (*branchpb.Create_Response, error) {
	repo := r.choreo.GetRootChoreoInstance().GetRepo()
	if err := repo.CreateBranch(req.Branch); err != nil {
		return &branchpb.Create_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	return &branchpb.Create_Response{}, nil
}
func (r *srv) Delete(ctx context.Context, req *branchpb.Delete_Request) (*branchpb.Delete_Response, error) {
	repo := r.choreo.GetRootChoreoInstance().GetRepo()
	if err := repo.DeleteBranch(req.Branch); err != nil {
		return &branchpb.Delete_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	return &branchpb.Delete_Response{}, nil
}

func (r *srv) Diff(ctx context.Context, req *branchpb.Diff_Request) (*branchpb.Diff_Response, error) {
	repo := r.choreo.GetRootChoreoInstance().GetRepo()
	diffs, err := repo.DiffBranch(req.SrcBranch, req.DstBranch)
	if err != nil {
		return &branchpb.Diff_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	return &branchpb.Diff_Response{
		Diffs: diffs,
	}, nil
}

func (r *srv) Merge(ctx context.Context, req *branchpb.Merge_Request) (*branchpb.Merge_Response, error) {
	repo := r.choreo.GetRootChoreoInstance().GetRepo()
	if err := repo.MergeBranch(req.SrcBranch, req.DstBranch); err != nil {
		return &branchpb.Merge_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	return &branchpb.Merge_Response{}, nil
}

func (r *srv) Stash(ctx context.Context, req *branchpb.Stash_Request) (*branchpb.Stash_Response, error) {
	repo := r.choreo.GetRootChoreoInstance().GetRepo()
	if err := repo.StashBranch(req.Branch); err != nil {
		return &branchpb.Stash_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	return &branchpb.Stash_Response{}, nil
}

func (r *srv) Checkout(ctx context.Context, req *branchpb.Checkout_Request) (*branchpb.Checkout_Response, error) {
	repo := r.choreo.GetRootChoreoInstance().GetRepo()
	if err := repo.Checkout(req.Branch); err != nil {
		return &branchpb.Checkout_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	return &branchpb.Checkout_Response{}, nil
}

func (r *srv) StreamFiles(req *branchpb.Get_Request, stream branchpb.Branch_StreamFilesServer) error {
	repo := r.choreo.GetRootChoreoInstance().GetRepo()
	return repo.StreamFiles(req.Branch, &repository.FileWriter{Stream: stream})
}

func (r *srv) Watch(req *branchpb.Watch_Request, stream branchpb.Branch_WatchServer) error {
	ctx := stream.Context()
	log := log.FromContext(ctx)

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	wi, err := r.choreo.GetBranchStore().WatchBranches(ctx, &store.ListOptions{})
	if err != nil {
		return err
	}

	go r.watch(ctx, wi, stream, cancel)

	// context got cancelled -. proxy got stopped
	<-ctx.Done()
	log.Debug("grpc watch goroutine stopped")
	return nil
}

func (r *srv) watch(ctx context.Context, wi watch.WatchInterface[*choreo.BranchCtx], clientStream branchpb.Branch_WatchServer, cancel func()) {
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

			if err := clientStream.Send(&branchpb.Watch_Response{
				BranchObj: GetbranchObj(watchEvent.Object),
				EventType: GetBranchPbEventType(watchEvent.Type),
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

func GetbranchObj(branchCtx *choreo.BranchCtx) *branchpb.BranchObject {
	return &branchpb.BranchObject{
		Name:       branchCtx.Branch,
		CheckedOut: branchCtx.State.String() == "CheckedOut",
	}
}

func GetBranchPbEventType(evtype watch.EventType) branchpb.Watch_EventType {
	switch evtype {
	case watch.Added:
		return branchpb.Watch_ADDED
	case watch.Deleted:
		return branchpb.Watch_DELETED
	case watch.Modified:
		return branchpb.Watch_MODIFIED
	case watch.Bookmark:
		return branchpb.Watch_BOOKMARK
	case watch.Error:
		return branchpb.Watch_ERROR
	default:
		return branchpb.Watch_ERROR
	}
}
