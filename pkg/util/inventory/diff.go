/*
Copyright 2024 Nokia.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package inventory

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func diffconfig(beforeByte, afterByte any) (string, error) {
	diff := cmp.Diff(beforeByte, afterByte)
	return diff, nil
}

func diff2(beforeu, afteru *unstructured.Unstructured) (string, error) {
	//var reporter DiffReporter
	//cmp.Equal(beforeu.Object, afteru.Object, cmp.Reporter(&reporter))
	//return reporter.String(), nil
	beforeByte, err := yaml.Marshal(beforeu)
	if err != nil {
		return "", err
	}
	afterByte, err := yaml.Marshal(afteru)
	if err != nil {
		return "", err
	}
	diff := cmp.Diff(beforeByte, afterByte)
	return diff, nil

	/*
		diff := cmp.Diff(beforeu.Object, afteru.Object)
		return diff, nil
	*/
}

// DiffReporter is a simple custom reporter that only records differences
// detected during comparison.
type DiffReporter struct {
	path  cmp.Path
	diffs []string
}

func (r *DiffReporter) PushStep(ps cmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *DiffReporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		vx, vy := r.path.Last().Values()
		r.diffs = append(r.diffs, fmt.Sprintf("%#v:\n\t-: %+v\n\t+: %+v\n", simplifyPath(r.path), vx, vy))
	} else {
		r.diffs = append(r.diffs, simplifyPath(r.path))
	}
}

func (r *DiffReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r *DiffReporter) String() string {
	return strings.Join(r.diffs, "\n")
}

// simplifyPath converts a cmp.Path into a simplified dot notation string.
func simplifyPath(p cmp.Path) string {
	var result []string
	for _, step := range p {
		switch t := step.(type) {
		case cmp.MapIndex:
			// Convert map index steps into dot notation
			result = append(result, fmt.Sprintf("%v", t.Key()))
		case cmp.SliceIndex:
			// Append slice index in brackets
			result = append(result, fmt.Sprintf("[%d]", t.Key()))
		case cmp.TypeAssertion:
			// Ignore type assertions for cleaner paths
			continue
		case cmp.Indirect:
			// Dereference pointers or interfaces
			continue
		}
	}
	return strings.Join(result, ".")
}

func diff(beforeu, afteru *unstructured.Unstructured) (string, error) {
	before, err := yaml.Marshal(beforeu)
	if err != nil {
		return "", err
	}
	after, err := yaml.Marshal(afteru)
	if err != nil {
		return "", err
	}

	tempFileBefore, err := os.CreateTemp("", "before-")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFileBefore.Name()) // Clean up file afterwards

	tempFileAfter, err := os.CreateTemp("", "after-")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFileAfter.Name()) // Clean up file afterwards

	// Write data to temp files
	if _, err := tempFileBefore.WriteString(string(before)); err != nil {
		return "", err
	}
	if _, err := tempFileAfter.WriteString(string(after)); err != nil {
		return "", err
	}

	// Ensure the writes are flushed
	tempFileBefore.Sync()
	tempFileAfter.Sync()

	// Close files to ensure that diff can read them
	tempFileBefore.Close()
	tempFileAfter.Close()

	cmd := exec.Command("diff", []string{"-u", "-N", tempFileBefore.Name(), tempFileAfter.Name()}...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				switch status.ExitStatus() {
				case 0:
					return out.String(), nil
				case 1:
					return out.String(), nil
				default:
					return out.String(), err
				}
			}
		} else {
			return "", err
		}
	}

	return out.String(), nil
}
