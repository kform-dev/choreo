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
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/henderiw/logger/log"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/controller/collector"
	"github.com/kform-dev/choreo/pkg/controller/informers"
	"github.com/kform-dev/choreo/pkg/controller/reconciler"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/choreo/apiloader"
	"github.com/kform-dev/choreo/pkg/server/choreo/instance"
	"github.com/kform-dev/choreo/pkg/server/choreo/loader"
	"github.com/kform-dev/choreo/pkg/util/inventory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Runner interface {
	//AddResourceClientAndContext(ctx context.Context, client resourceclient.Client)
	Start(ctx context.Context, bctx *BranchCtx) (*runnerpb.Start_Response, error)
	Stop()
	RunOnce(ctx context.Context, bctx *BranchCtx, stream runnerpb.Runner_OnceServer) error
	Load(ctx context.Context, bctx *BranchCtx) error
}

func NewRunner(choreo Choreo) Runner {
	return &run{
		choreo: choreo,
	}
}

type run struct {
	m      sync.RWMutex
	status RunnerStatus
	//
	choreo Choreo
	// get ctx -> r.choreo.GetContext()
	// get client -> r.choreo.GetClient()
	// get commit -> r.choreo.status.Get().RootChoreoInstance.GetRepo().GetBranchCommit(branchObj.Name)
	// added dynamically before start
	//discoveryClient discovery.CachedDiscoveryInterface
	// dynamic data
	cancel context.CancelFunc
	//reconcilerConfigs  []*choreov1alpha1.Reconciler
	//libraries          []*choreov1alpha1.Library
	//reconcilerResultCh chan *runnerpb.Result
	//runResultCh        chan *runnerpb.Once_Response
	//collector          collector.Collector
	//informerfactory    informers.InformerFactory
	oncerspChan chan *runnerpb.Once_Response
}

func (r *run) Start(ctx context.Context, bctx *BranchCtx) (*runnerpb.Start_Response, error) {
	if r.getStatus() == RunnerStatus_Running {
		return &runnerpb.Start_Response{}, nil
	}
	if r.getStatus() == RunnerStatus_Once {
		return &runnerpb.Start_Response{},
			status.Errorf(codes.InvalidArgument, "runner is already running, status %s", r.status.String())
	}

	if err := r.Load(ctx, bctx); err != nil {
		return &runnerpb.Start_Response{}, err
	}

	rootChoreoInstance := r.choreo.GetRootChoreoInstance()
	reconcilers := []*choreov1alpha1.Reconciler{}
	libraries := []*choreov1alpha1.Library{}
	for _, childChoreoInstance := range rootChoreoInstance.GetChildren() {
		if childChoreoInstance.IsRootInstance() {
			reconcilers = append(reconcilers, childChoreoInstance.GetReconcilers()...)
			libraries = append(libraries, childChoreoInstance.GetLibraries()...)
		}
	}
	reconcilers = append(reconcilers, rootChoreoInstance.GetReconcilers()...)
	libraries = append(libraries, rootChoreoInstance.GetLibraries()...)

	// we use the server context to cancel/handle the status of the server
	// since the ctx we get is from the client
	runctx, cancel := context.WithCancel(r.choreo.GetContext())
	r.setStatusAndCancel(RunnerStatus_Running, cancel)

	go func() {
		defer r.Stop()
		select {
		case <-runctx.Done():
			return
		default:
			// use runctx since the ctx is from the cmd and it will be cancelled upon completion
			r.runReconciler(runctx, "root", bctx, reconcilers, libraries, false) // false -> run continuously, not once
		}
	}()
	return &runnerpb.Start_Response{}, nil
}

func (r *run) Stop() {
	if r.getCancel() != nil {
		r.cancel() // Cancel the context, which triggers stopping in the StartContinuous loop
	}
	r.setStatusAndCancel(RunnerStatus_Stopped, nil)
	// don't nilify the other resources, since they will be reinitialized
}

