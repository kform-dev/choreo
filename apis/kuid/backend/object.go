/*
Copyright 2023 The Nephio Authors.

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

package backend

import (
	"github.com/henderiw/idxtable/pkg/table"
	"github.com/henderiw/idxtable/pkg/tree"
	"github.com/henderiw/idxtable/pkg/tree/gtree"
	"github.com/henderiw/store"
	condv1alpha1 "github.com/kform-dev/choreo/apis/condition/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IndexObject interface {
	Object
	GetTree() gtree.GTree
	GetType() string
	GetMinID() *uint64
	GetMaxID() *uint64
	GetMinClaim() ClaimObject
	GetMaxClaim() ClaimObject
	GetMax() uint64
	GetStatus() (map[string]any, error)
}

type ClaimObject interface {
	Object
	GetIndex() string
	GetSelector() *metav1.LabelSelector
	GetOwnerSelector() (labels.Selector, error)
	GetLabelSelector() (labels.Selector, error)
	GetClaimLabels() labels.Set
	ValidateOwner(labels labels.Set) error
	GetClaimType() ClaimType
	GetStaticID() *uint64
	GetStaticTreeID(t string) tree.ID
	GetClaimID(t string, id uint64) tree.ID
	GetRange() *string
	GetRangeID(t string) (tree.Range, error)
	GetTable(t string, to, from uint64) table.Table
	SetStatusRange(*string)
	SetStatusID(*uint64)
	GetStatusID() *uint64
	GetClaimRequest() string
	GetClaimResponse() string
	GetStatus() (map[string]any, error)
}

type EntryObject interface {
	Object
	GetIndex() string
	GetClaimType() ClaimType
	GetSpecID() string
}

type Object interface {
	client.Object
	GetNamespacedName() types.NamespacedName
	GetKey() store.Key
	GetCondition(t condv1alpha1.ConditionType) condv1alpha1.Condition
	SetConditions(c ...condv1alpha1.Condition)
	ValidateSyntax(s string) field.ErrorList
}
