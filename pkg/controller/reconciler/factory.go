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

	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/controller/informers"
	"github.com/kform-dev/choreo/pkg/proto/runnerpb"
	//reconcilerstore "github.com/kform-dev/choreo/pkg/controller/reconciler/store"
)

type ReconcilerFactory interface {
	Start(ctx context.Context)
	Stop()
}

func NewReconcilerFactory(
	ctx context.Context,
	client resourceclient.Client,
	informerFactory informers.InformerFactory,
	reconcilersConfigs []*choreov1alpha1.Reconciler,
	libraries []*choreov1alpha1.Library,
	resultCh chan *runnerpb.ReconcileResult,
	branchName string,
) (ReconcilerFactory, error) {

	reconcilers, err := newReconcilers(
		ctx,
		client,
		informerFactory,
		reconcilersConfigs,
		libraries,
		resultCh,
		branchName,
	)

	return &reconcilerFactory{
		client:      client,
		reconcilers: reconcilers,
	}, err
}

type reconcilerFactory struct {
	client      resourceclient.Client
	reconcilers *reConcilers
}

func (r *reconcilerFactory) Start(ctx context.Context) {
	r.reconcilers.start(ctx)
}

func (r *reconcilerFactory) Stop() {
	r.reconcilers.stop()
}
