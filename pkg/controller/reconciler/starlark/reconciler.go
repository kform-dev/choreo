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

package starlark

import (
	"context"
	"fmt"
	"time"

	"github.com/henderiw/iputil"
	"github.com/henderiw/logger/log"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/controller/reconcile"
	"github.com/kform-dev/choreo/pkg/controller/reconciler/resources"
	"github.com/kform-dev/choreo/pkg/controller/reconciler/starlark/util"
	"github.com/kform-dev/choreo/pkg/proto/grpcerrors"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/util/object"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/structured-merge-diff/v4/value"
)

func NewReconcilerFn(client resourceclient.Client, reconcileConfig *choreov1alpha1.Reconciler, libs *unstructured.UnstructuredList, branch string) reconcile.TypedReconcilerFn {
	return func() (reconcile.TypedReconciler, error) {
		r := &reconciler{
			name:   reconcileConfig.Name,
			client: client,
			forgvk: reconcileConfig.GetForGVK(),
			owns:   reconcileConfig.GetOwnsGVKs(),
			branch: branch,
		}

		// Setup built-in functions
		builtins := starlark.StringDict{
			"client_get":           starlark.NewBuiltin("client_get", r.get),
			"client_list":          starlark.NewBuiltin("client_list", r.list),
			"client_update":        starlark.NewBuiltin("client_update", r.update),
			"client_update_status": starlark.NewBuiltin("client_update_status", r.updateStatus),
			"client_create":        starlark.NewBuiltin("client_create", r.create),
			"client_apply":         starlark.NewBuiltin("client_apply", r.apply),
			"client_delete":        starlark.NewBuiltin("client_delete", r.delete),
			"reconcile_result":     starlark.NewBuiltin("reconcile_result", r.reconcileResult),
			"get_resource":         starlark.NewBuiltin("get_resource", getResource),
			"del_finalizer":        starlark.NewBuiltin("del_finalizer", delFinalizer),
			"add_finalizer":        starlark.NewBuiltin("add_finalizer", addFinalizer),
			"get_prefixlength":     starlark.NewBuiltin("get_prefixlength", getPrefixLength),
			"get_subnetname":       starlark.NewBuiltin("get_subnetname", getSubnetName),
			"get_address":          starlark.NewBuiltin("get_address", getAddress),
			"isIPv4":               starlark.NewBuiltin("isIPv4", isIPv4),
			"isIPv6":               starlark.NewBuiltin("isIPv6", isIPv6),
			"is_conditionready":    starlark.NewBuiltin("is_condition_ready", isConditionReady),
		}

		// cache deals with library loading
		cache := newCache(libs)
		cc := new(cycleChecker)

		thread := &starlark.Thread{
			Name: "main",
			Load: func(thread *starlark.Thread, module string) (starlark.StringDict, error) {
				return cache.get(cc, module, builtins)
			},
		}

		startlarkReconciler := reconcileConfig.Spec.Code["reconciler.star"]
		starlarkReconciler, err := starlark.ExecFileOptions(&syntax.FileOptions{}, thread, "reconciler.star", startlarkReconciler, builtins)
		if err != nil {
			return nil, fmt.Errorf("reconciler %s err: %v", reconcileConfig.GetName(), err)
		}

		r.startlarkReconciler = starlarkReconciler
		return r, nil
	}
}

type reconciler struct {
	name                string
	startlarkReconciler starlark.StringDict
	client              resourceclient.Client
	forgvk              schema.GroupVersionKind
	owns                sets.Set[schema.GroupVersionKind]
	branch              string
	// dynamic data set on each reconcile
	ctx       context.Context // bad practice but allows for reuse of ctx
	resources *resources.Resources
}

