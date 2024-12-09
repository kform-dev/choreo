# TODO

## choreo

- {P2} snapshots are stored in memory. Is this the right approach ?
- (P2) garbage collector: do we ignore the version -> currently we do a special trick in kuid to change the ownereference to v1alpha1 in the backend apis( as, vlan, genid, etc)
- (P2) fetch repo assumes main branch -> need to check the actual used branch
- (P2) do we need to add reconcilers, libraries to the api or not ? -> right now we don't
- (P2) how to handle secrets? Vault ?
- (P2) k8s API versus grpc API ??
- (P2) project scaffold
- (P2) pydantic api definition
- float versus int64
- data export/import
- overall processors -> apply replacement alike

## kubenet
- (P1) addtional config templates

## choreo controller

- (P2) Variant controller
- (P2) Approval controller

## kuid

- (P2) proto generation
- (P2) per component enablement
- (P2) POSTGRES DB

## main

- (P2) align storage backend kuid and chorio