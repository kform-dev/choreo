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
	"fmt"

	"github.com/kform-dev/choreo/pkg/util/object"
	apiextensionsvalidation "k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	v1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type APIPrepator interface {
	PrepareForCreate(ctx context.Context, obj runtime.Unstructured) error
	PrepareForUpdate(ctx context.Context, obj, old runtime.Unstructured) error
	PrepareForDelete(ctx context.Context, obj runtime.Unstructured) error
}

type apiValidator struct {
	namespaceScoped       bool
	gvk                   schema.GroupVersionKind
	schemaValidator       apiextensionsvalidation.SchemaValidator
	statusSchemaValidator apiextensionsvalidation.SchemaValidator
}

func (r apiValidator) ValidateCreate(ctx context.Context, obj *unstructured.Unstructured) field.ErrorList {
	// check apiVersion and kind according the api
	if errs := r.ValidateTypeMeta(ctx, obj); len(errs) > 0 {
		return errs
	}
	var allErrs field.ErrorList
	allErrs = append(allErrs, validation.ValidateObjectMetaAccessor(obj, r.namespaceScoped, validation.NameIsDNSSubdomain, field.NewPath("metadata"))...)
	allErrs = append(allErrs, apiextensionsvalidation.ValidateCustomResource(nil, obj.UnstructuredContent(), r.schemaValidator)...)

	return allErrs
}

func (r apiValidator) ValidateUpdate(ctx context.Context, new, old *unstructured.Unstructured, options ...apiextensionsvalidation.ValidationOption) field.ErrorList {
	// check apiVersion and kind according the api
	if errs := r.ValidateTypeMeta(ctx, new); len(errs) > 0 {
		return errs
	}
	newGeneration := object.GetGeneration(new)
	oldGeneration := object.GetGeneration(old)

	var allErrs field.ErrorList
	allErrs = append(allErrs, ValidateObjectMetaAccessorUpdate(new, old, newGeneration, oldGeneration, field.NewPath("metadata"))...)
	allErrs = append(allErrs, apiextensionsvalidation.ValidateCustomResourceUpdate(nil, new.UnstructuredContent(), old.UnstructuredContent(), r.schemaValidator, options...)...)
	if status, hasStatus := new.UnstructuredContent()["status"]; hasStatus {
		allErrs = append(allErrs, apiextensionsvalidation.ValidateCustomResourceUpdate(nil, status, old.UnstructuredContent()["status"], r.statusSchemaValidator)...)
	}
	return allErrs
}

func (r apiValidator) ValidateTypeMeta(ctx context.Context, obj *unstructured.Unstructured) field.ErrorList {
	typeAccessor, err := meta.TypeAccessor(obj)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("kind"), nil, err.Error())}
	}

	var allErrs field.ErrorList
	if typeAccessor.GetKind() != r.gvk.Kind {
		allErrs = append(allErrs, field.Invalid(field.NewPath("kind"), typeAccessor.GetKind(), fmt.Sprintf("must be %v", r.gvk.Kind)))
	}
	if typeAccessor.GetAPIVersion() != r.gvk.GroupVersion().String() {
		allErrs = append(allErrs, field.Invalid(field.NewPath("apiVersion"), typeAccessor.GetAPIVersion(), fmt.Sprintf("must be %v", r.gvk.GroupVersion().String())))
	}
	return allErrs
}

// ValidateObjectMetaAccessorUpdate validates an object's metadata when updated.
func ValidateObjectMetaAccessorUpdate(newMeta, oldMeta metav1.Object, newGeneration, oldGeneration int64, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// Finalizers cannot be added if the object is already being deleted.
	if oldMeta.GetDeletionTimestamp() != nil {
		//fmt.Println("delete timestamp", oldMeta.GetDeletionTimestamp())
		allErrs = append(allErrs, validation.ValidateNoNewFinalizers(newMeta.GetFinalizers(), oldMeta.GetFinalizers(), fldPath.Child("finalizers"))...)
	}

	// Reject updates that don't specify a resource version
	if len(newMeta.GetResourceVersion()) == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("resourceVersion"), newMeta.GetResourceVersion(), "must be specified for an update"))
	}

	// Generation shouldn't be decremented
	if newGeneration < oldGeneration {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("generation"), newGeneration, "must not be decremented"))
	}

	allErrs = append(allErrs, validation.ValidateImmutableField(newMeta.GetName(), oldMeta.GetName(), fldPath.Child("name"))...)
	allErrs = append(allErrs, validation.ValidateImmutableField(newMeta.GetNamespace(), oldMeta.GetNamespace(), fldPath.Child("namespace"))...)
	allErrs = append(allErrs, validation.ValidateImmutableField(newMeta.GetUID(), oldMeta.GetUID(), fldPath.Child("uid"))...)
	allErrs = append(allErrs, validation.ValidateImmutableField(newMeta.GetCreationTimestamp(), oldMeta.GetCreationTimestamp(), fldPath.Child("creationTimestamp"))...)
	//allErrs = append(allErrs, validation.ValidateImmutableField(newMeta.GetDeletionTimestamp(), oldMeta.GetDeletionTimestamp(), fldPath.Child("deletionTimestamp"))...)
	allErrs = append(allErrs, validation.ValidateImmutableField(newMeta.GetDeletionGracePeriodSeconds(), oldMeta.GetDeletionGracePeriodSeconds(), fldPath.Child("deletionGracePeriodSeconds"))...)

	allErrs = append(allErrs, v1validation.ValidateLabels(newMeta.GetLabels(), fldPath.Child("labels"))...)
	allErrs = append(allErrs, validation.ValidateAnnotations(newMeta.GetAnnotations(), fldPath.Child("annotations"))...)
	allErrs = append(allErrs, validation.ValidateOwnerReferences(newMeta.GetOwnerReferences(), fldPath.Child("ownerReferences"))...)
	allErrs = append(allErrs, v1validation.ValidateManagedFields(newMeta.GetManagedFields(), fldPath.Child("managedFields"))...)

	return allErrs
}
