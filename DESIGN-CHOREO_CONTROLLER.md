# controller

Choreo
- repo
- dir
- ref

-> create a pod (configmap)
-> option 1
    - repo (URL, Ref, Dir, Secret)
    - file watcher
        -> check latest choreo status (compare to existing system)
        -> if different
            -> stop choreo runner
            -> stop choreo server
            -> clone repo or update repo
            -> checkout the ref
            -> start choreo server
            -> start choreo runner
-> option 2
    - create an api service
        -> supply url/ref/dir/secret
        -> 

Choreo

```yaml
apiVersion: choreo.kform.dev/v1alpha1
kind: Choreo
metadata:
  name: greeting
spec:
  url: https://github.com/kform-dev/choreo-examples.git
  directory: greeting-ref
  ref:
    type: hash
    name: f90ee1ae44bd6a1568e9d5f5e9d2ea1850de6693
```

-> create choreo deployment, service


ChoreoVariant

```yaml
apiVersion: choreo.kform.dev/v1alpha1
kind: Choreo
metadata:
  name: greeting
spec:
  ref: 
```