func (r *run) runOnce(ctx context.Context, stream runnerpb.Runner_OnceServer) error {
	log := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			log.Debug("grpc watch stopped, stopping storage watch")
			close(r.oncerspChan)
			return nil
		case rsp, ok := <-r.oncerspChan:
			if !ok {
				log.Debug("response channel closed, stopping once runner")
			}
			if err := stream.Send(rsp); err != nil {
				p, _ := peer.FromContext(stream.Context())
				addr := "unknown"
				if p != nil {
					addr = p.Addr.String()
				}
				log.Error("grpc watch send stream failed", "client", addr)
			}
		}
	}
}

func (r *run) RunOnce(ctx context.Context, bctx *BranchCtx, stream runnerpb.Runner_OnceServer) error {
	log := log.FromContext(ctx)
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	r.oncerspChan = make(chan *runnerpb.Once_Response)

	go r.runOnce(ctx, stream)

	r.once(ctx, bctx)
	<-ctx.Done()
	log.Debug("grpc runOnce goroutine stopped")
	return nil
}

func (r *run) onceResponseCompleted() {
	if r.oncerspChan == nil {
		return
	}
	r.oncerspChan <- &runnerpb.Once_Response{
		Type: runnerpb.Once_COMPLETED,
	}

}

func (r *run) onceResponseProgressUpdate(msg string) {
	if r.oncerspChan == nil {
		return
	}
	r.oncerspChan <- &runnerpb.Once_Response{
		Type: runnerpb.Once_PROGRESS_UPDATE,
		Data: &runnerpb.Once_Response_ProgressUpdate{
			ProgressUpdate: &runnerpb.Once_ProgressUpdate{Message: msg},
		},
	}
}

func (r *run) onceResponseError(msg string) {
	if r.oncerspChan == nil {
		return
	}
	r.oncerspChan <- &runnerpb.Once_Response{
		Type: runnerpb.Once_ERROR,
		Data: &runnerpb.Once_Response_Error{
			Error: &runnerpb.Once_Error{Message: msg},
		},
	}
}

func (r *run) onceResponseRunResult(rsp *runnerpb.Once_Response_RunResponse) {
	if r.oncerspChan == nil {
		return
	}
	r.oncerspChan <- &runnerpb.Once_Response{
		Type: runnerpb.Once_RUN_RESPONSE,
		Data: rsp,
	}
}

/*
func (r *run) onceResponseSDCResult(rsp *runnerpb.Once_Response_SdcResponse) {
	r.oncerspChan <- &runnerpb.Once_Response{
		Type: runnerpb.Once_SDC_RESPONSE,
		Data: rsp,
	}
}
*/

func (r *run) once(ctx context.Context, bctx *BranchCtx) {
	log := log.FromContext(ctx)
	if r.getStatus() != RunnerStatus_Stopped {
		r.onceResponseError(fmt.Sprintf("runner is already running, status %s", r.status.String()))
		return
	}
	r.onceResponseProgressUpdate("loading ...")
	if err := r.Load(ctx, bctx); err != nil {
		log.Error("loading failed", "err", err)
		r.onceResponseError(err.Error())
		return
	}
	r.onceResponseProgressUpdate("loading done")

	// we use the server context to cancel/handle the status of the server
	// since the ctx we get is from the client
	_, cancel := context.WithCancel(r.choreo.GetContext())
	r.setStatusAndCancel(RunnerStatus_Once, cancel)

	defer r.Stop()

	r.onceResponseProgressUpdate("running reconcilers ...")
	rsp, err := r.runReconcilers(ctx, bctx, true) // run once
	if err != nil {
		r.onceResponseError(err.Error())
		return
	}
	r.onceResponseRunResult(rsp)

	if r.choreo.GetConfig().ServerFlags.SDC != nil && *r.choreo.GetConfig().ServerFlags.SDC {
		r.onceResponseProgressUpdate("running config validator ...")
		configValidator := NewConfigValidator(r.choreo)
		if err := configValidator.runConfigValidation(ctx, bctx); err != nil {
			r.onceResponseError(err.Error())
			return
		}
	}

	r.createSnapshot(ctx, bctx, rsp)
	r.onceResponseCompleted()
}

