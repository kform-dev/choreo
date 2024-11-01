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
	"fmt"
	"reflect"

	"github.com/kform-dev/choreo/pkg/client/go/resourcemapper"
	"k8s.io/apimachinery/pkg/runtime"
)

// Mapper is a convenience struct for holding references to the interfaces
// needed to create Info for arbitrary objects.
type mapper struct {
	// localFn indicates the call can't make server requests
	//localFn func() bool // TBD if we need these

	resourceMapper resourcemapper.Mapper
	//resourceClient resourceclient.Client // TBD if we need these

	decoder runtime.Decoder
}

// InfoForData creates an Info object for the given data. An error is returned
// if any of the decoding or client lookup steps fail. Name and namespace will be
// set into Info if the mapping's MetadataAccessor can retrieve them.
func (r *mapper) infoForData(data []byte, source string) (*Info, error) {
	obj, _, err := r.decoder.Decode(data, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to decode %q: %v", source, err)
	}

	//fmt.Println("mapper infoForData gvk", gvk.String())

	name, _ := metadataAccessor.Name(obj)
	namespace, _ := metadataAccessor.Namespace(obj)

	return &Info{
		Source:    source,
		Namespace: namespace,
		Name:      name,

		Object: obj,
	}, nil
}

// InfoForObject creates an Info object for the given Object. An error is returned
// if the object cannot be introspected. Name and namespace will be set into Info
// if the mapping's MetadataAccessor can retrieve them.
func (m *mapper) infoForObject(obj runtime.Object, typer runtime.ObjectTyper) (*Info, error) {
	_, _, err := typer.ObjectKinds(obj)
	if err != nil {
		return nil, fmt.Errorf("unable to get type info from the object %q: %v", reflect.TypeOf(obj), err)
	}

	name, _ := metadataAccessor.Name(obj)
	namespace, _ := metadataAccessor.Namespace(obj)
	ret := &Info{
		Namespace: namespace,
		Name:      name,

		Object: obj,
	}

	return ret, nil
}
