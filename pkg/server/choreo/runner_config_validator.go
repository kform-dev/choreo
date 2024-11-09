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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type NodeConfig struct {
	Provider      string
	PlatformType  string
	Version       string
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
			Vendor:  nodeInfo.Provider,
			Version: nodeInfo.Version,
		})

		tc := tree.NewTreeContext(tree.NewTreeSchemaCacheClient(node, nil, scb), "test")
		root, err := tree.NewTreeRoot(ctx, tc)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		//
		// Load running
		//
		jti := treejson.NewJsonTreeImporter(nodeInfo.RunningConfig)
		err = root.ImportConfig(ctx, jti, tree.RunningIntentName, tree.RunningValuesPrio)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		root.FinishInsertionPhase()

		// Load configs
		for _, u := range nodeInfo.Configs {
			fmt.Println("loading config", node, u.GetName())
			config := &Config{Unstructured: u}
			for _, val := range config.GetConfigs() {
				jti = treejson.NewJsonTreeImporter(val)
				err = root.ImportConfig(ctx, jti, u.GetName(), 20)
				if err != nil {
					return err
				}
				root.FinishInsertionPhase()
			}
		}
		/*
			title := "CONFIG"

			fmt.Printf("\n%s IN %q FORMAT:\n\n", title, strings.ToUpper("json"))
			j, err := root.ToJson(false)
			if err != nil {
				return err
			}
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
		node := &Node{Unstructured: &u}
		r.nodeConfigs[u.GetName()] = &NodeConfig{
			Provider:     node.GetProvider(),
			PlatformType: node.GetPlatformType(),
			Version:      node.GetVersion(),
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
