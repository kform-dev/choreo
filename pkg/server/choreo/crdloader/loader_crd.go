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

package crdloader

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"github.com/kform-dev/choreo/pkg/server/api"
	"github.com/kform-dev/choreo/pkg/server/apiserver/registry"
	"github.com/kform-dev/choreo/pkg/server/apiserver/rest"
	"github.com/kuidio/kuid/pkg/registry/options"
	apiextensionsinternal "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	structuraldefaulting "k8s.io/apiextensions-apiserver/pkg/apiserver/schema/defaulting"
	apiservervalidation "k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apiextensions-apiserver/pkg/controller/openapi/builder"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/managedfields"
	"k8s.io/kube-openapi/pkg/spec3"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// loadCRD loads the storage version of the CRD. if no storage version is supplied this call will fail
func LoadCRD(ctx context.Context, pathInRepo, dbpath string, crd *apiextensionsv1.CustomResourceDefinition, internalAPIs map[string]*BackendConfig, choreoAPI bool) (*api.ResourceContext, error) {
	if internalAPIs == nil {
		internalAPIs = map[string]*BackendConfig{}
	}

	rctx := &api.ResourceContext{}
	var errm error
	for _, v := range crd.Spec.Versions {
		internal := true
		//version := runtime.APIVersionInternal
		version := v.Name
		_, _, _, ok := parseKubeVersion(version)
		if ok {
			internal = false
		}

		//fmt.Println("version", ok, version, internal, v.Name, majorVersion, minorVersion, vType)
		gv := schema.GroupVersion{Group: crd.Spec.Group, Version: version}
		gvk := gv.WithKind(crd.Spec.Names.Kind)
		// strange the LIST is empty when unmarshaling
		gvklist := gv.WithKind(crd.Spec.Names.Kind + "List")
		//gvklist := gv.WithKind(crd.Spec.Names.ListKind)
		// we only use the storage version indicated in the crd to perform the validation
		if v.Storage {
			if rctx.Storage != nil {
				errm = errors.Join(errm, fmt.Errorf("invalid crd %s, a crd can only have 1 storage, got storage version %s", crd.Name, v.Name))
				continue
			}

			val := v.Schema

			var structuralSchema *structuralschema.Structural
			var internalSchemaProps *apiextensionsinternal.JSONSchemaProps
			internalValidationSchema := &apiextensionsinternal.CustomResourceValidation{}
			if val != nil {
				if err := apiextensionsv1.Convert_v1_CustomResourceValidation_To_apiextensions_CustomResourceValidation(val, internalValidationSchema, nil); err != nil {
					errm = errors.Join(errm, fmt.Errorf("invalid crd %s failed converting validation to internal %v", crd.Name, err))
					continue
				}
				s, err := structuralschema.NewStructural(internalValidationSchema.OpenAPIV3Schema)
				if !crd.Spec.PreserveUnknownFields && err != nil {
					errm = errors.Join(errm, fmt.Errorf("invalid crd %s, failed converting schema to structural", crd.Name))
					continue
				}
				if !crd.Spec.PreserveUnknownFields {
					if err := structuraldefaulting.PruneDefaults(s); err != nil {
						errm = errors.Join(errm, fmt.Errorf("invalid crd %s, failed to prune defaults", crd.Name))
						continue
					}
				}
				structuralSchema = s
				internalSchemaProps = internalValidationSchema.OpenAPIV3Schema
			}

			staticOpenAPISpec := make(map[string]*spec.Schema)
			openAPIModels, err := buildOpenAPIModelsForApply(staticOpenAPISpec, crd)
			if err != nil {
				errm = errors.Join(errm, fmt.Errorf("invalid crd %s, failed building openapi models: %v", crd.Name, err))
				continue
			}

			var typeConverter managedfields.TypeConverter = managedfields.NewDeducedTypeConverter()
			if len(openAPIModels) > 0 {
				typeConverter, err = managedfields.NewTypeConverter(openAPIModels, crd.Spec.PreserveUnknownFields)
				if err != nil {
					errm = errors.Join(errm, fmt.Errorf("invalid crd %s, failed getting type convertor: %v", crd.Name, err))
					continue
				}
			}

			schemaValidator, _, err := apiservervalidation.NewSchemaValidator(internalSchemaProps)
			if err != nil {
				errm = errors.Join(errm, fmt.Errorf("invalid crd %s, failed getting schema validator: %v", crd.Name, err))
				continue
			}

			var statusSpec *apiextensionsinternal.CustomResourceSubresourceStatus
			var statusValidator apiservervalidation.SchemaValidator
			if v.Subresources != nil && v.Subresources.Status != nil {
				statusSpec = &apiextensionsinternal.CustomResourceSubresourceStatus{}
				if err := apiextensionsv1.Convert_v1_CustomResourceSubresourceStatus_To_apiextensions_CustomResourceSubresourceStatus(v.Subresources.Status, statusSpec, nil); err != nil {
					errm = errors.Join(errm, fmt.Errorf("invalid crd %s, failed converting CRD status subresource to internal version: %v", crd.Name, err))
					continue
				}
				if internalValidationSchema.OpenAPIV3Schema != nil && internalValidationSchema.OpenAPIV3Schema.Properties != nil {
					if statusSchema, ok := internalValidationSchema.OpenAPIV3Schema.Properties["status"]; ok {
						statusValidator, _, err = apiservervalidation.NewSchemaValidator(&statusSchema)
						if err != nil {
							errm = errors.Join(errm, fmt.Errorf("invalid crd %s, failed getting schema validator: %v", crd.Name, err))
							continue
						}
					}
				}
			}

			defaulter := unstructuredDefaulter{structuralSchema: structuralSchema}
			creator := unstructuredCreator{}

			fieldManager, err := managedfields.NewDefaultCRDFieldManager(
				typeConverter,
				&safeConverterWrapper{
					unsafe: &crConverter{
						converter:     &nopConverter{},
						clusterScoped: false,
					},
				}, // we dont support conversion right now
				defaulter,
				creator,
				gvk,
				gv,
				"",  //subresource
				nil, // resetfields
			)
			if err != nil {
				return nil, err
			}

			var strategy rest.Strategy
			var invoker options.BackendInvoker
			if backendConfig, exists := internalAPIs[crd.Spec.Group]; exists {
				if crd.Spec.Names.Kind == backendConfig.ClaimKind {
					invoker = backendConfig.ClaimInvoker
				}
				if crd.Spec.Names.Kind == backendConfig.IndexKind {
					invoker = backendConfig.IndexInvoker
				}
			}
			if invoker != nil {
				strategy = registry.NewStrategy(
					crd.Spec.Scope == apiextensionsv1.NamespaceScoped, // namespaced
					gvk,
					schemaValidator,
					statusValidator,
					structuralSchema,
					defaulter,
					invoker, // this is to handle the internal api resources using synchronous calls
				)
			} else {
				strategy = registry.NewStrategy(
					crd.Spec.Scope == apiextensionsv1.NamespaceScoped, // namespaced
					gvk,
					schemaValidator,
					statusValidator,
					structuralSchema,
					defaulter,
					nil,
				)
			}

			//printStructuralSchema(structuralSchema)
			storage := registry.NewStorage(
				ctx,
				pathInRepo,
				dbpath,
				schema.GroupResource{Group: crd.Spec.Group, Resource: crd.Spec.Names.Plural},
				func() runtime.Unstructured {
					// set the expected group/version/kind in the new object as a signal to the versioning decoder
					u := &unstructured.Unstructured{}
					u.SetGroupVersionKind(gvk)
					return u
				},
				func() runtime.Unstructured {
					// lists are never stored, only manufactured, so stomp in the right kind
					u := &unstructured.UnstructuredList{}
					u.SetGroupVersionKind(gvklist)
					return u
				},
				strategy,
				fieldManager,
			)

			rctx.Storage = storage
			apiresource := &discoverypb.APIResource{
				Group:      gv.Group,
				Version:    gv.Version,
				Kind:       crd.Spec.Names.Kind,
				ListKind:   crd.Spec.Names.ListKind,
				Resource:   crd.Spec.Names.Plural,
				Namespaced: crd.Spec.Scope == apiextensionsv1.NamespaceScoped,
				Categories: crd.Spec.Names.Categories,
				ChoreoAPI:  choreoAPI,
			}
			if internal {
				rctx.Internal = apiresource
			} else {
				rctx.External = apiresource
			}
		} else {
			// no storage version
			if internal {
				errm = errors.Join(errm, fmt.Errorf("invalid crd %s, an internal version %s need to be a storage version", crd.Name, v.Name))
				continue
			}
			if rctx.External != nil {
				errm = errors.Join(errm, fmt.Errorf("invalid crd %s, a crd can only have 1 external version, got %s", crd.Name, v.Name))
				continue
			}
			rctx.External = &discoverypb.APIResource{
				Group:      gv.Group,
				Version:    gv.Version,
				Kind:       crd.Spec.Names.Kind,
				ListKind:   crd.Spec.Names.ListKind,
				Resource:   crd.Spec.Names.Plural,
				Namespaced: crd.Spec.Scope == apiextensionsv1.NamespaceScoped,
				Categories: crd.Spec.Names.Categories,
				ChoreoAPI:  choreoAPI,
			}
		}
	}
	if rctx.Storage == nil {
		errm = errors.Join(errm, fmt.Errorf("invalid crd %s, no storage version", crd.Name))
	}
	if rctx.External == nil {
		errm = errors.Join(errm, fmt.Errorf("invalid crd %s, no external version", crd.Name))
	}
	return rctx, errm

}

