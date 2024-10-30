# conversion

kuid resource allocators use the internal version

## loading crd
- if storage version has k8s version we retain only 1 version
- if storage version has no k8s version we retain both (this is to accomodate reuse of the backend code)

-> we expect only a single external version
-> we expect the internal and external version to be the same

Internal reconcilers can have an internal storage version (ipam, vlan, extcomm)
-> we can validate this during crd loading
-> conversions are known ahead of time

## where do we handle the conversion

ideally before handing to the storage layer

we assume right now the spec/status of internal and external versions are the same -> hence we do a trick to just modify the apiVersion in the unstructured context.



## kube version 

src: apimachinery

```go

type versionType int

const (
	// Bigger the version type number, higher priority it is
	versionTypeAlpha versionType = iota
	versionTypeBeta
	versionTypeGA
)

var kubeVersionRegex = regexp.MustCompile("^v([\\d]+)(?:(alpha|beta)([\\d]+))?$")

func parseKubeVersion(v string) (majorVersion int, vType versionType, minorVersion int, ok bool) {
	var err error
	submatches := kubeVersionRegex.FindStringSubmatch(v)
	if len(submatches) != 4 {
		return 0, 0, 0, false
	}
	switch submatches[2] {
	case "alpha":
		vType = versionTypeAlpha
	case "beta":
		vType = versionTypeBeta
	case "":
		vType = versionTypeGA
	default:
		return 0, 0, 0, false
	}
	if majorVersion, err = strconv.Atoi(submatches[1]); err != nil {
		return 0, 0, 0, false
	}
	if vType != versionTypeGA {
		if minorVersion, err = strconv.Atoi(submatches[3]); err != nil {
			return 0, 0, 0, false
		}
	}
	return majorVersion, vType, minorVersion, true
}
```