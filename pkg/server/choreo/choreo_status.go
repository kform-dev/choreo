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
	"fmt"
	"sync"

	"github.com/kform-dev/choreo/pkg/proto/choreopb"
)

type ChoreoStatus struct {
	Status             bool
	Reason             string
	RootChoreoInstance ChoreoInstance
	ChoreoCtx          *choreopb.ChoreoContext
}

type Status struct {
	m sync.RWMutex
	ChoreoStatus
}

func (r *Status) Get() ChoreoStatus {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.ChoreoStatus
}

func (r *Status) Changed(newChoreoCtx *choreopb.ChoreoContext) bool {
	r.m.RLock()
	defer r.m.RUnlock()

	oldChoreoCtx := r.ChoreoCtx
	if oldChoreoCtx == nil {
		return true
	}
	if oldChoreoCtx.Branch != newChoreoCtx.Branch ||
		oldChoreoCtx.Url != newChoreoCtx.Url ||
		oldChoreoCtx.Directory != newChoreoCtx.Directory ||
		oldChoreoCtx.Ref != newChoreoCtx.Ref {
		return true
	}
	return false
}

func (r *Status) Set(s ChoreoStatus) {
	r.m.Lock()
	defer r.m.Unlock()
	r.ChoreoStatus = s
}

func Initializing() ChoreoStatus {
	return ChoreoStatus{
		Status:             false,
		Reason:             "initializing",
		RootChoreoInstance: nil,
		ChoreoCtx:          nil,
	}
}

func Success(c ChoreoInstance, choreoCtx *choreopb.ChoreoContext) ChoreoStatus {
	fmt.Println("success", choreoCtx)
	return ChoreoStatus{
		Status:             true,
		Reason:             "",
		RootChoreoInstance: c,
		ChoreoCtx:          choreoCtx,
	}
}

func Failed(msg string) ChoreoStatus {
	return ChoreoStatus{
		Status:             true,
		Reason:             "",
		RootChoreoInstance: nil,
		ChoreoCtx:          nil,
	}
}
