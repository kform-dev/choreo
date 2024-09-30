package choreo

import (
	"context"
	"fmt"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/choreo/crdloader"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// No longer used, we decided that apis should always be loaded through files/git
type Reconciler interface {
	Start(ctx context.Context)
	Stop()
}

func NewAPIReconciler(
	choreoInstance ChoreoInstance,
	client resourceclient.Client,
	branch string,
	apiStore *api.APIStore,
) Reconciler {
	return &reconciler{
		choreoInstance: choreoInstance,
		client:         client,
		branch:         branch,
		apiStore:       apiStore,
	}
}

type reconciler struct {
	choreoInstance ChoreoInstance
	client         resourceclient.Client
	branch         string
	apiStore       *api.APIStore
	// dynamic
	cancel func()
}

func (r *reconciler) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *reconciler) Start(ctx context.Context) {
	log := log.FromContext(ctx)
	ctx, r.cancel = context.WithCancel(ctx)

	u := &unstructured.Unstructured{}
	u.SetAPIVersion(apiextensionsv1.SchemeGroupVersion.Identifier())
	u.SetKind("CustomResourceDefinition")
	ch := r.client.Watch(ctx, u, &resourceclient.ListOptions{
		Branch:            r.branch,
		ShowManagedFields: false,
		ExprSelector: &resourcepb.ExpressionSelector{
			Match: map[string]string{},
		},
	})
	for {
		select {
		case <-ctx.Done():
			r.Stop()
		case rsp, ok := <-ch:
			if !ok {
				r.Stop()
				return
			}

			crd := &apiextensionsv1.CustomResourceDefinition{}
			if err := yaml.Unmarshal(rsp.Object, crd); err != nil {
				log.Error("cannot unmarchal reconciler resource", "error", err)
				return
			}
			fmt.Println("api reconciler event", r.branch, rsp.EventType.String(), crd.GetName())

			// TODO handle internal apis

			resctx, err := crdloader.LoadCRD(ctx, r.choreoInstance.GetPathInRepo(), r.choreoInstance.GetDBPath(), crd, nil)
			if err != nil {
				log.Error("cannot load crd resource", "error", err)
				return
			}

			switch rsp.EventType {
			case resourcepb.Watch_ADDED, resourcepb.Watch_MODIFIED:
				if err := r.apiStore.Apply(resctx.GVK(), resctx); err != nil {
					log.Error("cannot apply api to apistore", "error", err)
				}
			case resourcepb.Watch_DELETED:
				if err := r.apiStore.Delete(resctx.GVK()); err != nil {
					log.Error("cannot delete api from apistore", "error", err)
				}
			}
		}
	}
}