func (r *reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)
	// bad practice but allows to reuse the context
	r.ctx = ctx
	// get the resource
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(r.forgvk)
	if err := r.client.Get(r.ctx, req.NamespacedName, u, &resourceclient.GetOptions{
		ShowManagedFields: true,
		Origin:            r.name,
		Branch:            r.branch,
	}); err != nil {
		if !grpcerrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
		// stop the reconcile loop since the object dissapeared
		return reconcile.Result{}, nil
	}
	//u = u.DeepCopy()

	/*
		match := false
		if r.name == "infra.kuid.dev_node_if-si-ni" || r.name == "infra.kuid.dev_node_bgp" {
			match = true
		}
		if match {
			condition := object.GetCondition(u.Object, "IPClaimready")
			fmt.Println(r.name, req.Name, "reconciler", "ipclaimready condition before", condition["status"], condition["message"])
		}
	*/

	obj, err := util.UnstructuredToStarlarkValue(u)
	if err != nil {
		return reconcile.Result{}, err
	}

	// reinitialize the resource on each reconcile
	r.resources = resources.New(r.name, r.client, u, r.owns, r.branch)

	// call the python code; it will call various hooks we build
	// returns an error message
	reconciler := r.startlarkReconciler["reconcile"]
	thread := &starlark.Thread{Name: "main"}
	v, err := starlark.Call(thread, reconciler, starlark.Tuple{starlark.Value(obj)}, nil)
	if err != nil {
		// this is a starlark execution runtime failure
		return reconcile.Result{}, fmt.Errorf("starlark execution runtime failure: %s", err.Error())
	}

	result, err := convertReconcileResult(v)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("starlark execution cannot convert result: %s", err.Error())
	}

	if result.Fatal {
		return reconcile.Result{}, fmt.Errorf(result.Message)
	}
	var requeue time.Duration
	if result.RequeueAfter != 0 {
		requeue = time.Duration(result.RequeueAfter) * time.Second
	}
	if result.Message != "" {
		log.Debug("reconcile failed", "msg", result.Message)
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: requeue,
			Message:      result.Message,
		}, nil
	}

	return reconcile.Result{
		Requeue:      result.Requeue,
		RequeueAfter: requeue,
	}, nil
}

type ReconcileResult struct {
	Requeue      bool
	RequeueAfter int64
	Message      string
	Fatal        bool
}

func convertReconcileResult(v starlark.Value) (*ReconcileResult, error) {
	result, err := util.StarlarkValueToMap(v)
	if err != nil {
		return nil, err
	}

	reconcileResult := &ReconcileResult{}
	if v, ok := result["requeue"]; ok {
		vv, ok := v.(bool)
		if !ok {
			return nil, fmt.Errorf("reconcileResult requeue is not a bool, got %T", v)
		}
		reconcileResult.Requeue = vv

	}
	if v, ok := result["requeueAfter"]; ok {
		vv, ok := v.(int64)
		if !ok {
			return nil, fmt.Errorf("reconcileResult requeueAfter is not a int64, got %T", v)
		}
		reconcileResult.RequeueAfter = vv
	}
	if v, ok := result["fatal"]; ok {
		vv, ok := v.(bool)
		if !ok {
			return nil, fmt.Errorf("reconcileResult fatal is not a bool, got %T", v)
		}
		reconcileResult.Fatal = vv

	}
	if v, ok := result["error"]; ok {
		vv, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("reconcileResult error is not a string, got %T", v)
		}
		reconcileResult.Message = vv
	}
	return reconcileResult, nil
}

func reconcileResult(requeue bool, requeueAfter int64, err error, fatal bool) *starlark.Dict {
	// Prepare the result dict
	result := starlark.NewDict(0)
	if err != nil {
		result.SetKey(starlark.String("fatal"), starlark.Bool(fatal))
		result.SetKey(starlark.String("error"), starlark.String(err.Error()))
		return result
	}
	result.SetKey(starlark.String("fatal"), starlark.Bool(false))
	result.SetKey(starlark.String("error"), starlark.None)
	result.SetKey(starlark.String("requeueAfter"), starlark.MakeInt64(requeueAfter))
	result.SetKey(starlark.String("requeue"), starlark.Bool(requeue))
	return result
}

