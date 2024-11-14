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

package resource

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/proto/grpcerrors"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kform-dev/choreo/pkg/server/apiserver/watch"
	"github.com/kform-dev/choreo/pkg/server/choreo"
	"github.com/kform-dev/choreo/pkg/server/selector"
	uobject "github.com/kform-dev/choreo/pkg/util/object"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func New(choreo choreo.Choreo) resourcepb.ResourceServer {
	return &srv{
		choreo: choreo,
	}
}

type srv struct {
	resourcepb.UnimplementedResourceServer
	choreo choreo.Choreo
}

func (r *srv) Get(ctx context.Context, req *resourcepb.Get_Request) (*resourcepb.Get_Response, error) {
	log := log.FromContext(ctx)

	bctx, err := r.getBranchContext(req.Options.Branch)
	if err != nil {
		return &resourcepb.Get_Response{}, err
	}

	u, err := uobject.GetUnstructured(req.Object)
	if err != nil {
		return &resourcepb.Get_Response{}, err
	}

	log.Debug("get", "apiVersion", u.GetAPIVersion(), "kind", u.GetKind(), "name", u.GetName())

	rctx, err := r.getAPIContext(bctx, u)
	if err != nil {
		return &resourcepb.Get_Response{}, err
	}
	convertToInternal(rctx, u)

	var commit *object.Commit
	if bctx.State.String() != "CheckedOut" {
		commit, err = r.choreo.GetStatus().Get().RootChoreoInstance.GetRepo().GetBranchCommit(bctx.Branch)
		if err != nil {
			return &resourcepb.Get_Response{}, err
		}
	}

	// invoke storage
	obj, err := rctx.Storage.Get(ctx, u.GetName(), &rest.GetOptions{
		ShowManagedFields: req.Options.ShowManagedField,
		Trace:             req.Options.Trace,
		Origin:            req.Options.Origin,
		Commit:            commit,
	})
	if err != nil {
		fmt.Println("get error", err)
		return &resourcepb.Get_Response{}, err
	}

	u = &unstructured.Unstructured{Object: obj.UnstructuredContent()}
	convertFromInternal(rctx, u)

	b, err := json.Marshal(u.Object)
	if err != nil {
		return &resourcepb.Get_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
	}

	return &resourcepb.Get_Response{Object: b}, nil
}

func (r *srv) List(ctx context.Context, req *resourcepb.List_Request) (*resourcepb.List_Response, error) {
	log := log.FromContext(ctx)

	bctx, err := r.getBranchContext(req.Options.Branch)
	if err != nil {
		fmt.Println("list bctx", req)
		return &resourcepb.List_Response{}, err
	}

	u, err := uobject.GetUnstructured(req.Object)
	if err != nil {
		return &resourcepb.List_Response{}, err
	}

	selector, err := selector.ResourceExprSelectorAsSelector(req.Options.ExprSelector)
	if err != nil {
		return &resourcepb.List_Response{}, status.Errorf(codes.InvalidArgument, "invalid selector, err: %s", err.Error())
	}

	log.Debug("list", "apiVersion", u.GetAPIVersion(), "kind", u.GetKind(), "name", u.GetName(), "options", req.Options)

	rctx, err := r.getAPIContext(bctx, u)
	if err != nil {
		return &resourcepb.List_Response{}, err
	}
	convertToInternal(rctx, u)

	var commit *object.Commit
	if bctx.State.String() != "CheckedOut" {
		commit, err = r.choreo.GetStatus().Get().RootChoreoInstance.GetRepo().GetBranchCommit(bctx.Branch)
		if err != nil {
			return &resourcepb.List_Response{}, err
		}
	}

	// invoke storage
	obj, err := rctx.Storage.List(ctx, &rest.ListOptions{
		Selector:          selector,
		ShowManagedFields: req.Options.ShowManagedField,
		Trace:             req.Options.Trace,
		Origin:            req.Options.Origin,
		Commit:            commit,
	})
	if err != nil {
		return &resourcepb.List_Response{}, err
	}

	if !obj.IsList() {
		return &resourcepb.List_Response{}, status.Errorf(codes.Internal, "expecting a list")
	}
	ul := &unstructured.UnstructuredList{}
	ul.SetUnstructuredContent(obj.UnstructuredContent())
	convertListFromInternal(rctx, ul)

	b, err := json.Marshal(ul)
	if err != nil {
		return &resourcepb.List_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
	}

	return &resourcepb.List_Response{Object: b}, nil
}

