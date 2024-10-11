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

package choreoctx

import (
	"github.com/kform-dev/choreo/pkg/client/go/branchclient"
	"github.com/kform-dev/choreo/pkg/client/go/choreoclient"
	"github.com/kform-dev/choreo/pkg/client/go/discoveryclient"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
)

type ChoreoCtx struct {
	Ready           bool
	BranchClient    branchclient.BranchClient
	DiscoveryClient discoveryclient.DiscoveryClient
	ResourceClient  resourceclient.ResourceClient
	ChoreoClient    choreoclient.ChoreoClient
}