func result(resource starlark.Value, err error, fatal bool) *starlark.Dict {
	// Prepare the result dict
	result := starlark.NewDict(0)
	if err != nil {
		result.SetKey(starlark.String("fatal"), starlark.Bool(fatal))
		result.SetKey(starlark.String("error"), starlark.String(err.Error()))
		result.SetKey(starlark.String("resource"), starlark.None)
		return result
	}
	result.SetKey(starlark.String("fatal"), starlark.Bool(false))
	result.SetKey(starlark.String("error"), starlark.None)
	result.SetKey(starlark.String("resource"), resource)
	return result
}

func (r *reconciler) reconcileResult(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var obj starlark.Value
	var requeue starlark.Bool
	var requeueAfter starlark.Int
	var conditionType starlark.String
	var msg starlark.String
	var fatal starlark.Bool

	if err := starlark.UnpackArgs("reconcile_result", args, nil, "obj", &obj, "requeue", &requeue, "requeueAfter", &requeueAfter, "conditionType", &conditionType, "msg", &msg, "fatal", &fatal); err != nil {
		return reconcileResult(
				bool(requeue),
				util.StarlarkIntToInt64(requeueAfter),
				fmt.Errorf("error: %s, msg: %s", err.Error(), msg.GoString()),
				true),
			nil
	}

	u, err := util.StarlarkValueToUnstructured(obj)
	if err != nil {
		return reconcileResult(
				bool(requeue),
				util.StarlarkIntToInt64(requeueAfter),
				fmt.Errorf("error: %s, msg: %s", err.Error(), msg.GoString()),
				true),
			nil
	}

	// removes the fields that are not managed by this reconciler based on the managedFields info in the resource
	// done before conditions are set
	r.pruneUnmanagedFields(u)

	if conditionType != "" {
		/*
			uobj := &unstructured.Unstructured{Object: u.UnstructuredContent()}
			reqname := uobj.GetName()
			if r.name == "infra.kuid.dev_node_id" || r.name == "infra.kuid.dev_node_if-si-ni" || r.name == "infra.kuid.dev_node_bgp" {
				c := object.GetCondition(u.UnstructuredContent(), "IPClaimReady")
				fmt.Println(r.name, reqname, "reconcile", "ipclaimready condition after", fmt.Sprintf("%v", c["status"]), fmt.Sprintf("%v", c["message"]))
			}
		*/
		object.SetCondition(u.UnstructuredContent(), conditionType.GoString(), msg.GoString())
		/*
			if r.name == "infra.kuid.dev_node_id" || r.name == "infra.kuid.dev_node_if-si-ni" || r.name == "infra.kuid.dev_node_bgp" {
				c := object.GetCondition(u.UnstructuredContent(), "IPClaimReady")
				fmt.Println(r.name, reqname, "reconcile", "ipclaimready condition after", fmt.Sprintf("%v", c["status"]), fmt.Sprintf("%v", c["message"]))

			}
		*/
	}

	if err := r.client.Apply(r.ctx, u, &resourceclient.ApplyOptions{
		FieldManager: r.name,
		Origin:       r.name,
		Branch:       r.branch,
	}); err != nil {
		if grpcerrors.IsNotFound(err) {
			// for not found we dont return fatal -> other we do
			return reconcileResult(
					bool(requeue),
					util.StarlarkIntToInt64(requeueAfter),
					err,
					false),
				nil
		}
		return reconcileResult(
				bool(requeue),
				util.StarlarkIntToInt64(requeueAfter),
				err,
				true),
			nil
	}

	if msg.GoString() != "" {
		err := fmt.Errorf("reconcile failed %s", msg.GoString())
		return reconcileResult(
				bool(requeue),
				util.StarlarkIntToInt64(requeueAfter),
				err,
				bool(fatal)),
			nil
	}

	return reconcileResult(
			bool(requeue),
			util.StarlarkIntToInt64(requeueAfter),
			nil,
			false),
		nil
}

