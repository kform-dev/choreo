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
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/henderiw/logger/log"
	"github.com/kform-dev/choreo/pkg/client/go/resourceclient"
	"github.com/kform-dev/choreo/pkg/proto/resourcepb"
	"github.com/kform-dev/choreo/pkg/server/choreo/loader"
	infrav1alpha1 "github.com/kuidio/kuid/apis/infra/v1alpha1"
	"github.com/sdcio/config-diff/schemaclient"
	"github.com/sdcio/config-server/apis/config"
	configv1alpha1 "github.com/sdcio/config-server/apis/config/v1alpha1"
	"github.com/sdcio/data-server/pkg/tree"
	treejson "github.com/sdcio/data-server/pkg/tree/importer/json"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

type NodeConfig struct {
	Node          *Node
	RunningConfig any                          // json struct
	Configs       []*unstructured.Unstructured // json struct
}

type ConfigValidator struct {
	choreo      Choreo
	nodeConfigs map[string]*NodeConfig
}

func NewConfigValidator(choreo Choreo) *ConfigValidator {
	return &ConfigValidator{
		choreo:      choreo,
		nodeConfigs: map[string]*NodeConfig{},
	}
}

func (r *ConfigValidator) runConfigValidation(ctx context.Context, bctx *BranchCtx) error {

	if err := r.gatherNodeInfo(ctx, bctx); err != nil {
		return err
	}
	if err := r.gatherConfigs(ctx, bctx); err != nil {
		return err
	}
	if err := r.gatherRunningConfigs(ctx, bctx); err != nil {
		return err
	}

	schemastore := r.choreo.GetRootChoreoInstance().SchemaStore()

	var errs error
	for node, nodeInfo := range r.nodeConfigs {
		if nodeInfo.RunningConfig == nil {
			errs = errors.Join(errs, fmt.Errorf("cannot run config validation w/o a running config"))
			continue
		}
		scb := schemaclient.NewMemSchemaClientBound(schemastore, &sdcpb.Schema{
			Vendor:  nodeInfo.Node.GetProvider(),
			Version: nodeInfo.Node.GetVersion(),
		})

		tc := tree.NewTreeContext(tree.NewTreeSchemaCacheClient(node, nil, scb), "test")
		root, err := tree.NewTreeRoot(ctx, tc)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("new root tree err: %v", err))
			continue
		}
		//
		// Load running
		//
		jti := treejson.NewJsonTreeImporter(nodeInfo.RunningConfig)
		err = root.ImportConfig(ctx, jti, tree.RunningIntentName, tree.RunningValuesPrio)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("import config err: %v", err))
			continue
		}
		root.FinishInsertionPhase()

		if err := r.applyRunningConfig(ctx, bctx, nodeInfo.Node, nodeInfo.RunningConfig, ""); err != nil {
			errs = errors.Join(errs, fmt.Errorf("apply running config err: %v", err))
			continue
		}

		j1, err := root.ToJson(false, true)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if err := r.applyRunningConfig(ctx, bctx, nodeInfo.Node, j1, "tree"); err != nil {
			errs = errors.Join(errs, fmt.Errorf("j1 apply running config err: %v", err))
			continue
		}

		// Load configs
		for _, u := range nodeInfo.Configs {
			config := &Config{Unstructured: u}
			for _, val := range config.GetConfigs() {
				jti = treejson.NewJsonTreeImporter(val)
				err = root.ImportConfig(ctx, jti, u.GetName(), 20)
				if err != nil {
					errs = errors.Join(errs, fmt.Errorf("apply node config %serr: %v", u.GetName(), err))
					continue
				}
				root.FinishInsertionPhase()
			}
		}

		j2, err := root.ToJson(false, true)
		if err != nil {
			return err
		}
		if err := r.applyConfig(ctx, bctx, nodeInfo.Node, j2); err != nil {
			errs = errors.Join(errs, fmt.Errorf("j2 apply node config err: %v", err))
			continue
		}
		/*
			oldb, err := yaml.Marshal(j1)
			if err != nil {
				return err
			}
			newb, err := yaml.Marshal(j2)
			if err != nil {
				return err
			}
			fmt.Println("######### old config #########")
			fmt.Println(string(oldb))
			fmt.Println("##############################")
			fmt.Println("######### new config #########")
			fmt.Println(string(newb))
			fmt.Println("##############################")

			diff := cmp.Diff(j1, j2)
			fmt.Println(diff)
		*/

		/*
			newConfigbyteDoc, err := json.MarshalIndent(j, "", " ")
			if err != nil {
				return err
			}
			//fmt.Print(string(newConfigbyteDoc), "\n", "\n")

			oldCOnfigbyteDoc, err := json.MarshalIndent(nodeInfo.RunningConfig, "", " ")
			if err != nil {
				panic(err)
			}
		*/

		/*
			differ := diff.New()
			d, err := differ.Compare(oldCOnfigbyteDoc, newConfigbyteDoc)
			if err != nil {
				// No error can occur
			}

			var aJson map[string]interface{}
			json.Unmarshal(oldCOnfigbyteDoc, &aJson)
		*/

		/*
			config := formatter.AsciiFormatterConfig{
				ShowArrayIndex: true,
				Coloring:       false,
			}

			formatter := formatter.NewAsciiFormatter(aJson, config)
			diffString, err := formatter.Format(d)
			if err != nil {
				// No error can occur
			}
		*/

		/*
			formatter := formatter.NewDeltaFormatter()
			diffString, err := formatter.Format(d)
			if err != nil {
				// No error can occur
			}

		*/
		//fmt.Print(diffString)

		/*
			diffString := cmp.Diff(oldCOnfigbyteDoc, newConfigbyteDoc)
			fmt.Println(diffString)
		*/

		var validationErrors error
		validationErrChan := make(chan error)
		go func() {
			root.Validate(ctx, validationErrChan, true)
			close(validationErrChan)
		}()

		// read from the Error channel
		for e := range validationErrChan {
			validationErrors = errors.Join(validationErrors, e)
		}
		if validationErrors != nil {
			errs = errors.Join(errs, validationErrors)
		}
	}
	return errs
}

