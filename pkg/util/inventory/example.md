apiVersion: topo.kubenet.dev/v1alpha1
kind:Topology
metadata":
    creationTimestamp: 2024-08-26T18:54:27Z
  	managedFields: 
  	- apiVersion: topo.kubenet.dev/v1alpha1
      fieldsType: FieldsV1 
      fieldsV1: 
        f:spec: map[string]any{"f:defaults": map[string]any{"f:platformType": map[string]any{}, "f:provider": map[string]any{}, "f:region": map[string]any{}, "f:site": map[string]any{}}, "f:links": map[string]any{}, "f:nodes": map[string]any{}}}, "manager": string("choreo"), ...},
+ 			map[string]any{
+ 				"apiVersion": string("topo.kubenet.dev/v1alpha1"),
+ 				"fieldsType": string("FieldsV1"),
+ 				"fieldsV1": map[string]any{
+ 					"f:spec": map[string]any{
+ 						"f:defaults": map[string]any{},
+ 						"f:links":    map[string]any{},
+ 						"f:nodes":    map[string]any{},
+ 					},
+ 					"f:status": map[string]any{"f:conditions": map[string]any{...}},
+ 				},
+ 				"manager":   string("topo.kubenet.dev_topology_index"),
+ 				"operation": string("Apply"),
+ 				"time":      string("2024-08-26T18:54:28Z"),
+ 			},
+ 			map[string]any{
+ 				"apiVersion": string("topo.kubenet.dev/v1alpha1"),
+ 				"fieldsType": string("FieldsV1"),
+ 				"fieldsV1": map[string]any{
+ 					"f:spec": map[string]any{
+ 						"f:defaults": map[string]any{},
+ 						"f:links":    map[string]any{},
+ 						"f:nodes":    map[string]any{},
+ 					},
+ 					"f:status": map[string]any{"f:conditions": map[string]any{...}},
+ 				},
+ 				"manager":   string("topo.kubenet.dev_topology_nodelink"),
+ 				"operation": string("Apply"),
+ 				"time":      string("2024-08-26T18:54:28Z"),
+ 			},
  		},
  		"name":            string("kubenet"),
  		"namespace":       string("default"),
- 		"resourceVersion": string("0"),
+ 		"resourceVersion": string("2"),
  		"uid":             string("d8b4fb81-ab43-4e0d-bdcf-5fc06edc6cc1"),
  	},
  	"spec": map[string]any{"defaults": map[string]any{"platformType": string("ixrd3"), "provider": string("srlinux.nokia.com"), "region": string("region1"), "site": string("us-east")}, "links": []any{map[string]any{"endpoints": []any{map[string]any{"adaptor": string("sfp"), "endpoint": float64(1), "node": string("node1"), "port": float64(1)}, map[string]any{"adaptor": string("sfp"), "endpoint": float64(1), "node": string("node2"), "port": float64(1)}}}}, "nodes": []any{map[string]any{"name": string("node1")}, map[string]any{"name": string("node2")}}},
+ 	"status": map[string]any{
+ 		"conditions": []any{
+ 			map[string]any{
+ 				"lastTransitionTime": string("2024-08-26T18:54:28Z"),
+ 				"message":            string(""),
+ 				"reason":             string("Ready"),
+ 				"status":             string("True"),
+ 				...
+ 			},
+ 			map[string]any{
+ 				"lastTransitionTime": string("2024-08-26T18:54:28Z"),
+ 				"message":            string(""),
+ 				"reason":             string("Ready"),
+ 				"status":             string("True"),
+ 				...
+ 			},
+ 		},
+ 	},
  }