func (r *reconciler) pruneUnmanagedFields(obj runtime.Unstructured) {
	// get the managed fields from the resource
	managedFields, found, err := unstructured.NestedSlice(obj.UnstructuredContent(), "metadata", "managedFields")
	if err != nil || !found {
		return
	}

	// 1. before-first-apply: fields set before the SSA was performed
	// since we always use SSA this should always result in NO fields in the map
	// we cannot prune these fields if they were present
	// 2. any field not managed by this reconciler aka fieldmanager should be pruned
	nonManagedFields := map[string]any{}
	var beforeFirstApplyFields map[string]any
	for _, field := range managedFields {
		fieldMap, ok := field.(map[string]interface{})
		if !ok {
			continue
		}
		if fieldMap["manager"] != r.name && fieldMap["operation"] == "Apply" {
			fieldsV1, ok := fieldMap["fieldsV1"].(map[string]interface{})
			if ok {
				mergeFieldMaps(nonManagedFields, fieldsV1)
			}
		}
		if fieldMap["manager"] == "before-first-apply" && fieldMap["operation"] == "Update" {
			beforeFirstApplyFields = fieldMap["fieldsV1"].(map[string]interface{})
		}
	}
	// fields managed using before apply should not be pruned
	pruneFieldsUsingBeforeApply(nonManagedFields, beforeFirstApplyFields)
	// debug only node object for now
	debug := ""
	/*
		if obj.GetObjectKind().GroupVersionKind().Kind == "Link" {
			debug = "link"
			accessor, err := meta.Accessor(obj)
			if err == nil {
				debug = fmt.Sprintf("%s %s %s", r.name, obj.GetObjectKind().GroupVersionKind().Kind, accessor.GetName())
				fmt.Println(debug, "nonManagedFields", nonManagedFields)
				fmt.Println(debug, "beforeFirstApplyFields", beforeFirstApplyFields)
			}
		}
	*/

	data := obj.UnstructuredContent()
	// prune fields from unstructured data
	pruneFieldsFromUnstructured(debug, data, nonManagedFields)
	obj.SetUnstructuredContent(data)
	// SSA does not like to see managed fields in the object
	removeManagedFieldsFromUnstructured(obj)
	// resource version and generation are picked up from the stored resource
	// -> prune data from the resource
	removeResourceVersionAndGenerationFromUnstructured(obj)
	// debug only node object for now
	/*
		if obj.GetObjectKind().GroupVersionKind().Kind == "Link" {
			accessor, err := meta.Accessor(obj)
			if err == nil {
				fmt.Println(r.name, "link", accessor.GetName(), "object\n", obj.UnstructuredContent())
			}
		}
	*/
}

func pruneFieldsFromUnstructured(debug string, obj map[string]interface{}, nonManagedFields map[string]interface{}) {
	for nonManagedFieldKey, nonManagedFieldValue := range nonManagedFields {
		// Remove the "f:" prefix to match against the actual fields in obj
		realKey := nonManagedFieldKey[2:]
		if nonManagedFieldMap, ok := nonManagedFieldValue.(map[string]interface{}); ok {
			if objMap, exists := obj[realKey]; exists {
				if debug != "" {
					fmt.Println(debug, "pruneFieldsFromUnstructured, map exists realKey", realKey, nonManagedFieldValue, objMap, nonManagedFieldMap)
				}
				switch objMap := objMap.(type) {
				case map[string]interface{}:
					// this takes care of the fact that a map contains no element in the nonManagedFieldMap
					// as such we can delete the entry w/o going through the list
					if len(nonManagedFieldMap) == 0 {
						delete(obj, realKey)
					} else {
						pruneFieldsFromUnstructured(debug, objMap, nonManagedFieldMap)
						// when the map is empty is pruned we delete the complete element from the map
						if len(objMap) == 0 {
							delete(obj, realKey)
						}
					}
				case []interface{}:
					// this takes care of the fact that a map contains no element in the nonManagedFieldMap
					// as such we can delete the entry w/o going through the list
					if len(nonManagedFieldMap) == 0 {
						delete(obj, realKey)
					} else {
						obj[realKey] = pruneList(objMap, nonManagedFieldMap)
						// when the last entry is pruned we delete the complete element from the map
						if len(obj[realKey].([]interface{})) == 0 {
							delete(obj, realKey)
						}
					}
				default:
					// the objMap represents the real value
					delete(obj, realKey)
				}
			}
		} else {
			delete(obj, realKey)
		}
	}
}

