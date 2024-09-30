# change management

1. User creates a branch and checkout this branch (we need to detect the references)
2. changes input (data, reconcilers, library) -> we map these files to the database (db directory)
    -> we keep the input seperate in memory (to allow)
    -> alternative is using an api model where the user creates a script to import (this allows to restire)
3. run the reconcilers
4. iterate between 2/3
5. stash changes
6. push to remote
7. pull + iterate
6. merge 


Is the directory a source of truth or not

## create a blueprint

- apis
- reconcilers
- libs
- input is used as examples -> maybe we use a testdb to test things out

flow
1. create a repo dir (git init -b main)
2. create api(s) - generate using crdgen
3. create reconcilers/libs
4. create test input -> unit test environment 
5. test with choreo run
6. validate
7. checkin

## create a blueprint instance (could be done using a package variant) e.g. topology

flow
1. create a repo dir (git init -b main)
2. using a ref we copy clone the data from an existing repo (in this repo)
3. add your input
4. choreo run
5. iterate
6. validate
7. checkin

## create a blueprint instance which references a blueprint instance

upstream ref -> copy the data in the repo, in the respective directories
data ref -> need to data to run the reconciler (you cannot change the data but just use it/consume it)
    -> maybe for identifiers: like ip/vlan/as/etc we might need to provide an update

repo1 (blueprint)
-> dir1
-> dir2


```yaml
apiVersion: choreo.kform.dev/v1alpha1
kind: UpstreamRef
metadata:
  name: topology
spec:
  url: guthub.com/org/repo
  directory: a/b/topology
  ref:
    type: hash
    id: 1234567
```

local repo
- apis
- reconciler/libs
- inputs
- refs/upstreamref.yaml
- refs/dataref.yaml
- refs/<repo>/
- refs/<repo>/
