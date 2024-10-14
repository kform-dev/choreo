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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/repository/repogit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

type Choreo interface {
	Get(ctx context.Context, req *choreopb.Get_Request) (*choreopb.Get_Response, error)
	Apply(ctx context.Context, req *choreopb.Apply_Request) (*choreopb.Apply_Response, error)
	Start(ctx context.Context)
	GetRootChoreoInstance() ChoreoInstance
	GetBranchStore() *BranchStore
	Runner() Runner
	SnapshotManager() *SnapshotManager
	// updates resource (yaml) in the input directory
	Store(obj runtime.Unstructured) error
	// remove resource (yaml) from the input directory
	Destroy(obj runtime.Unstructured) error
}

func New(flags *genericclioptions.ConfigFlags) Choreo {
	r := &choreo{
		status: &Status{},
		flags:  flags,
	}
	r.status.Set(Initializing())
	r.branchStore = NewBranchStore(r)
	r.runner = NewRunner(r)
	r.snapshotMgr = NewSnapshotManager()
	return r
}

type choreo struct {
	status      *Status
	branchStore *BranchStore
	runner      Runner
	snapshotMgr *SnapshotManager
	flags       *genericclioptions.ConfigFlags

	client resourceclient.Client
}

func (r *choreo) Get(ctx context.Context, req *choreopb.Get_Request) (*choreopb.Get_Response, error) {
	status := r.status.Get()
	return &choreopb.Get_Response{
		ChoreoContext: status.ChoreoCtx,
		Status:        status.Status,
		Reason:        status.Reason,
	}, nil

}

func (r *choreo) Apply(ctx context.Context, req *choreopb.Apply_Request) (*choreopb.Apply_Response, error) {
	log := log.FromContext(ctx)
	if req.ChoreoContext.Path != "" {
		// we dont deal with change in this case
		rootChoreoInstance, err := NewRootChoreoInstance(ctx, &Config{Path: req.ChoreoContext.Path, Flags: r.flags})
		if err != nil {
			r.status.Set(Failed(err.Error()))
			return &choreopb.Apply_Response{}, status.Errorf(codes.InvalidArgument, "err: %s", err.Error())
		}
		r.status.Set(Success(rootChoreoInstance, req.ChoreoContext))
		return &choreopb.Apply_Response{}, nil

	} else {
		// dynamic
		if !r.status.Changed(req.ChoreoContext) {
			return &choreopb.Apply_Response{}, nil
		}

		// stop the runner if it was active - safe call
		r.Runner().Stop()
		// ...
		rootChoreoInstance := r.status.Get().RootChoreoInstance
		if rootChoreoInstance != nil {
			if err := rootChoreoInstance.Destroy(); err != nil {
				log.Error("destroy failed", "error", err)
			}
		}
		r.status.Set(Failed("reinitializing"))

		url := req.ChoreoContext.Url
		replace := strings.NewReplacer("/", "-", ":", "-")
		childRepoPath := filepath.Join(".", replace.Replace(url))

		log.Info("apply new choreo context", "url", url, "childRepoPath", childRepoPath, "ref", req.ChoreoContext.Ref)

		repo, commit, err := repogit.NewUpstreamRepo(ctx, childRepoPath, url, req.ChoreoContext.Ref)
		if err != nil {
			r.status.Set(Failed(err.Error()))
			return &choreopb.Apply_Response{}, status.Errorf(codes.InvalidArgument, "err: %s", err.Error())
		}
		rootChoreoInstance, err = NewRootChoreoInstance(ctx, &Config{
			Flags:      r.flags,
			Path:       childRepoPath,
			Repo:       repo,
			Commit:     commit,
			PathInRepo: req.ChoreoContext.Directory,
		})
		if err != nil {
			r.status.Set(Failed(err.Error()))
			return &choreopb.Apply_Response{}, status.Errorf(codes.InvalidArgument, "err: %s", err.Error())
		}
		r.branchStore = NewBranchStore(r)
		r.status.Set(Success(rootChoreoInstance, req.ChoreoContext))
		return &choreopb.Apply_Response{}, nil
	}
}

func (r *choreo) GetBranchStore() *BranchStore {
	return r.branchStore
}

func (r *choreo) GetRootChoreoInstance() ChoreoInstance {
	return r.status.Get().RootChoreoInstance
}

func (r *choreo) Runner() Runner {
	return r.runner
}

func (r *choreo) SnapshotManager() *SnapshotManager {
	return r.snapshotMgr
}

func (r *choreo) Start(ctx context.Context) {
	log := log.FromContext(ctx)
	var err error
	r.client, err = r.flags.ToResourceClient()
	if err != nil {
		panic(err)
	}
	r.runner.AddResourceClientAndContext(ctx, r.client)
	r.branchStore.store.Start(ctx)
	defer r.branchStore.store.Stop()

	// Ticker to check the repository every 10 seconds
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	if err := r.updateBranches(ctx); err != nil {
		log.Error("update branches failed", "err", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := r.updateBranches(ctx); err != nil {
				log.Error("update branches failed", "err", err)
			}

		case <-ctx.Done():
			time.Sleep(1 * time.Second)
			log.Info("choreo done")
			return
		}
	}
}

func (r *choreo) updateBranches(ctx context.Context) error {
	status := r.status.Get()
	rootChoreoInstance := status.RootChoreoInstance

	if rootChoreoInstance == nil {
		return nil
	}
	if err := r.branchStore.Delete(ctx, rootChoreoInstance.GetRepo().GetBranchSet()); err != nil {
		return err
	}
	if err := r.branchStore.Update(ctx, rootChoreoInstance.GetRepo().GetBranches()); err != nil {
		return err
	}

	if status.ChoreoCtx.Continuous {
		bctx, err := r.branchStore.GetCheckedOut()
		if err != nil {
			return err
		}
		r.Runner().Start(ctx, bctx)
	}
	return nil
}

func (r *choreo) Store(obj runtime.Unstructured) error {
	u := &unstructured.Unstructured{
		Object: obj.UnstructuredContent(),
	}
	gv, err := schema.ParseGroupVersion(u.GetAPIVersion())
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(u.Object)
	if err != nil {
		return err
	}

	rootChoreoInstance := r.GetRootChoreoInstance()
	path := filepath.Join(
		rootChoreoInstance.GetRepoPath(),
		rootChoreoInstance.GetPathInRepo(),
		*rootChoreoInstance.GetFlags().InputPath,
		fmt.Sprintf("%s.%s.%s.yaml",
			gv.Group,
			strings.ToLower(u.GetKind()),
			u.GetName(),
		),
	)
	return os.WriteFile(path, b, 0644)
}

func (r *choreo) Destroy(obj runtime.Unstructured) error {
	u := &unstructured.Unstructured{
		Object: obj.UnstructuredContent(),
	}
	gv, err := schema.ParseGroupVersion(u.GetAPIVersion())
	if err != nil {
		return err
	}

	rootChoreoInstance := r.GetRootChoreoInstance()
	fileName := filepath.Join(
		rootChoreoInstance.GetRepoPath(),
		rootChoreoInstance.GetPathInRepo(),
		*rootChoreoInstance.GetFlags().InputPath,
		fmt.Sprintf("%s.%s.%s.yaml",
			gv.Group,
			strings.ToLower(u.GetKind()),
			u.GetName(),
		))

	return os.Remove(fileName)
}