func (r *ConfigValidator) applyRunningConfig(ctx context.Context, bctx *BranchCtx, node *Node, data any, suffix string) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	name := node.GetName()
	if suffix != "" {
		name = fmt.Sprintf("%s.%s", name, suffix)
	}

	runningConfig := configv1alpha1.BuildRunningConfig(
		metav1.ObjectMeta{
			Name:      name,
			Namespace: node.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: node.GetAPIVersion(),
					Kind:       node.GetKind(),
					Name:       node.GetName(),
					UID:        node.GetUID(),
					Controller: ptr.To(true),
				},
			},
		},
		configv1alpha1.RunningConfigSpec{},
		configv1alpha1.RunningConfigStatus{Value: runtime.RawExtension{Raw: b}},
	)

	b, err = json.Marshal(runningConfig)
	if err != nil {
		return err
	}
	obj := map[string]any{}
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	return r.choreo.GetClient().Apply(ctx, &unstructured.Unstructured{Object: obj}, &resourceclient.ApplyOptions{
		FieldManager: loader.ManagedFieldManagerInput,
		Branch:       bctx.Branch,
	})
}

func (r *ConfigValidator) applyConfig(ctx context.Context, bctx *BranchCtx, node *Node, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	config := configv1alpha1.BuildConfig(
		metav1.ObjectMeta{
			Name:      node.GetName(),
			Namespace: node.GetNamespace(),
			Labels: map[string]string{
				config.TargetNameKey:           node.GetName(),
				config.TargetNamespaceKey:      node.GetNamespace(),
				"config.sdcio.dev/finalconfig": "true",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: node.GetAPIVersion(),
					Kind:       node.GetKind(),
					Name:       node.GetName(),
					UID:        node.GetUID(),
					Controller: ptr.To(true),
				},
			},
		},
		configv1alpha1.ConfigSpec{
			Priority: 10,
			Config: []configv1alpha1.ConfigBlob{
				{
					Path: "/",
					Value: runtime.RawExtension{
						Raw: b,
					},
				},
			},
		},
		configv1alpha1.ConfigStatus{},
	)

	b, err = json.Marshal(config)
	if err != nil {
		return err
	}
	obj := map[string]any{}
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	return r.choreo.GetClient().Apply(ctx, &unstructured.Unstructured{Object: obj}, &resourceclient.ApplyOptions{
		FieldManager: loader.ManagedFieldManagerInput,
		Branch:       bctx.Branch,
	})
}

