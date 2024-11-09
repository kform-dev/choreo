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

package loader

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/henderiw/store"
	"github.com/kform-dev/choreo/pkg/cli/genericclioptions"
	"github.com/kform-dev/choreo/pkg/server/choreo/instance"
	"github.com/kform-dev/kform/pkg/fsys"
	invv1alpha1 "github.com/sdcio/config-server/apis/inv/v1alpha1"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	syaml "sigs.k8s.io/yaml"
)

// UpstreamLoader
type SchemaLoader struct {
	Parent instance.ChoreoInstance
	Cfg    *genericclioptions.ChoreoConfig
	//Client     resourceclient.Client
	//Branch     string
	RepoPath   string
	PathInRepo string
	TempDir    string
}

func (r *SchemaLoader) Load(ctx context.Context) error {
	gvks := []schema.GroupVersionKind{
		invv1alpha1.SchemeGroupVersion.WithKind(invv1alpha1.SchemaKind),
	}
	abspath := filepath.Join(r.RepoPath, r.PathInRepo, *r.Cfg.ServerFlags.SchemaPath)

	if !fsys.PathExists(abspath) {
		return nil
	}
	reader := GetFSYAMLReader(abspath, gvks)
	datastore, err := reader.Read(ctx)
	if err != nil {
		return err
	}

	var errs error
	datastore.List(func(k store.Key, rn *yaml.RNode) {
		schema := &invv1alpha1.Schema{}
		if err := syaml.Unmarshal([]byte(rn.MustString()), schema); err != nil {
			errs = errors.Join(errs, fmt.Errorf("invalid schema %s, err: %v", k.Name, err))
			return
		}
		fmt.Println("schemaloader schema", schema.Name)

		// upload the schema to the schemastore
		_, err = r.Parent.SchemaStore().GetSchemaDetails(ctx, &sdcpb.GetSchemaDetailsRequest{
			Schema: &sdcpb.Schema{
				Vendor:  schema.Spec.Provider,
				Version: schema.Spec.Version,
			},
		})
		if err != nil {
			_, err = r.Parent.SchemaLoader().LoadSchema(ctx, schema)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("cannot load schema: %v", err))
			}
		}
	})
	return errs
}
