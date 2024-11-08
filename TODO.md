# TODO

## choreo

- update choreo docs
- garbage collector: do we ignore the version -> currently we do a special trick in kuid to change the ownereference to v1alpha1 in the backend apis( as, vlan, genid, etc)
- fetch repo assumes main branch -> need to check the actual used branch
- snapshots are stored in memory. Is this the right approach ?
- snapshots add the detailed result.
- do we need to add reconcilers, libraries to the api or not ?
    - right now we dont
- how to handle secrets? Vault ?
- k8s API versus grpc API ??
- project scaffold

## kubenet
- config generation
- config-diff integration

## choreo controller

- Variant controller
- Approval controller

## kuid

- rework the generic backend in the same way as IPAM -> allows for real claims
- proto generation
- reconcilers (non IPAM)

## main

- align storage backend kuid and chorio