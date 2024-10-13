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

	"github.com/kform-dev/choreo/pkg/proto/choreopb"
	"github.com/kform-dev/choreo/pkg/server/choreo"
)

func New(choreo choreo.Choreo) choreopb.ChoreoServer {
	return &srv{
		choreo: choreo,
	}
}

type srv struct {
	choreopb.UnimplementedChoreoServer
	choreo choreo.Choreo
}

func (r *srv) Get(ctx context.Context, req *choreopb.Get_Request) (*choreopb.Get_Response, error) {
	return r.choreo.Get(ctx, req)
}

func (r *srv) Apply(ctx context.Context, req *choreopb.Apply_Request) (*choreopb.Apply_Response, error) {
	return r.choreo.Apply(ctx, req)
}