func pruneList(objList []interface{}, nonManagedFields map[string]interface{}) []interface{} {
	for i := len(objList) - 1; i >= 0; i-- {
		item := objList[i]
		// Handle "k:" for list items
		for nonManagedFieldKey := range nonManagedFields {
			pe, err := NewPathElement(nonManagedFieldKey)
			if err != nil {
				continue
			}
			switch {
			case pe.Key != nil:
				// Handle "k:" prefix for list items (keys)
				objMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				matches := true
				for _, v := range *pe.Key {
					if objVal, exists := objMap[v.Name]; !exists || fmt.Sprintf("%v", objVal) != v.Value.AsString() {
						matches = false
						break
					}
				}
				if matches {
					objList = append(objList[:i], objList[i+1:]...)
				}
			case pe.Value != nil:
				// Handle "v:" prefix for list items (values)
				if fmt.Sprintf("%q", item) == value.ToString(*pe.Value) {
					// Remove the item if it matches the managed value
					objList = append(objList[:i], objList[i+1:]...)
				}
			}
		}
	}
	return objList
}

// pruneFieldsUsingBeforeApply prunes fields in nonManagedFields that were set using beforeApplyFields
// beforeApplyFields are fields set before the first SSA was done, so these are owned by a special manager
// (before-first-apply)
func pruneFieldsUsingBeforeApply(nonManagedFields, beforeApplyFields map[string]interface{}) {
	for beforeKey, beforeValue := range beforeApplyFields {
		if value, found := nonManagedFields[beforeKey]; found {
			// If it's a nested map, recursively prune
			if nestedMap, ok := value.(map[string]interface{}); ok {
				pruneFieldsUsingBeforeApply(nestedMap, beforeValue.(map[string]interface{}))
				if len(nestedMap) == 0 {
					delete(nonManagedFields, beforeKey)
				}
			}
		}
	}
}

// mergeFieldMaps merges several managed fieldsmaps together
func mergeFieldMaps(dest, src map[string]interface{}) {
	for key, value := range src {
		if existingValue, exists := dest[key]; exists {
			// If the key exists and both are maps, merge recursively
			if existingMap, ok := existingValue.(map[string]interface{}); ok {
				if valueMap, ok := value.(map[string]interface{}); ok {
					mergeFieldMaps(existingMap, valueMap)
					continue
				}
			}
		}
		// Otherwise, just set the value
		dest[key] = value
	}
}

func removeManagedFieldsFromUnstructured(obj runtime.Unstructured) {
	log := log.FromContext(context.Background())
	// Access the unstructured content
	unstructuredContent := obj.UnstructuredContent()

	// Access the metadata section
	metadata, found, err := unstructured.NestedMap(unstructuredContent, "metadata")
	if err != nil || !found {
		return
	}

	// Remove the managedFields key from metadata
	delete(metadata, "managedFields")

	// Set the updated metadata back to the unstructured content
	err = unstructured.SetNestedMap(unstructuredContent, metadata, "metadata")
	if err != nil {
		log.Error("error setting metadata", "error", err)
	}
}

