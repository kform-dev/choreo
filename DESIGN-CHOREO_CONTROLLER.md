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



# Add the Traefik Helm repository
helm repo add traefik https://helm.traefik.io/traefik

# Update your local Helm chart repository cache
helm repo update

# Install Traefik
helm install traefik traefik/traefik --set "ports.websecure.exposedPort=52000"


helm upgrade --install traefik traefik/traefik \
  --set "additionalArguments={--api.dashboard=true,--entrypoints.grpc.address=:52000/tcp}" \
  --set "service.ports.grpc.port=52000" \
  --set "service.ports.grpc.expose=true" \
  --set "service.ports.grpc.protocol=TCP" 


helm upgrade --install traefik traefik/traefik \
  --namespace default \
  --set "additionalArguments={--api.dashboard=true,--entrypoints.grpc.address=:52000/tcp}" \
  --set "service.ports.grpc={port: 52000, expose: true, protocol: TCP}"

```yaml
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: traefik-config
  namespace: default
data:
  traefik.yaml: |
    api:
      dashboard: true
    entryPoints:
      web:
        address: ":80"
      websecure:
        address: ":443"
      grpc:
        address: ":52000"
    providers:
      kubernetesIngress:
        publishedService:
          enabled: true
    serversTransport:
      insecureSkipVerify: true
EOF
```

```
volumes:
- name: config
  configMap:
    name: traefik-config
volumeMounts:
- name: config
  mountPath: "/config"
  readOnly: true
args:
- --configFile=/config/traefik.yaml
```


apiVersion: traefik.containo.us/v1alpha1
kind: IngressRouteTCP
metadata:
  name: grpc-route
spec:
  entryPoints:
    - grpc
  routes:
  - match: HostSNI(`*`)
    services:
    - name: my-service
      port: 51000



kubectl port-forward --namespace kube-system service/traefik 52000:52000

```yaml
kubectl apply -f - <<EOF
apiVersion: traefik.io/v1alpha1
kind: IngressRouteTCP
metadata:
  name: grpc-route
spec:
  entryPoints:
    - grpc
  routes:
  - match: HostSNI(`*`)
    services:
    - name: my-service
      port: 51000
EOF
```