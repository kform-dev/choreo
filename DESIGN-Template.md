# proxy

```yaml
apiVersion: config.sdcio.dev/v1alpha1
kind: Config
metadata:
  name: {{ metadata.name }}
  namespace: default
  labels:
    config.sdcio.dev/targetName: dev1
    config.sdcio.dev/targetNamespace: default
spec:
  priority: 10
  config:
  - path: /
    value: 
      {{- template "srlsystem" .}}
```