func removeResourceVersionAndGenerationFromUnstructured(obj runtime.Unstructured) {
	log := log.FromContext(context.Background())
	// Access the unstructured content
	unstructuredContent := obj.UnstructuredContent()

	// Access the metadata section
	metadata, found, err := unstructured.NestedMap(unstructuredContent, "metadata")
	if err != nil || !found {
		return
	}

	// Remove the resourceVersion key from metadata
	delete(metadata, "resourceVersion")
	delete(metadata, "generation")

	// Set the updated metadata back to the unstructured content
	err = unstructured.SetNestedMap(unstructuredContent, metadata, "metadata")
	if err != nil {
		log.Error("error setting metadata", "error", err)
	}
}

func (r *reconciler) get(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name, namespace starlark.String
	var obj starlark.Value

	if err := starlark.UnpackArgs("client_get", args, nil, "name", &name, "namespace", &namespace, "obj", &obj); err != nil {
		return result(nil, fmt.Errorf("reconciler %s client_get error: %s", r.name, err.Error()), true), fmt.Errorf("reconciler %s client_get error: %s", r.name, err.Error())
	}

	u, err := util.StarlarkValueToUnstructured(obj)
	if err != nil {
		return result(nil, err, true), nil
	}

	if err := r.client.Get(r.ctx,
		types.NamespacedName{Namespace: namespace.GoString(), Name: name.GoString()},
		u,
		&resourceclient.GetOptions{
			Origin: r.name,
			Branch: r.branch,
		},
	); err != nil {
		return result(nil, err, false), nil
	}
	v, err := util.UnstructuredToStarlarkValue(u)
	if err != nil {
		return result(nil, err, true), nil
	}
	return result(v, nil, false), nil
}

func (r *reconciler) list(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var obj, fieldSelector starlark.Value

	if err := starlark.UnpackArgs("client_list", args, nil, "obj", &obj, "fieldSelector", &fieldSelector); err != nil {
		return result(nil, fmt.Errorf("reconciler %s client_list error: %s", r.name, err.Error()), true), fmt.Errorf("reconciler %s client_list error: %s", r.name, err.Error())
	}

	u, err := util.StarlarkValueToUnstructured(obj)
	if err != nil {
		return result(nil, err, true), nil
	}

	match, err := util.StarlarkValueToGoMap(fieldSelector)
	if err != nil {
		return result(nil, err, true), nil
	}

	if err := r.client.List(r.ctx,
		u,
		&resourceclient.ListOptions{
			ExprSelector: &resourcepb.ExpressionSelector{
				Match: match,
			},
			Origin: r.name,
			Branch: r.branch,
		},
	); err != nil {
		return result(nil, err, false), nil
	}

	v, err := util.UnstructuredToStarlarkValue(u)
	if err != nil {
		return result(nil, err, true), nil
	}
	return result(v, nil, false), nil
}

func (r *reconciler) update(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var obj starlark.Value

	if err := starlark.UnpackArgs("client_update", args, nil, "obj", &obj); err != nil {
		return result(nil, fmt.Errorf("reconciler %s client_update error: %s", r.name, err.Error()), true), fmt.Errorf("reconciler %s client_update error: %s", r.name, err.Error())
	}

	u, err := util.StarlarkValueToUnstructured(obj)
	if err != nil {
		return result(nil, err, true), nil
	}

	if err := r.client.Apply(r.ctx,
		u,
		&resourceclient.ApplyOptions{
			FieldManager: r.name,
			Origin:       r.name,
			Branch:       r.branch,
		},
	); err != nil {
		if grpcerrors.InvalidArgument(err) {
			// this meaans the content on the resource we try to update is badly formatted
			return result(nil, err, true), nil
		}
		return result(nil, err, false), nil
	}

	v, err := util.UnstructuredToStarlarkValue(u)
	if err != nil {
		return result(nil, err, true), nil
	}
	return result(v, nil, false), nil
}