func (r *ConfigValidator) gatherNodeInfo(ctx context.Context, bctx *BranchCtx) error {
	ul := &unstructured.UnstructuredList{}
	ul.SetAPIVersion(infrav1alpha1.SchemeGroupVersion.Identifier())
	ul.SetKind(infrav1alpha1.NodeKind)
	if err := r.choreo.GetClient().List(ctx, ul, &resourceclient.ListOptions{
		Branch:       bctx.Branch,
		ExprSelector: &resourcepb.ExpressionSelector{},
	}); err != nil {
		return err
	}

	for _, u := range ul.Items {
		r.nodeConfigs[u.GetName()] = &NodeConfig{
			Node:    &Node{Unstructured: &u},
			Configs: []*unstructured.Unstructured{},
		}

	}
	return nil
}

type Node struct {
	*unstructured.Unstructured
}

func (r *Node) GetProvider() string {
	return getNestedString(r.Object, "spec", "provider")
}

func (r *Node) GetPlatformType() string {
	return getNestedString(r.Object, "spec", "platformType")
}

func (r *Node) GetVersion() string {
	return getNestedString(r.Object, "spec", "version")
}

func getNestedString(obj map[string]interface{}, fields ...string) string {
	val, found, err := unstructured.NestedString(obj, fields...)
	if !found || err != nil {
		return ""
	}
	return val
}

func (r *ConfigValidator) gatherConfigs(ctx context.Context, bctx *BranchCtx) error {
	log := log.FromContext(ctx)
	ul := &unstructured.UnstructuredList{}
	ul.SetAPIVersion(configv1alpha1.SchemeGroupVersion.Identifier())
	ul.SetKind(configv1alpha1.ConfigKind)
	if err := r.choreo.GetClient().List(ctx, ul, &resourceclient.ListOptions{
		Branch:       bctx.Branch,
		ExprSelector: &resourcepb.ExpressionSelector{},
	}); err != nil {
		return err
	}

	for _, u := range ul.Items {
		nodeName, ok := u.GetLabels()[config.TargetNameKey]
		if !ok {
			log.Info("got config without a target name key", "name", u.GetName())
			continue
		}

		nodeConfig, ok := r.nodeConfigs[nodeName]
		if !ok {
			log.Info("got config w/o a matching node %s", "name", u.GetName(), "nodeName", nodeName)
			continue
		}
		// dont add the final configs to the tree since the config was derived from the config snippets
		if nodeName == u.GetName() {
			continue
		}
		if nodeConfig.Configs == nil {
			nodeConfig.Configs = []*unstructured.Unstructured{}
		}
		nodeConfig.Configs = append(nodeConfig.Configs, &u)
	}
	return nil
}

func (r *ConfigValidator) gatherRunningConfigs(ctx context.Context, _ *BranchCtx) error {
	log := log.FromContext(ctx)
	runningConfigs := map[string]any{}
	rootChoreoInstance := r.choreo.GetRootChoreoInstance()
	runningconfigLoader := loader.RunningConfigLoader{
		Cfg:            r.choreo.GetConfig(),
		RepoPath:       rootChoreoInstance.GetRepoPath(),
		PathInRepo:     rootChoreoInstance.GetPathInRepo(),
		RunningConfigs: runningConfigs,
	}
	// this loads schema to the schemastore
	if err := runningconfigLoader.Load(ctx); err != nil {
		return err
	}

	for nodeName, runningConfig := range runningConfigs {
		nodeConfig, ok := r.nodeConfigs[nodeName]
		if !ok {
			log.Info("got runningconfig w/o a matching node %s", "nodeName", nodeName)
			continue
		}
		nodeConfig.RunningConfig = runningConfig
	}
	return nil
}

type Config struct {
	*unstructured.Unstructured
}

func (u *Config) GetConfigs() []any {
	v, found, err := unstructured.NestedFieldNoCopy(u.Object, "spec", "config")
	if !found || err != nil {
		return nil
	}
	items, ok := v.([]interface{})
	if !ok {
		return nil
	}
	configs := []any{}
	for _, item := range items {
		v, ok := item.(map[string]interface{})
		if !ok {
			// TODO log error
			return nil
		}
		val, found, err := unstructured.NestedFieldNoCopy(v, "value")
		if !found || err != nil {
			// TODO log error
			continue
		}
		configs = append(configs, val)
	}
	return configs
}
