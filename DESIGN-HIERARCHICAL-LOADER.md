# hierarchical locader

1. load upstream refs in a hierarchical way
    - crd refs
    - child choreoinstance refs

2. we load the apis from all of them from the beginning
    - TODO: need to check conflict (different upstream refs use the same api object with a different reference)
    - We load the libraries in the same way as the CRDs


3. running the instances
    - we can run the instances in sequence as per priority field -> first focus
    - we can run them simulteniously


current approach:

- assumption is that the library and input data is rendered to input directory
- at server start: for all root instances and child instances
    - load API
- at runner start: 
    - get api(s)
    - load input data
    - look at inventory/garbage collect (non choreo resources)
    - load reconcilers and libs from the apis
    - start the runner

new approach:

1. build the parent child relationships 
2. load the apis + reconcilers + libraries in memory per child root
    -> we do this at the runner since things can change along the path
3. snapshot should keep track of this context


root (priority 100)
data for child 
+- childroot1 (priority 10)
  +- child crd
  +- child crd
+- childroot2 (priority 20)
  +- child crd
  +- child crd
+- child crd
+- child crd

root (priority 100)
+- child crd
+- child crd



examples

topology
- data for childroot1 (topology)
+- childroot1 topologyref
  +- child crd
  +- child crd
  data templates

sequence
- load apis, libraries and reconcilers per choreoInstance hierarchy
- run childroot topologyref
    - load data 
    - run reconcilers when data is available
- run root
    - load data
    - run topology child reconcilers since data is attached to the child reconcilers

kubenet
- data for childroot1 (topology)
+- childroot1 topologyref priority 10
  +- child crd
  +- child crd
  data templates
- data for kubenet (ipindex, network design)
+- childroot1 kubenet
  +- child crd
  +- child crd

sequence
- load apis, libraries and reconcilers per choreoInstance hierarchy
- run childroot topologyref
    - load data 
    - run reconcilers when data is available
- run childroot kubenet
    - load data 
    - don't reconcilers since no data is available
- run root
    - load data
    - run kubenet child reconcilers since data is attached to the kubenet child reconciler


run data globallay

