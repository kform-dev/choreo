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

package configgenerator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/henderiw/store"
	"github.com/henderiw/store/memory"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	configv1alpha1 "github.com/kform-dev/choreo/apis/config/v1alpha1"
	"github.com/kform-dev/choreo/pkg/client/go/discovery"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/postfunctions/configgenerator/gotemplates/parser"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/kustomize/kyaml/utils"
	"sigs.k8s.io/yaml"
)

type RenderFn func(ctx context.Context, template string, u *unstructured.Unstructured, w io.Writer) error

type ProviderCtx struct {
	LangTech choreov1alpha1.LangTechType
	Files    map[string]string
	Renderer RenderFn
}

type ConfigGenerator struct {
	ConfigGenerator *choreov1alpha1.ConfigGenerator
	Providers       store.Storer[*ProviderCtx]
	GVs             sets.Set[schema.GroupResource]
	resourceClient  resourceclient.Client
}

func New(cg *choreov1alpha1.ConfigGenerator) *ConfigGenerator {
	return &ConfigGenerator{
		Providers:       memory.NewStore[*ProviderCtx](nil),
		GVs:             sets.New[schema.GroupResource](),
		ConfigGenerator: cg,
	}
}

func (r *ConfigGenerator) AddFiles(ctx context.Context, provider, name, data string) error {
	key := store.ToKey(provider)
	providerCtx, err := r.Providers.Get(key)
	if err != nil {
		providerCtx := &ProviderCtx{
			Files: map[string]string{
				name: data,
			},
		}
		return r.Providers.Create(key, providerCtx)
	}
	if providerCtx.Files == nil {
		providerCtx.Files = map[string]string{}
	}
	providerCtx.Files[name] = data
	return r.Providers.Update(key, providerCtx)
}

func (r *ConfigGenerator) AddGroupResource(gr schema.GroupResource) {
	r.GVs.Insert(gr)
}

// we validate if the langtech is valid and if there is consistency with all the
// files as only a single langtech type per vendor is supported
func (r *ConfigGenerator) ValidateLangTech(ctx context.Context, provider, ext string) error {
	newLangTech := getLangType(ext)
	if newLangTech == choreov1alpha1.Invalid_LangTechType {
		return fmt.Errorf("invalid langtechtype %s", ext)
	}

	key := store.ToKey(provider)
	providerCtx, err := r.Providers.Get(key)
	if err != nil {
		providerCtx := &ProviderCtx{
			LangTech: newLangTech,
			Files:    map[string]string{},
		}
		return r.Providers.Create(key, providerCtx)
	}
	if providerCtx.LangTech != newLangTech {
		return fmt.Errorf("inconistsent langtechtype got %s and %s", newLangTech.String(), providerCtx.LangTech.String())
	}
	return nil

}

func getLangType(ext string) choreov1alpha1.LangTechType {
	switch ext {
	case ".tpl":
		return choreov1alpha1.GoTemplate_LangTechType
	default:
		return choreov1alpha1.Invalid_LangTechType
	}
}

func (r *ConfigGenerator) UpdateRenderer(ctx context.Context) error {
	// update the renderer
	providers := map[string]*ProviderCtx{}
	var errm error
	r.Providers.List(func(key store.Key, pc *ProviderCtx) {
		switch pc.LangTech {
		case choreov1alpha1.GoTemplate_LangTechType:
			p, err := parser.New(pc.Files)
			if err != nil {
				errm = errors.Join(errm, err)
				return
			}
			pc.Renderer = p.Render // register the renderer
			providers[key.Name] = pc
		default:
			errm = errors.Join(errm, fmt.Errorf("unsupported langtype %s", pc.LangTech.String()))
		}
	})
	if errm != nil {
		return errm
	}
	for provider, providerCtx := range providers {
		r.Providers.Update(store.ToKey(provider), providerCtx)
	}
	return nil
}

