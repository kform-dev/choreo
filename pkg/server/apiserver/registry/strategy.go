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

package registry

import (
	"context"

	"github.com/kuidio/kuid/pkg/registry/options"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel/model"
	structurallisttype "k8s.io/apiextensions-apiserver/pkg/apiserver/schema/listtype"
	schemaobjectmeta "k8s.io/apiextensions-apiserver/pkg/apiserver/schema/objectmeta"
	apiextensionsvalidation "k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	celconfig "k8s.io/apiserver/pkg/apis/cel"
	"k8s.io/apiserver/pkg/cel/common"
)

func NewStrategy(
	namespaceScoped bool,
	gvk schema.GroupVersionKind,
	schemaValidator, statusSchemaValidator apiextensionsvalidation.SchemaValidator,
	structuralSchema *structuralschema.Structural,
	defaulter runtime.ObjectDefaulter,
	invoker options.BackendInvoker,
) strategy {
	return strategy{
		namespaceScoped: namespaceScoped,
		gvk:             gvk,
		defaulter:       defaulter,
		validator: apiValidator{
			namespaceScoped:       namespaceScoped,
			gvk:                   gvk,
			schemaValidator:       schemaValidator,
			statusSchemaValidator: statusSchemaValidator,
		},
		invoker:          invoker,
		structuralSchema: structuralSchema,
		celValidator:     cel.NewValidator(structuralSchema, true, celconfig.PerCallLimit),
	}
}

type strategy struct {
	namespaceScoped  bool
	gvk              schema.GroupVersionKind
	defaulter        runtime.ObjectDefaulter
	validator        apiValidator
	invoker          options.BackendInvoker
	structuralSchema *structuralschema.Structural
	celValidator     *cel.Validator
}

func (r strategy) InvokeCreate(ctx context.Context, obj runtime.Object, recursion bool) (runtime.Object, error) {
	if r.invoker != nil {
		return r.invoker.InvokeCreate(ctx, obj, recursion)
	}
	return obj, nil
}

func (r strategy) ValidateCreate(ctx context.Context, obj runtime.Unstructured) field.ErrorList {
	r.defaulter.Default(obj)

	uNew := &unstructured.Unstructured{Object: obj.UnstructuredContent()}

	var errs field.ErrorList
	errs = append(errs, r.validator.ValidateCreate(ctx, uNew)...)
	// validate embedded resources
	errs = append(errs, schemaobjectmeta.Validate(nil, uNew.Object, r.structuralSchema, false)...)
	// validate x-kubernetes-list-type "map" and "set" invariant
	errs = append(errs, structurallisttype.ValidateListSetsAndMaps(nil, r.structuralSchema, uNew.Object)...)
	// validate x-kubernetes-validations rules
	if celValidator := r.celValidator; celValidator != nil {
		if has, err := hasBlockingErr(errs); has {
			errs = append(errs, err)
		} else {
			err, _ := celValidator.Validate(ctx, nil, r.structuralSchema, uNew.Object, nil, celconfig.RuntimeCELCostBudget)
			errs = append(errs, err...)
		}
	}
	return errs
}

func (r strategy) InvokeUpdate(ctx context.Context, obj, old runtime.Object, recursion bool) (runtime.Object, runtime.Object, error) {
	if r.invoker != nil {
		return r.invoker.InvokeUpdate(ctx, obj, old, recursion)
	}
	return obj, old, nil
}

func (r strategy) ValidateUpdate(ctx context.Context, obj, old runtime.Unstructured) field.ErrorList {
	r.defaulter.Default(obj)

	uNew := &unstructured.Unstructured{Object: obj.UnstructuredContent()}
	uOld := &unstructured.Unstructured{Object: old.UnstructuredContent()}

	correlatedObject := common.NewCorrelatedObject(uNew.Object, uOld.Object, &model.Structural{Structural: r.structuralSchema})
	options := []apiextensionsvalidation.ValidationOption{
		apiextensionsvalidation.WithRatcheting(correlatedObject),
	}
	celOptions := []cel.Option{
		cel.WithRatcheting(correlatedObject),
	}

	var errs field.ErrorList
	errs = append(errs, r.validator.ValidateUpdate(ctx, uNew, uOld, options...)...)
	// Checks the embedded objects. We don't make a difference between update and create for those.
	errs = append(errs, schemaobjectmeta.Validate(nil, uNew.Object, r.structuralSchema, false)...)

	// ratcheting validation of x-kubernetes-list-type value map and set
	if oldErrs := structurallisttype.ValidateListSetsAndMaps(nil, r.structuralSchema, uOld.Object); len(oldErrs) == 0 {
		errs = append(errs, structurallisttype.ValidateListSetsAndMaps(nil, r.structuralSchema, uNew.Object)...)
	}

	// validate x-kubernetes-validations rules
	if celValidator := r.celValidator; celValidator != nil {
		if has, err := hasBlockingErr(errs); has {
			errs = append(errs, err)
		} else {
			err, _ := celValidator.Validate(ctx, nil, r.structuralSchema, uNew.Object, uOld.Object, celconfig.RuntimeCELCostBudget, celOptions...)
			errs = append(errs, err...)
		}
	}
	return errs
}

func (r strategy) InvokeDelete(ctx context.Context, obj runtime.Object, recursion bool) (runtime.Object, error) {
	if r.invoker != nil {
		return r.invoker.InvokeDelete(ctx, obj, recursion)
	}
	return obj, nil
}

// OpenAPIv3 type/maxLength/maxItems/MaxProperties/required/enum violation/wrong type field validation failures are viewed as blocking err for CEL validation
func hasBlockingErr(errs field.ErrorList) (bool, *field.Error) {
	for _, err := range errs {
		if err.Type == field.ErrorTypeNotSupported || err.Type == field.ErrorTypeRequired || err.Type == field.ErrorTypeTooLong || err.Type == field.ErrorTypeTooMany || err.Type == field.ErrorTypeTypeInvalid {
			return true, field.Invalid(nil, nil, "some validation rules were not checked because the object was invalid; correct the existing errors to complete validation")
		}
	}
	return false, nil
}
