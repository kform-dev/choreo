# apply streaming

- flags: dynamic file flags
    - Apply the configuration in pod.json to a pod: `kubectl apply -f ./pod.json`
    - Apply the JSON passed into stdin to a pod: `cat pod.json | kubectl apply -f -`
    - Apply the configuration from all files that end with '.json': `kubectl apply -f '*.json'`

- uses resource builder to get objects

```go
func (o *ApplyOptions) GetObjects() ([]*resource.Info, error) {
	var err error = nil
	if !o.objectsCached {
		r := o.Builder.
			Unstructured().
			Schema(o.Validator).
			ContinueOnError().
			NamespaceParam(o.Namespace).DefaultNamespace().
			FilenameParam(o.EnforceNamespace, &o.DeleteOptions.FilenameOptions).
			LabelSelectorParam(o.Selector).
			Flatten().
			Do()

		o.objects, err = r.Infos()

		if o.ApplySet != nil {
			if err := o.ApplySet.AddLabels(o.objects...); err != nil {
				return nil, err
			}
		}

		o.objectsCached = true
	}
	return o.objects, err
}
```