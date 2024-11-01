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

package resource

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/kform-dev/choreo/pkg/client/go/resourcemapper"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured/unstructuredscheme"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// Builder creates context and builds vistors
// when calling do a result is returned which is a prepared vistor list
// when calling result.Infos the visitor list is ran and the objects are returned
// e.g. for resource and names given this is provided

var FileExtensions = []string{".json", ".yaml", ".yml"}
var InputExtensions = append(FileExtensions, "stdin")

const defaultHttpGetAttempts = 3
const pathNotExistError = "the path %q does not exist"

type Builder struct {
	proxy  types.NamespacedName
	branch string
	// local indicates that the builder does not interact with the server
	//local bool
	mapper *mapper

	// objectTyper is statically determinant per-command invocation based on your internal or unstructured choice
	// it does not ever need to rely upon discovery.
	objectTyper runtime.ObjectTyper

	// resources based on gr, names
	gr    *schema.GroupResource
	names []string
	// this is fille using -f
	paths []Visitor
	// resource context
	stream     bool
	stdinInUse bool
	dir        bool

	flatten bool

	visitorConcurrency int

	continueOnError bool
	errs            error
}

type FilenameOptions struct {
	Filenames []string
	Kustomize string
	Recursive bool
}

func (o *FilenameOptions) validate() error {
	return nil
}

func NewBuilder(resourceMapper resourcemapper.Mapper, proxy types.NamespacedName, branch string) *Builder {
	return newBuilder(resourceMapper, proxy, branch)
}

func newBuilder(resourceMapper resourcemapper.Mapper, proxy types.NamespacedName, branch string) *Builder {
	return &Builder{
		mapper: &mapper{
			resourceMapper: resourceMapper,
		},
		branch: branch,
		proxy:  proxy,
	}
}

// Unstructured updates the builder so that it will request and send unstructured
// objects. Unstructured objects preserve all fields sent by the server in a map format
// based on the object's JSON structure which means no data is lost when the client
// reads and then writes an object. Use this mode in preference to Internal unless you
// are working with Go types directly.
func (b *Builder) Unstructured() *Builder {
	/*
		if b.mapper != nil {
			b.errs = errors.Join(b.errs, fmt.Errorf("another mapper was already selected, cannot use unstructured types"))
			return b
		}
	*/
	b.objectTyper = unstructuredscheme.NewUnstructuredObjectTyper()
	b.mapper.decoder = &metadataValidatingDecoder{unstructured.UnstructuredJSONScheme}
	return b
}

// Mapper returns a copy of the current mapper.
func (b *Builder) Mapper() *mapper {
	mapper := *b.mapper
	return &mapper
}

// Flatten will convert any objects with a field named "Items" that is an array of runtime.Object
// compatible types into individual entries and give them their own items. The original object
// is not passed to any visitors.
func (r *Builder) Flatten() *Builder {
	r.flatten = true
	return r
}

// ContinueOnError will attempt to load and visit as many objects as possible, even if some visits
// return errors or some objects cannot be loaded. The default behavior is to terminate after
// the first error is returned from a VisitorFunc.
func (b *Builder) ContinueOnError() *Builder {
	b.continueOnError = true
	return b
}

// URL accepts a number of URLs directly.
func (b *Builder) URL(httpAttemptCount int, urls ...*url.URL) *Builder {
	for _, u := range urls {
		b.paths = append(b.paths, &URLVisitor{
			URL:              u,
			StreamVisitor:    NewStreamVisitor(nil, b.mapper, u.String()),
			HttpAttemptCount: httpAttemptCount,
		})
	}
	return b
}

// Path accepts a set of paths that may be files, directories (all can containing
// one or more resources). Creates a FileVisitor for each file and then each
// FileVisitor is streaming the content to a StreamVisitor. If ContinueOnError() is set
// prior to this method being called, objects on the path that are unrecognized will be
// ignored (but logged at V(2)).
func (b *Builder) Path(recursive bool, paths ...string) *Builder {
	for _, p := range paths {
		_, err := os.Stat(p)
		if os.IsNotExist(err) {
			b.errs = errors.Join(b.errs, fmt.Errorf(pathNotExistError, p))
			continue
		}
		if err != nil {
			b.errs = errors.Join(b.errs, fmt.Errorf("the path %q cannot be accessed: %v", p, err))
			continue
		}

		visitors, err := ExpandPathsToFileVisitors(b.mapper, p, recursive, FileExtensions)
		if err != nil {
			b.errs = errors.Join(b.errs, fmt.Errorf("error reading %q: %v", p, err))
		}
		if len(visitors) > 1 {
			b.dir = true
		}

		b.paths = append(b.paths, visitors...)
	}
	if len(b.paths) == 0 && b.errs != nil {
		b.errs = errors.Join(b.errs, fmt.Errorf("error reading %v: recognized file extensions are %v", paths, FileExtensions))
	}
	return b
}

// FilenameParam groups input in two categories: URLs and files (files, directories, STDIN)
// If enforceNamespace is false, namespaces in the specs will be allowed to
// override the default namespace. If it is true, namespaces that don't match
// will cause an error.
// If ContinueOnError() is set prior to this method, objects on the path that are not
// recognized will be ignored (but logged at V(2)).
func (b *Builder) FilenameParam(filenameOptions *FilenameOptions) *Builder {
	if errs := filenameOptions.validate(); errs != nil {
		b.errs = errs
		return b
	}
	recursive := filenameOptions.Recursive
	paths := filenameOptions.Filenames
	for _, s := range paths {
		switch {
		case s == "-":
			b.Stdin()
		case strings.Index(s, "http://") == 0 || strings.Index(s, "https://") == 0:
			url, err := url.Parse(s)
			if err != nil {
				b.errs = errors.Join(b.errs, fmt.Errorf("the URL passed to filename %q is not valid: %v", s, err))
				continue
			}
			b.URL(defaultHttpGetAttempts, url)
		default:
			matches, err := expandIfFilePattern(s)
			if err != nil {
				b.errs = errors.Join(b.errs, err)
				continue
			}
			b.Path(recursive, matches...)
		}
	}
	return b
}

