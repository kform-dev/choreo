# hierarchical locader

1. load the upstream ref in a hierarchical way
    -> we create child choreo instances

we allow for :
- root -> childroot (crd/reconcilers/data) -> childcrd
- root -> childcrd (crd/lib)
- root -> data/reconcilers/libs/crds

kubenet
+- childroot1 topologyref priority 10
  +- childinstance (crd)
  +- childinstance (crd)
  +- reconcilers
  +- data (templates)
+- reconcilers
+- in
+- refs

2. init apis globally (we assume no api overlap right now or inconsistencies)

3. libraries are loaded and stored per rootinstance or childrootinstance

4. reconcilers are loaded and stored per rootinstance or childrootinstance

5. data is loaded globally

6. run garbage collection

## Running

we run the instances in hierarchy