// buildOpenAPIModelsForApply constructs openapi models from any validation schemas specified in the custom resource,
// and merges it with the models defined in the static OpenAPI spec.
// Returns nil models ifthe static spec is nil, or an error is encountered.
func buildOpenAPIModelsForApply(staticOpenAPISpec map[string]*spec.Schema, crd *apiextensionsv1.CustomResourceDefinition) (map[string]*spec.Schema, error) {
	if staticOpenAPISpec == nil {
		return nil, nil
	}

	// Convert static spec to V3 format to be able to merge
	staticSpecV3 := &spec3.OpenAPI{
		Version: "3.0.0",
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title:   "Kubernetes CRD Swagger",
				Version: "v0.1.0",
			},
		},
		Components: &spec3.Components{
			Schemas: staticOpenAPISpec,
		},
	}

	specs := []*spec3.OpenAPI{staticSpecV3}
	for _, v := range crd.Spec.Versions {
		// Defaults are not pruned here, but before being served.
		// See flag description in builder.go for flag usage
		s, err := builder.BuildOpenAPIV3(crd, v.Name, builder.Options{})
		if err != nil {
			return nil, err
		}
		specs = append(specs, s)
	}

	mergedOpenAPI, err := builder.MergeSpecsV3(specs...)
	if err != nil {
		return nil, err
	}
	return mergedOpenAPI.Components.Schemas, nil
}

type versionType int

const (
	// Bigger the version type number, higher priority it is
	versionTypeAlpha versionType = iota
	versionTypeBeta
	versionTypeGA
)

var kubeVersionRegex = regexp.MustCompile("^v([\\d]+)(?:(alpha|beta)([\\d]+))?$")

func parseKubeVersion(v string) (majorVersion int, vType versionType, minorVersion int, ok bool) {
	var err error
	submatches := kubeVersionRegex.FindStringSubmatch(v)
	if len(submatches) != 4 {
		return 0, 0, 0, false
	}
	switch submatches[2] {
	case "alpha":
		vType = versionTypeAlpha
	case "beta":
		vType = versionTypeBeta
	case "":
		vType = versionTypeGA
	default:
		return 0, 0, 0, false
	}
	if majorVersion, err = strconv.Atoi(submatches[1]); err != nil {
		return 0, 0, 0, false
	}
	if vType != versionTypeGA {
		if minorVersion, err = strconv.Atoi(submatches[3]); err != nil {
			return 0, 0, 0, false
		}
	}
	return majorVersion, vType, minorVersion, true
}