// loads upstream refs, apis, reconcilers, data and garbage collect
func (r *run) Load(ctx context.Context, branchCtx *BranchCtx) error {
	// we only work with checkout branch
	rootChoreoInstance := r.choreo.GetRootChoreoInstance()

	if r.choreo.GetConfig().ServerFlags.SDC != nil && *r.choreo.GetConfig().ServerFlags.SDC {
		if err := r.loadSchemas(ctx, branchCtx, rootChoreoInstance); err != nil {
			fmt.Println("schema load error", err)
			return err
		}
	}
	// we reinitialize from scratch
	rootChoreoInstance.InitChildren()
	if err := r.loadUpstreamRefs(ctx, branchCtx, rootChoreoInstance); err != nil {
		return err
	}
	// we load the apis to the global context
	rootChoreoInstance.InitAPIs()
	apis := rootChoreoInstance.GetAPIs()
	if err := r.loadAPIs(ctx, branchCtx, rootChoreoInstance, apis); err != nil {
		fmt.Println("apis", err)
		return err
	}

	rootChoreoInstance.InitLibraries()
	libraries := rootChoreoInstance.GetLibraries()
	if err := r.loadLibraries(ctx, branchCtx, rootChoreoInstance, libraries); err != nil {
		return err
	}

	rootChoreoInstance.InitReconcilers()
	reconcilers := rootChoreoInstance.GetReconcilers()
	if err := r.loadReconcilers(ctx, branchCtx, rootChoreoInstance, reconcilers); err != nil {
		return err
	}

	// load and update the global apis
	for _, childChoreoInstance := range rootChoreoInstance.GetChildren() {
		branchCtx.APIStore.Import(childChoreoInstance.GetAPIs())
		for _, childChoreoInstance := range childChoreoInstance.GetChildren() {
			branchCtx.APIStore.Import(childChoreoInstance.GetAPIs())
		}
	}
	branchCtx.APIStore.Import(rootChoreoInstance.GetAPIs())

	if err := r.choreo.GetBranchStore().UpdateBranchCtx(branchCtx); err != nil {
		return err
	}
	apiResources := branchCtx.APIStore.GetAPIResources()
	/*
		for _, apiResource := range apiResources {
			fmt.Println("apiResource", apiResource.Kind)
		}
	*/

	for _, childChoreoInstance := range rootChoreoInstance.GetChildren() {
		if childChoreoInstance.IsRootInstance() {
			// load the data
			if err := r.loadData(ctx, branchCtx, childChoreoInstance, childChoreoInstance.GetAPIs().GetExternalGVKSet().UnsortedList()); err != nil {
				return err
			}
		}
	}
	if err := r.loadData(ctx, branchCtx, rootChoreoInstance, branchCtx.APIStore.GetExternalGVKSet().UnsortedList()); err != nil {
		return err
	}

	// run the garbage collection to cleanup childObjects for root Objects that got deleted
	// since the reconcilers dont run all the time this is needed
	inv := inventory.Inventory{}
	if err := inv.Build(ctx, r.choreo.GetClient(), apiResources, &inventory.BuildOptions{
		ShowManagedField: true,
		Branch:           branchCtx.Branch,
		ShowChoreoAPIs:   false, // TO CHANGE
	}); err != nil {
		return status.Errorf(codes.Internal, "err: %s", err.Error())
	}

	var errm error
	garbageSet := inv.CollectGarbage()
	for _, ref := range garbageSet.UnsortedList() {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind))
		u.SetName(ref.Name)
		u.SetNamespace(ref.Namespace)

		fmt.Println("delete garbage", u.GetAPIVersion(), u.GetKind(), u.GetName())

		if err := r.choreo.GetClient().Delete(ctx, u, &resourceclient.DeleteOptions{
			Branch: branchCtx.Branch,
		}); err != nil {
			errm = errors.Join(errm, err)
		}
	}
	return errm

}

