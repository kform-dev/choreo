# TODO

Priority
- streaming once runner

## choreo

- (P1) update choreo docs
- (P0) streaming once runner
- {P2} snapshots are stored in memory. Is this the right approach ?
- (P1) snapshots add the detailed result.
- (P2) garbage collector: do we ignore the version -> currently we do a special trick in kuid to change the ownereference to v1alpha1 in the backend apis( as, vlan, genid, etc)
- (P2) fetch repo assumes main branch -> need to check the actual used branch

- (P2) do we need to add reconcilers, libraries to the api or not ?
    - right now we dont
- (P2) how to handle secrets? Vault ?
- (P2) k8s API versus grpc API ??
- (P2) project scaffold
- (P2) pydantic api definition
- OK config-diff integration

## kubenet
- (P1,5) fix genid
- (P1) addtional logic and config generation

## choreo controller

- (P2) Variant controller
- (P2) Approval controller

## kuid

- (P2) rework the generic backend in the same way as IPAM -> allows for real claims
- (P2) proto generation
- (P2) reconcilers (non IPAM)

## main

- (P2) align storage backend kuid and chorio