func (r *srv) Apply(ctx context.Context, req *resourcepb.Apply_Request) (*resourcepb.Apply_Response, error) {
	log := log.FromContext(ctx)

	bctx, err := r.getBranchContext(req.Options.Branch)
	if err != nil {
		return &resourcepb.Apply_Response{}, err
	}

	u, err := uobject.GetUnstructured(req.Object)
	if err != nil {
		return &resourcepb.Apply_Response{}, err
	}

	orig := u.DeepCopy()

	log.Debug("apply", "apiVersion", u.GetAPIVersion(), "kind", u.GetKind(), "name", u.GetName(), "fieldmanager", req.Options.FieldManager, "force", req.Options.Force)

	rctx, err := r.getAPIContext(bctx, u)
	if err != nil {
		return &resourcepb.Apply_Response{}, err
	}
	convertToInternal(rctx, u)

	dryrun := req.Options.DryRun
	if req.Options.Origin == "choreoctl" {
		dryrun = []string{"choreoctl"}
	}

	obj, err := rctx.Storage.Apply(ctx, u, &rest.ApplyOptions{
		DryRun:       dryrun,
		FieldManager: req.Options.FieldManager,
		Force:        req.Options.Force,
		Trace:        req.Options.Trace,
		Origin:       req.Options.Origin,
	})
	if err != nil {
		return &resourcepb.Apply_Response{}, err
	}

	if req.Options.Origin == "choreoctl" {
		if err := r.choreo.Store(orig); err != nil {
			return &resourcepb.Apply_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
		}
	}
	u = &unstructured.Unstructured{Object: obj.UnstructuredContent()}
	convertFromInternal(rctx, u)
	b, err := json.Marshal(u.Object)
	if err != nil {
		return &resourcepb.Apply_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	return &resourcepb.Apply_Response{Object: b}, nil
}

func (r *srv) Create(ctx context.Context, req *resourcepb.Create_Request) (*resourcepb.Create_Response, error) {
	log := log.FromContext(ctx)

	bctx, err := r.getBranchContext(req.Options.Branch)
	if err != nil {
		return &resourcepb.Create_Response{}, err
	}

	u, err := uobject.GetUnstructured(req.Object)
	if err != nil {
		return &resourcepb.Create_Response{}, err
	}

	log.Debug("create", "apiVersion", u.GetAPIVersion(), "kind", u.GetKind(), "name", u.GetName())

	rctx, err := r.getAPIContext(bctx, u)
	if err != nil {
		return &resourcepb.Create_Response{}, err
	}
	convertToInternal(rctx, u)
	obj, err := rctx.Storage.Create(ctx, u, &rest.CreateOptions{
		DryRun: req.Options.DryRun,
		Trace:  req.Options.Trace,
		Origin: req.Options.Origin,
	})
	if err != nil {
		return &resourcepb.Create_Response{}, err
	}

	u = &unstructured.Unstructured{Object: obj.UnstructuredContent()}
	convertFromInternal(rctx, u)
	b, err := json.Marshal(u.Object)
	if err != nil {
		return &resourcepb.Create_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	return &resourcepb.Create_Response{Object: b}, nil
}

func (r *srv) Update(ctx context.Context, req *resourcepb.Update_Request) (*resourcepb.Update_Response, error) {
	log := log.FromContext(ctx)

	bctx, err := r.getBranchContext(req.Options.Branch)
	if err != nil {
		return &resourcepb.Update_Response{}, err
	}

	u, err := uobject.GetUnstructured(req.Object)
	if err != nil {
		return &resourcepb.Update_Response{}, err
	}

	log.Debug("update", "apiVersion", u.GetAPIVersion(), "kind", u.GetKind(), "name", u.GetName())

	rctx, err := r.getAPIContext(bctx, u)
	if err != nil {
		return &resourcepb.Update_Response{}, err
	}
	convertToInternal(rctx, u)
	obj, err := rctx.Storage.Update(ctx, u, &rest.UpdateOptions{
		DryRun: req.Options.DryRun,
		Trace:  req.Options.Trace,
		Origin: req.Options.Origin,
	})
	if err != nil {
		return &resourcepb.Update_Response{}, err
	}

	u = &unstructured.Unstructured{Object: obj.UnstructuredContent()}
	convertFromInternal(rctx, u)
	b, err := json.Marshal(u.Object)
	if err != nil {
		return &resourcepb.Update_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	return &resourcepb.Update_Response{Object: b}, nil
}

func (r *srv) Delete(ctx context.Context, req *resourcepb.Delete_Request) (*resourcepb.Delete_Response, error) {
	log := log.FromContext(ctx)

	bctx, err := r.getBranchContext(req.Options.Branch)
	if err != nil {
		return &resourcepb.Delete_Response{}, err
	}

	u, err := uobject.GetUnstructured(req.Object)
	if err != nil {
		return &resourcepb.Delete_Response{}, err
	}

	log.Debug("delete", "apiVersion", u.GetAPIVersion(), "kind", u.GetKind(), "name", u.GetName())

	rctx, err := r.getAPIContext(bctx, u)
	if err != nil {
		return &resourcepb.Delete_Response{}, err
	}
	convertToInternal(rctx, u)

	dryrun := req.Options.DryRun
	if req.Options.Origin == "choreoctl" {
		dryrun = []string{"choreoctl"}
	}
	if _, err := rctx.Storage.Delete(ctx, u.GetName(), &rest.DeleteOptions{
		DryRun: dryrun,
		Trace:  req.Options.Trace,
		Origin: req.Options.Origin,
	}); err != nil {
		if !grpcerrors.IsNotFound(err) {
			return &resourcepb.Delete_Response{}, err
		}
	}

	if req.Options.Origin == "choreoctl" {
		if err := r.choreo.Destroy(u); err != nil {
			return &resourcepb.Delete_Response{}, status.Errorf(codes.Internal, "err: %s", err.Error())
		}
	}

	return &resourcepb.Delete_Response{}, nil
}

func (r *srv) Watch(req *resourcepb.Watch_Request, stream resourcepb.Resource_WatchServer) error {
	ctx := stream.Context()
	log := log.FromContext(ctx)

	/*
		if err := r.validatebranch(req.Options.Branch); err != nil {
			return err
		}
	*/
	bctx, err := r.getBranchContext(req.Options.Branch)
	if err != nil {
		return err
	}

	u, err := uobject.GetUnstructured(req.Object)
	if err != nil {
		return err
	}

	selector, err := selector.ResourceExprSelectorAsSelector(req.Options.ExprSelector)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid selector, err: %s", err.Error())
	}

	log.Debug("watch", "apiVersion", u.GetAPIVersion(), "kind", u.GetKind(), "options", req.Options)

	rctx, err := r.getAPIContext(bctx, u)
	if err != nil {
		return status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	convertToInternal(rctx, u)

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	wi, err := rctx.Storage.Watch(ctx, &rest.ListOptions{
		Watch:    false,
		Selector: selector,
		Trace:    req.Options.Trace,
		Origin:   req.Options.Origin,
	})
	if err != nil {
		return err
	}

	go r.watch(ctx, rctx, wi, stream, cancel)

	// context got cancelled -. proxy got stopped
	<-ctx.Done()
	log.Debug("grpc watch goroutine stopped")
	return nil
}

func (r *srv) watch(ctx context.Context, rctx *api.ResourceContext, wi watch.Interface, clientStream resourcepb.Resource_WatchServer, cancel func()) {
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

			if watchEvent.Type == resourcepb.Watch_ERROR {
				log.Error("received watch error", "event", watchEvent)
				cancel()
				continue
			}

			if watchEvent.Object == nil {
				log.Error("received nil object in watch event", "event", watchEvent)
				cancel()
				continue
			}

			// used for debug/display
			gvk := watchEvent.Object.GetObjectKind().GroupVersionKind()
			if !ok {
				log.Info("grpc watch done")
				return
			}
			log.Debug("grpc watch send event", "eventType", watchEvent.Type.String(), "gvk", gvk)

			u := &unstructured.Unstructured{Object: watchEvent.Object.UnstructuredContent()}
			convertFromInternal(rctx, u)
			b, err := json.Marshal(u.Object)
			if err != nil {
				log.Error("grpc watch failed to marshal object", "error", err.Error())
			}

			if err := clientStream.Send(&resourcepb.Watch_Response{
				Object:    b,
				EventType: watchEvent.Type,
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