func (r *run) loadSchemas(ctx context.Context, _ *BranchCtx, choreoInstance instance.ChoreoInstance) error {
	schemaloader := loader.SchemaLoader{
		Parent: choreoInstance,
		Cfg:    r.choreo.GetConfig(),
		//Client:     r.choreo.GetClient(), // used to upload the upstream ref
		//Branch:     branchCtx.Branch,
		RepoPath:   choreoInstance.GetRepoPath(),
		PathInRepo: choreoInstance.GetPathInRepo(),
		TempDir:    r.choreo.GetRootChoreoInstance().GetTempPath(), // this is the temppath of the rootInstance
	}
	// this loads schema to the schemastore
	if err := schemaloader.Load(ctx); err != nil {
		return err
	}
	return nil
}

func (r *run) loadUpstreamRefs(ctx context.Context, branchCtx *BranchCtx, choreoInstance instance.ChoreoInstance) error {
	upstreamloader := loader.UpstreamLoader{
		Parent:     choreoInstance,
		Cfg:        r.choreo.GetConfig(),
		Client:     r.choreo.GetClient(), // used to upload the upstream ref
		Branch:     branchCtx.Branch,
		RepoPath:   choreoInstance.GetRepoPath(),
		PathInRepo: choreoInstance.GetPathInRepo(),
		TempDir:    r.choreo.GetRootChoreoInstance().GetTempPath(), // this is the temppath of the rootInstance
		ProgressFn: r.onceResponseProgressUpdate,
	}
	// this loads additional choreoinstances
	if err := upstreamloader.Load(ctx); err != nil {
		return err
	}

	var errs error
	for _, choreoinstance := range choreoInstance.GetChildren() {
		if err := r.loadUpstreamRefs(ctx, branchCtx, choreoinstance); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (r *run) loadAPIs(ctx context.Context, branchCtx *BranchCtx, choreoInstance instance.ChoreoInstance, apiStore *api.APIStore) error {
	// load api files to apistore and apiserver
	if choreoInstance.IsRootInstance() {
		choreoInstance.InitAPIs()
		apiStore = choreoInstance.GetAPIs()
	}

	rootChoreoInstance := r.choreo.GetRootChoreoInstance()
	loader := &apiloader.APILoaderFile2APIStoreAndAPI{
		Cfg:      r.choreo.GetConfig(),
		Client:   r.choreo.GetClient(),
		APIStore: apiStore,
		Branch:   branchCtx.Branch,
		// InternalGVKs are only kept in the root choreo instance
		InternalGVKs: rootChoreoInstance.GetInternalAPIStore().GetExternalGVKSet(),
		RepoPath:     choreoInstance.GetRepoPath(),
		PathInRepo:   choreoInstance.GetPathInRepo(),
		DBPath:       rootChoreoInstance.GetDBPath(),
	}
	// TBD if we need to use the commit loader or not
	if err := loader.Load(ctx); err != nil {
		return err
	}
	choreoInstance.AddAPIs(loader.APIStore)
	for _, choreoinstance := range choreoInstance.GetChildren() {
		r.loadAPIs(ctx, branchCtx, choreoinstance, apiStore)
	}
	return nil
}

func (r *run) loadLibraries(ctx context.Context, branchCtx *BranchCtx, choreoInstance instance.ChoreoInstance, libraries []*choreov1alpha1.Library) error {
	// load api files to apistore and apiserver
	//rootChoreoInstance := r.Choreo.status.Get().RootChoreoInstance
	// overwrite libraries when we are a root instance
	if choreoInstance.IsRootInstance() {
		choreoInstance.InitLibraries()
		libraries = choreoInstance.GetLibraries()
	}

	devloader := &loader.DevLoader{
		Cfg:        r.choreo.GetConfig(),
		RepoPath:   choreoInstance.GetRepoPath(),
		PathInRepo: choreoInstance.GetPathInRepo(),
		Libraries:  libraries,
	}

	if err := devloader.LoadLibraries(ctx); err != nil {
		return err
	}
	choreoInstance.AddLibraries(devloader.Libraries...)
	for _, choreoinstance := range choreoInstance.GetChildren() {
		r.loadLibraries(ctx, branchCtx, choreoinstance, libraries)
	}
	return nil
}

func (r *run) loadReconcilers(ctx context.Context, branchCtx *BranchCtx, choreoInstance instance.ChoreoInstance, reconcilers []*choreov1alpha1.Reconciler) error {
	// overwrite reconcilers when we are a root instance since we start from scratch
	if choreoInstance.IsRootInstance() {
		choreoInstance.InitReconcilers()
		reconcilers = choreoInstance.GetReconcilers()
	}

	devloader := &loader.DevLoader{
		Cfg:         r.choreo.GetConfig(),
		RepoPath:    choreoInstance.GetRepoPath(),
		PathInRepo:  choreoInstance.GetPathInRepo(),
		Reconcilers: reconcilers,
	}

	if err := devloader.LoadReconcilers(ctx); err != nil {
		return err
	}

	choreoInstance.AddReconcilers(devloader.Reconcilers...)
	for _, choreoinstance := range choreoInstance.GetChildren() {
		r.loadReconcilers(ctx, branchCtx, choreoinstance, reconcilers)
	}
	return nil
}

func (r *run) loadData(ctx context.Context, branchCtx *BranchCtx, choreoInstance instance.ChoreoInstance, gvks []schema.GroupVersionKind) error {
	rootChoreoInstance := r.choreo.GetRootChoreoInstance()

	annotation := choreov1alpha1.FileLoaderAnnotation.String()
	if upstreamRef := choreoInstance.GetUpstreamRef(); upstreamRef != nil {
		annotation = upstreamRef.LoaderAnnotation().String()
	}

	dataloader := &loader.DataLoader{
		Cfg:        r.choreo.GetConfig(),
		Client:     r.choreo.GetClient(),
		Branch:     branchCtx.Branch,
		GVKs:       gvks,
		RepoPth:    choreoInstance.GetRepoPath(),
		PathInRepo: choreoInstance.GetPathInRepo(),
		//APIStore:       branchCtx.APIStore,
		InternalAPISet: rootChoreoInstance.GetInternalAPIStore().GetExternalGVKSet(),
		Annotation:     annotation,
	}
	return dataloader.Load(ctx)
}

func PrintAPIResource(apiResources []*discoverypb.APIResource) bool {
	for _, apiResource := range apiResources {
		gvk := schema.GroupVersionKind{
			Group:   apiResource.Group,
			Version: apiResource.Version,
			Kind:    apiResource.Kind,
		}
		fmt.Println("print apiresources", gvk.String())
	}
	return false
}

func HasAPIResource(apiResources []*discoverypb.APIResource, gvk schema.GroupVersionKind) bool {
	for _, apiResource := range apiResources {
		if apiResource.Group == gvk.Group && apiResource.Version == gvk.Version && apiResource.Kind == gvk.Kind {
			return true
		}
	}
	return false
}

func (r *run) runReconcilers(ctx context.Context, branchCtx *BranchCtx, once bool) (*runnerpb.Once_Response_RunResponse, error) {
	rsp := &runnerpb.Once_Response_RunResponse{
		RunResponse: &runnerpb.Once_RunResponse{
			Success: true,
			Results: []*runnerpb.Once_RunResult{},
		},
	}

	rootChoreoInstance := r.choreo.GetRootChoreoInstance()
	rootLibraries := []*choreov1alpha1.Library{}
	for _, childChoreoInstance := range rootChoreoInstance.GetChildren() {
		// gather the libraries - libraries are used for the child choreo instance runner
		libraries := []*choreov1alpha1.Library{}
		// gather the crd childinstance libraries if present
		for _, childChoreoInstance := range childChoreoInstance.GetChildren() {
			libraries = append(libraries, childChoreoInstance.GetLibraries()...)
		}
		libraries = append(libraries, childChoreoInstance.GetLibraries()...)
		rootLibraries = append(rootLibraries, libraries...)
		// only run the reconciler when there are reconcilers
		if len(childChoreoInstance.GetReconcilers()) > 0 {
			r.onceResponseProgressUpdate(fmt.Sprintf("running child reconciler %s", childChoreoInstance.GetUpstreamRef().GetName()))
			upstreamRefName := childChoreoInstance.GetUpstreamRef().GetName()
			runrsp, err := r.runReconciler(
				ctx,
				upstreamRefName,
				branchCtx,
				childChoreoInstance.GetReconcilers(),
				libraries,
				once,
			) // run once
			if err != nil {
				rsp.RunResponse.Success = false
				rsp.RunResponse.Results = append(rsp.RunResponse.Results, runrsp)
				return rsp, err
			}
			if !runrsp.Success {
				rsp.RunResponse.Success = false
			}
			rsp.RunResponse.Results = append(rsp.RunResponse.Results, runrsp)
		}
	}
	rootLibraries = append(rootLibraries, rootChoreoInstance.GetLibraries()...)
	if len(rootChoreoInstance.GetReconcilers()) > 0 {
		r.onceResponseProgressUpdate(fmt.Sprintf("running root reconciler %s", rootChoreoInstance.GetName()))
		runrsp, err := r.runReconciler(
			ctx,
			"root",
			branchCtx,
			rootChoreoInstance.GetReconcilers(),
			rootLibraries,
			once,
		) // run once
		if err != nil {
			rsp.RunResponse.Success = false
			rsp.RunResponse.Results = append(rsp.RunResponse.Results, runrsp)
			return rsp, err
		}
		if !runrsp.Success {
			rsp.RunResponse.Success = false
		}
		rsp.RunResponse.Results = append(rsp.RunResponse.Results, runrsp)
	}
	return rsp, nil

}

func (r *run) runReconciler(ctx context.Context, ref string, branchCtx *BranchCtx, reconcilers []*choreov1alpha1.Reconciler, libraries []*choreov1alpha1.Library, once bool) (*runnerpb.Once_RunResult, error) {
	reconcilerGVKs := sets.New[schema.GroupVersionKind]()
	for _, reconciler := range reconcilers {
		reconcilerGVKs.Insert(reconciler.GetGVKs().UnsortedList()...)
	}

	reconcilerResultCh := make(chan *runnerpb.ReconcileResult)
	runResultCh := make(chan *runnerpb.Once_RunResult)
	collector := collector.New(reconcilerResultCh, runResultCh)
	informerfactory := informers.NewInformerFactory(r.choreo.GetClient(), reconcilerGVKs, branchCtx.Branch)

	reconcilerfactory, err := reconciler.NewReconcilerFactory(
		ctx,
		r.choreo.GetClient(),
		informerfactory,
		reconcilers,
		libraries,
		reconcilerResultCh,
		branchCtx.Branch,
	)

	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		go collector.Start(ctx, once)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		go reconcilerfactory.Start(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		go informerfactory.Start(ctx)
	}()

	select {
	case result, ok := <-runResultCh:
		if !ok {
			return nil, nil
		}
		result.ReconcilerRunner = ref
		return result, nil

	case <-ctx.Done():
		wg.Wait()
		return nil, nil
	}
}

func (r *run) createSnapshot(ctx context.Context, bctx *BranchCtx, rsp *runnerpb.Once_Response_RunResponse) error {
	uid := uuid.New().String()

	apiResources := bctx.APIStore.GetAPIResources()

	inv := inventory.Inventory{}
	if err := inv.Build(ctx, r.choreo.GetClient(), apiResources, &inventory.BuildOptions{
		ShowManagedField: true,
		Branch:           bctx.Branch,
		ShowChoreoAPIs:   true,
	}); err != nil {
		return status.Errorf(codes.Internal, "err: %s", err.Error())
	}
	fmt.Println("create snapshot", uid)

	r.choreo.SnapshotManager().Create(uid, apiResources, inv, rsp)

	return nil
}

func (r *run) getStatus() RunnerStatus {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.status
}

func (r *run) setStatusAndCancel(status RunnerStatus, cancel func()) {
	r.m.Lock()
	defer r.m.Unlock()
	r.status = status
	r.cancel = cancel
}

func (r *run) getCancel() func() {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.cancel
}
