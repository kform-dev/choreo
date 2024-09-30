package inventory

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

func (inv Inventory) Extract(path string) error {
	var errm error
	for _, treenode := range inv {
		u := treenode.Resource
		b, err := yaml.Marshal(treenode.Resource)
		if err != nil {
			errm = errors.Join(errm, err)
			continue
		}

		file, err := os.Create(filepath.Join(path, fmt.Sprintf(
			"%s.%s.%s.%s.yaml",
			strings.ReplaceAll(u.GetAPIVersion(), "/", "_"),
			u.GetKind(),
			u.GetNamespace(),
			u.GetName(),
		)))
		if err != nil {
			return err
		}
		defer file.Close()
		fmt.Fprintf(file, "---\n%s\n", string(b))
	}
	return nil
}
