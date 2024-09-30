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

package reconciler

import (
	"context"

	"github.com/henderiw/logger/log"
	"github.com/henderiw/store"
	"github.com/henderiw/store/memory"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/controller/collector/result"
	"github.com/kform-dev/choreo/pkg/controller/informers"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type reConcilers struct {
	reconcilers store.Storer[Reconciler]

	cancel func()
}

func newReconcilers(
	ctx context.Context,
	client resourceclient.Client,
	informerFactory informers.InformerFactory,
	reconcilerConfigs []*choreov1alpha1.Reconciler,
	libs *unstructured.UnstructuredList,
	resultCh chan result.Result,
	branchName string,
) (*reConcilers, error) {
	reconcilers := memory.NewStore[Reconciler](nil)
	var errm error
	for _, reconcilerConfig := range reconcilerConfigs {
		log := log.FromContext(ctx)
		if err := reconcilers.Create(store.ToKey(reconcilerConfig.GetName()), newReconciler(
			ctx,
			reconcilerConfig.GetName(),
			client,
			informerFactory,
			reconcilerConfig,
			libs,
			resultCh,
			branchName,
		)); err != nil {
			log.Error("unexpected error, duplicate application name")
		}
	}

	return &reConcilers{
		reconcilers: reconcilers,
	}, errm
}

func (r *reConcilers) start(ctx context.Context) {
	log := log.FromContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	r.reconcilers.List(func(k store.Key, r Reconciler) {
		go func() {
			go r.start(ctx)
		}()
	})

	<-ctx.Done()
	log.Debug("reconcilers stopped...")
	r.cancel()
}

func (r *reConcilers) stop() {
	if r.cancel != nil {
		r.cancel()
	}
}