func (r *reconciler) updateStatus(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var obj starlark.Value

	if err := starlark.UnpackArgs("client_update_status", args, nil, "obj", &obj); err != nil {
		return result(nil, fmt.Errorf("reconciler %s client_update_status error: %s", r.name, err.Error()), true), fmt.Errorf("reconciler %s client_update_status error: %s", r.name, err.Error())
	}

	u, err := util.StarlarkValueToUnstructured(obj)
	if err != nil {
		return result(nil, err, true), nil
	}

	if err := r.client.Apply(r.ctx,
		u,
		&resourceclient.ApplyOptions{
			FieldManager: r.name,
			Origin:       r.name,
			Branch:       r.branch,
		},
	); err != nil {
		if grpcerrors.InvalidArgument(err) {
			// this meaans the content on the resource we try to update is badly formatted
			return result(nil, err, true), nil
		}
		return result(nil, err, false), nil
	}

	v, err := util.UnstructuredToStarlarkValue(u)
	if err != nil {
		return result(nil, err, true), nil
	}
	return result(v, nil, false), nil
}

func (r *reconciler) create(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var obj starlark.Value

	if err := starlark.UnpackArgs("client_create", args, nil, "obj", &obj); err != nil {
		return result(nil, fmt.Errorf("reconciler %s client_create error: %s", r.name, err.Error()), true), fmt.Errorf("reconciler %s client_create error: %s", r.name, err.Error())
	}

	u, err := util.StarlarkValueToUnstructured(obj)
	if err != nil {
		return result(nil, err, true), nil
	}

	r.resources.AddNewResource(r.ctx, &unstructured.Unstructured{
		Object: u.UnstructuredContent(),
	})

	v, err := util.UnstructuredToStarlarkValue(u)
	if err != nil {
		return result(nil, err, true), nil
	}
	return result(v, nil, false), nil
}

func (r *reconciler) apply(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("client_apply", args, nil); err != nil {
		return result(nil, fmt.Errorf("reconciler %s client_apply error: %s", r.name, err.Error()), true), fmt.Errorf("reconciler %s client_apply error: %s", r.name, err.Error())
	}
	if err := r.resources.Apply(r.ctx); err != nil {
		if grpcerrors.InvalidArgument(err) {
			// this meaans the content on the resource we try to update is badly formatted
			// we return a fatal error
			return result(nil, err, true), nil
		}
		return result(nil, err, false), nil
	}

	return result(nil, nil, false), nil
}

func (r *reconciler) delete(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("client_delete", args, nil); err != nil {
		return result(nil, fmt.Errorf("reconciler %s client_delete error: %s", r.name, err.Error()), true), fmt.Errorf("reconciler %s client_delete error: %s", r.name, err.Error())
	}
	if err := r.resources.Delete(r.ctx); err != nil {
		if grpcerrors.InvalidArgument(err) {
			// this meaans the content on the resource we try to update is badly formatted
			return result(nil, err, true), nil
		}
		return result(nil, err, false), nil
	}
	return result(nil, nil, false), nil
}

func getResource(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var apiVersion, kind starlark.String

	if err := starlark.UnpackArgs("get_resource", args, nil, "apiVersion", &apiVersion, "kind", &kind); err != nil {
		return result(nil, err, true), nil
	}

	u := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": apiVersion.GoString(),
			"kind":       kind.GoString(),
		},
	}

	v, err := util.UnstructuredToStarlarkValue(u)
	if err != nil {
		return result(nil, err, true), nil
	}
	return result(v, nil, false), nil
}