// expandIfFilePattern returns all the filenames that match the input pattern
// or the filename if it is a specific filename and not a pattern.
// If the input is a pattern and it yields no result it will result in an error.
func expandIfFilePattern(pattern string) ([]string, error) {
	if _, err := os.Stat(pattern); os.IsNotExist(err) {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) == 0 {
			return nil, fmt.Errorf(pathNotExistError, pattern)
		}
		if err == filepath.ErrBadPattern {
			return nil, fmt.Errorf("pattern %q is not valid: %v", pattern, err)
		}
		return matches, err
	}
	return []string{pattern}, nil
}

// ResourceTypeOrNameArgs indicates that the builder should accept arguments
// of the form `(<type1>[,<type2>,...]|<type> <name1>[,<name2>,...])`. When one argument is
// received, the types provided will be retrieved from the server (and be comma delimited).
// When two or more arguments are received, they must be a single type and resource name(s).
// The allowEmptySelector permits to select all the resources (via Everything func).
func (b *Builder) ResourceTypeOrNameArgs(args ...string) *Builder {
	if len(args) == 0 {
		return b
	}
	gr, err := getSchemaGroupResource(args[0])
	if err != nil {
		b.errs = errors.Join(b.errs, err)
		return b
	}
	b.gr = gr
	b.names = args[1:]
	return b
}

func getSchemaGroupResource(s string) (*schema.GroupResource, error) {
	parts := strings.SplitN(s, ".", 2)
	if len(parts) != 2 {
		return &schema.GroupResource{}, fmt.Errorf("expecting <resource>.<group>, got: %v", parts)
	}
	return &schema.GroupResource{Group: parts[1], Resource: parts[0]}, nil
}

// Stdin will read objects from the standard input. If ContinueOnError() is set
// prior to this method being called, objects in the stream that are unrecognized
// will be ignored (but logged at V(2)). If StdinInUse() is set prior to this method
// being called, an error will be recorded as there are multiple entities trying to use
// the single standard input stream.
func (b *Builder) Stdin() *Builder {
	b.stream = true
	if b.stdinInUse {
		b.errs = errors.Join(b.errs, StdinMultiUseError)
	}
	b.stdinInUse = true
	b.paths = append(b.paths, FileVisitorForSTDIN(b.mapper))
	return b
}

// Stream will read objects from the provided reader, and if an error occurs will
// include the name string in the error message. If ContinueOnError() is set
// prior to this method being called, objects in the stream that are unrecognized
// will be ignored (but logged at V(2)).
func (b *Builder) Stream(r io.Reader, name string) *Builder {
	b.stream = true
	b.paths = append(b.paths, NewStreamVisitor(r, b.mapper, name))
	return b
}

// Do returns a Result object with a Visitor for the resources identified by the Builder.
// The visitor will respect the error behavior specified by ContinueOnError. Note that stream
// inputs are consumed by the first execution - use Infos() or Object() on the Result to capture a list
// for further iteration.
func (b *Builder) Do() *Result {
	r := b.visitorResult()
	r.mapper = b.Mapper()
	if r.err != nil {
		return r
	}
	if b.flatten {
		r.visitor = NewFlattenListVisitor(r.visitor, b.objectTyper, b.mapper)
	}
	helpers := []VisitorFunc{}
	r.visitor = NewDecoratedVisitor(r.visitor, helpers...)
	return r
}

func (b *Builder) visitorResult() *Result {
	if b.errs != nil {
		return &Result{err: b.errs}
	}
	// visit items specified by paths
	if len(b.paths) != 0 {
		return b.visitByPaths()
	}
	if len(b.names) != 0 {
		return b.visitByName()
	}
	// TODO visit by resources
	// TODO visit by selector
	return &Result{err: missingResourceError}
}

func (b *Builder) visitByPaths() *Result {
	result := &Result{
		singleItemImplied:  !b.dir && !b.stream && len(b.paths) == 1,
		targetsSingleItems: true,
	}

	var visitors Visitor
	if b.continueOnError {
		visitors = EagerVisitorList(b.paths)
	} else {
		visitors = ConcurrentVisitorList{
			visitors:    b.paths,
			concurrency: b.visitorConcurrency,
		}
	}

	result.visitor = visitors
	result.sources = b.paths
	return result

}

func (b *Builder) visitByName() *Result {
	result := &Result{
		singleItemImplied:  len(b.names) == 1,
		targetsSingleItems: true,
	}

	if len(b.paths) != 0 {
		return result.withError(fmt.Errorf("when paths, URLs, or stdin is provided as input, you may not specify a resource by arguments as well"))
	}

	gvk, err := b.mapper.resourceMapper.KindFor(context.Background(), *b.gr, b.proxy, b.branch)
	if err != nil {
		return result.withError(fmt.Errorf("cannot get resource mapping for %s: %v", b.gr.String(), err))
	}

	namespace := "default"
	visitors := []Visitor{}
	for _, name := range b.names {
		u := &unstructured.Unstructured{}
		u.SetAPIVersion(gvk.GroupVersion().Identifier())
		u.SetKind(gvk.Kind)
		u.SetName(name)
		u.SetNamespace(namespace)
		info := &Info{
			Namespace: namespace,
			Name:      name,
			Object:    u,
		}
		visitors = append(visitors, info)
	}
	result.visitor = VisitorList(visitors)
	result.sources = visitors
	return result
}
