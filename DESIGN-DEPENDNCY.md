# 

- API(s)/CRD(s)
- Reconcilers
- Libraries

semantic versioning
compatibility -> reconciler2CRD => ??? can we not use GVK
pkg mgmt for libraries -> see Cargo
manifest file
use a service catalog -> where libs/reconcilers/libs reside

us e a dependency graph to resolve dependencies

## upstream ref

like topology -> if we don't store the result we might need to run once within their context; we run them in sequence and provide a context of the upstream