func addFinalizer(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var finalizer starlark.String
	var obj starlark.Value

	if err := starlark.UnpackArgs("add_finalizer", args, nil, "obj", &obj, "finalizer", &finalizer); err != nil {
		return result(nil, err, true), nil
	}

	u, err := util.StarlarkValueToUnstructured(obj)
	if err != nil {
		return result(nil, err, true), nil
	}

	objectMeta, err := meta.Accessor(u)
	if err != nil {
		return result(nil, err, true), nil
	}

	finalizers := objectMeta.GetFinalizers()
	found := false
	for _, f := range finalizers {
		if f == finalizer.GoString() {
			found = true
		}
	}
	if !found {
		if len(finalizers) == 0 {
			finalizers = []string{}
		}
		finalizers = append(finalizers, finalizer.GoString())
	}
	objectMeta.SetFinalizers(finalizers)

	v, err := util.UnstructuredToStarlarkValue(u)
	if err != nil {
		return result(nil, err, true), nil
	}
	return result(v, nil, false), nil
}

func delFinalizer(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var finalizer starlark.String
	var obj starlark.Value

	if err := starlark.UnpackArgs("del_finalizer", args, nil, "obj", &obj, "finalizer", &finalizer); err != nil {
		return result(nil, err, true), nil
	}

	u, err := util.StarlarkValueToUnstructured(obj)
	if err != nil {
		return result(nil, err, true), nil
	}

	objectMeta, err := meta.Accessor(u)
	if err != nil {
		return result(nil, err, true), nil
	}

	finalizers := objectMeta.GetFinalizers()
	for i, f := range finalizers {
		if f == finalizer.GoString() {
			finalizers = append(finalizers[:i], finalizers[i+1:]...)
		}
	}
	objectMeta.SetFinalizers(finalizers)

	v, err := util.UnstructuredToStarlarkValue(u)
	if err != nil {
		return result(nil, err, true), nil
	}
	return result(v, nil, false), nil
}

func getPrefixLength(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var prefix starlark.String

	if err := starlark.UnpackArgs("get_prefixlength", args, nil, "prefix", &prefix); err != nil {
		return result(nil, err, true), nil
	}

	pi, err := iputil.New(prefix.GoString())
	if err != nil {
		return nil, err
	}
	return starlark.MakeInt64(int64(pi.GetPrefixLength())), nil
}

func getAddress(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var prefix starlark.String

	if err := starlark.UnpackArgs("get_address", args, nil, "prefix", &prefix); err != nil {
		return result(nil, err, true), nil
	}

	pi, err := iputil.New(prefix.GoString())
	if err != nil {
		return nil, err
	}
	return starlark.String(pi.GetIPAddress().String()), nil
}

func getSubnetName(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var prefix starlark.String

	if err := starlark.UnpackArgs("get_subnetname", args, nil, "prefix", &prefix); err != nil {
		return result(nil, err, true), nil
	}

	pi, err := iputil.New(prefix.GoString())
	if err != nil {
		return nil, err
	}
	return starlark.String(pi.GetSubnetName()), nil
}

func isIPv4(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var prefix starlark.String

	if err := starlark.UnpackArgs("isIPv4", args, nil, "prefix", &prefix); err != nil {
		return result(nil, err, true), nil
	}

	pi, err := iputil.New(prefix.GoString())
	if err != nil {
		return nil, err
	}

	return starlark.Bool(pi.IsIpv4()), nil
}

func isIPv6(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var prefix starlark.String

	if err := starlark.UnpackArgs("isIPv6", args, nil, "prefix", &prefix); err != nil {
		return result(nil, err, true), nil
	}

	pi, err := iputil.New(prefix.GoString())
	if err != nil {
		return nil, err
	}

	return starlark.Bool(pi.IsIpv6()), nil
}

func isConditionReady(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var obj starlark.Value
	var conditionType starlark.String

	if err := starlark.UnpackArgs("is_conditionready", args, nil, "obj", &obj, "conditionType", &conditionType); err != nil {
		return result(nil, err, true), nil
	}

	u, err := util.StarlarkValueToUnstructured(obj)
	if err != nil {
		return starlark.False, err
	}

	if object.IsConditionTypeReady(u.UnstructuredContent(), conditionType.GoString()) {
		return starlark.True, nil
	}
	return starlark.False, err
}