func (r *ConfigGenerator) Build(ctx context.Context, resourceClient resourceclient.Client, discoveryClient discovery.CachedDiscoveryInterface, branchName string) error {
	r.resourceClient = resourceClient
	apiResources, err := discoveryClient.APIResources(ctx, branchName)
	if err != nil {
		return err
	}
	ul, err := r.GetResources(ctx)
	if err != nil {
		return err
	}
	providermap, err := r.BuildProviderMap(ul)
	if err != nil {
		return err
	}

	var errm error
	for _, gr := range r.GVs.UnsortedList() {
		apiGroup := apiResources.GetAPIResourceGroup(gr.Group, gr.Resource)
		apiVersion := schema.GroupVersion{Group: apiGroup.Group, Version: apiGroup.Version}.String()
		kind := apiGroup.Kind

		ul := &unstructured.UnstructuredList{}
		ul.SetAPIVersion(apiVersion)
		ul.SetKind(kind)
		if err := r.resourceClient.List(ctx, ul, &resourceclient.ListOptions{
			ExprSelector: &resourcepb.ExpressionSelector{},
		}); err != nil {
			errm = errors.Join(errm, fmt.Errorf("cannot list apiVersion %s, kind %s", apiVersion, kind))
		}

		// walk over the resources
		for _, u := range ul.Items {
			// get key
			key, err := r.getDstMapString(&u)
			if err != nil {
				errm = errors.Join(errm, fmt.Errorf("cannot get key for %s %s %s, err %s", u.GetAPIVersion(), u.GetKind(), u.GetName(), err.Error()))
				continue
			}

			// get provider
			provider, ok := providermap[key]
			if !ok {
				errm = errors.Join(errm, fmt.Errorf("cannot get provider for %s %s %s", u.GetAPIVersion(), u.GetKind(), u.GetName()))
				continue
			}

			// get providerCtx
			providerCtx, err := r.Providers.Get(store.ToKey(provider))
			if err != nil {
				errm = errors.Join(errm, fmt.Errorf("cannot get provider context for %s %s %s", u.GetAPIVersion(), u.GetKind(), u.GetName()))
				continue
			}

			var buf bytes.Buffer
			if err := providerCtx.Renderer(ctx, GroupResource2String(gr), &u, &buf); err != nil {
				errm = errors.Join(errm, fmt.Errorf("render failed for %s %s %s, err %s", u.GetAPIVersion(), u.GetKind(), u.GetName(), err.Error()))
				continue
			}

			cfgdata := map[string]any{}
			if err := yaml.Unmarshal(buf.Bytes(), &cfgdata); err != nil {
				errm = errors.Join(errm, err)
				continue
			}
			jsoncfgData, err := json.Marshal(cfgdata)
			if err != nil {
				errm = errors.Join(errm, err)
				continue
			}
			node, ok, err := unstructured.NestedString(u.Object, "spec", "node")
			if !ok || err != nil {
				errm = errors.Join(errm, fmt.Errorf("no node information for %s %s %s", u.GetAPIVersion(), u.GetKind(), u.GetName()))
				continue
			}

			cfg := configv1alpha1.BuildConfig(
				metav1.ObjectMeta{
					Namespace: u.GetNamespace(),
					Name:      u.GetName(),
					Labels: map[string]string{
						"config.sdcio.dev/targetName":      node,
						"config.sdcio.dev/targetNamespace": u.GetNamespace(),
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: u.GetAPIVersion(),
							Kind:       u.GetKind(),
							Name:       u.GetName(),
							UID:        u.GetUID(),
						},
					},
				},
				configv1alpha1.ConfigSpec{
					Priority: 10,
					Config: []configv1alpha1.ConfigBlob{
						{
							Path: "/",
							Value: runtime.RawExtension{
								Raw: jsoncfgData,
							},
						},
					},
				},
				configv1alpha1.ConfigStatus{},
			)
			b, err := json.Marshal(cfg)
			if err != nil {
				errm = errors.Join(errm, err)
				continue
			}

			obj := map[string]any{}
			if err := json.Unmarshal(b, &obj); err != nil {
				errm = errors.Join(errm, err)
				continue
			}

			if err := r.resourceClient.Apply(ctx, &unstructured.Unstructured{Object: obj}, &resourceclient.ApplyOptions{
				FieldManager: "config",
			}); err != nil {
				errm = errors.Join(errm, err)
				continue
			}
		}
	}
	return errm
}

func (r *ConfigGenerator) GetResources(ctx context.Context) (*unstructured.UnstructuredList, error) {
	gvk := r.ConfigGenerator.GetProviderSelectorGVK()
	// for every node check the providers
	ul := &unstructured.UnstructuredList{}
	ul.SetAPIVersion(gvk.GroupVersion().String())
	ul.SetKind(gvk.Kind)
	if err := r.resourceClient.List(ctx, ul, &resourceclient.ListOptions{
		ExprSelector: &resourcepb.ExpressionSelector{},
	}); err != nil {
		return nil, err
	}
	return ul, nil
}

func (r *ConfigGenerator) BuildProviderMap(ul *unstructured.UnstructuredList) (map[string]string, error) {
	// get the providers from all the nodes
	src2providermap := map[string]string{}
	matchKeys := r.ConfigGenerator.GetMatchKeys()

	var errm error
	for _, u := range ul.Items {
		provider, err := getValue(&u, r.ConfigGenerator.Spec.ProviderSelector.FieldPath)
		if err != nil {
			errm = errors.Join(errm, err)
		}
		matchStr, err := getSrcMapString(&u, matchKeys)
		if err != nil {
			errm = errors.Join(errm, err)
		}
		src2providermap[matchStr] = provider
	}
	return src2providermap, errm
}

func getValue(u *unstructured.Unstructured, fieldpath string) (string, error) {
	fieldPath := utils.SmarterPathSplitter(fieldpath, ".")
	value, ok, err := unstructured.NestedFieldCopy(u.Object, fieldPath...)
	if !ok || err != nil {
		return "", fmt.Errorf("obj %s.%s.%s value not found based on fieldpath %s, err %v", u.GetAPIVersion(), u.GetKind(), u.GetName(), fieldpath, err)
	}
	return fmt.Sprintf("%v", value), nil
}

func getSrcMapString(u *unstructured.Unstructured, keys []string) (string, error) {
	var errm error
	var sb strings.Builder
	for i, k := range keys {
		v, err := getValue(u, k)
		if err != nil {
			errm = errors.Join(errm, err)
			continue
		}
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%s=%s", k, v))
	}
	return sb.String(), errm
}

func (r *ConfigGenerator) getDstMapString(u *unstructured.Unstructured) (string, error) {
	match := r.ConfigGenerator.GetProviderMatch()
	var errm error
	var sb strings.Builder
	for i, k := range r.ConfigGenerator.GetMatchKeys() {
		fieldpath, ok := match[k]
		if !ok {
			errm = errors.Join(errm, fmt.Errorf("key: %s not found", k))
			continue
		}

		v, err := getValue(u, fieldpath)
		if err != nil {
			errm = errors.Join(errm, err)
			continue
		}
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%s=%s", k, v))
	}
	return sb.String(), errm
}

func String2GroupResource(s string) (schema.GroupResource, error) {
	parts := strings.Split(s, "_")
	if len(parts) != 2 {
		return schema.GroupResource{}, fmt.Errorf("expecting <group>_<version>")
	}
	return schema.GroupResource{Group: parts[0], Resource: parts[1]}, nil
}

func GroupResource2String(gr schema.GroupResource) string {
	return fmt.Sprintf("%s_%s", gr.Group, gr.Resource)
}
