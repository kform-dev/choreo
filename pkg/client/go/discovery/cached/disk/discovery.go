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

package disk

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/henderiw/logger/log"
	choreov1alpha1 "github.com/kform-dev/choreo/apis/choreo/v1alpha1"
	"github.com/kform-dev/choreo/pkg/client/go/config"
	"github.com/kform-dev/choreo/pkg/client/go/discovery"
	"github.com/kform-dev/choreo/pkg/client/go/discovery/cached/memory"
	discoveryclient "github.com/kform-dev/choreo/pkg/client/go/discovery/client"
	"github.com/kform-dev/choreo/pkg/proto/discoverypb"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

type CachedDiscoveryClient struct {
	delegate discovery.DiscoveryInterface
	// cacheDirectory is the directory where discovery docs are held.  It must be unique per host:port combination to work well.
	cacheDirectory string

	// ttl is how long the cache should be considered valid
	ttl time.Duration

	// mutex protects the variables below
	m sync.Mutex

	// ourFiles are all filenames of cache files created by this process
	ourFiles sets.Set[string]
	// invalidated is true if all cache files should be ignored that are not ours (e.g. after Invalidate() was called)
	invalidated bool
}

var _ discovery.CachedDiscoveryInterface = &CachedDiscoveryClient{}

func NewCachedDiscoveryClient(config *config.Config, discoveryCacheDir string, ttl time.Duration) (*CachedDiscoveryClient, error) {
	discoveryClient, err := discoveryclient.NewDiscoveryClient(config)
	if err != nil {
		return nil, err
	}
	return newCachedDiscoveryClient(memory.NewMemCacheDiscoveryClient(discoveryClient), discoveryCacheDir, ttl), nil
}

// NewCachedDiscoveryClient creates a new DiscoveryClient.  cacheDirectory is the directory where discovery docs are held.  It must be unique per host:port combination to work well.
func newCachedDiscoveryClient(discoveryClient discovery.DiscoveryInterface, cacheDirectory string, ttl time.Duration) *CachedDiscoveryClient {
	return &CachedDiscoveryClient{
		delegate:       discoveryClient,
		cacheDirectory: cacheDirectory,
		ttl:            ttl,
		ourFiles:       sets.New[string](),
	}
}

func (r *CachedDiscoveryClient) Close() error {
	return r.delegate.Close()
}

func (r *CachedDiscoveryClient) Invalidate() {
	r.m.Lock()
	defer r.m.Unlock()

	r.ourFiles = sets.New[string]()
	r.invalidated = true
	if d, ok := r.delegate.(discovery.CachedDiscoveryInterface); ok {
		d.Invalidate()
	}
}

func (r *CachedDiscoveryClient) APIResources(ctx context.Context, branchName string) (*choreov1alpha1.APIResources, error) {
	log := log.FromContext(ctx)
	// REMOVED FOR NOW TO AVOID DISK CACHING
	/*
		filename := filepath.Join(r.cacheDirectory, "apiresources.json")
		cachedBytes, err := r.getCachedFile(filename)
		// don't fail on errors, since we can get the data from the apiserver
		if err == nil {
			apiResources, err := decode(cachedBytes)
			if err == nil {
				return apiResources, nil
			}
		}
	*/
	apiResources, err := r.delegate.APIResources(ctx, branchName)
	if err != nil {
		return nil, err
	}

	if apiResources == nil || len(apiResources.Spec.Groups) == 0 {
		// skip writing to cache
		log.Error("no apiresource retrieved from server")
		return apiResources, nil
	}
	// REMOVED FOR NOW TO AVOID DISK CACHING
	/*
		if err := r.writeCachedFile(filename, apiResources); err != nil {
			log.Error("eriting apiresources to cache failed", "error", err)
		}
	*/
	return apiResources, nil
}

func (r *CachedDiscoveryClient) Watch(ctx context.Context, req *discoverypb.Watch_Request) chan *discoverypb.Watch_Response {
	return r.delegate.Watch(ctx, req)
}

func (r *CachedDiscoveryClient) getCachedFile(filename string) ([]byte, error) {
	// after invalidation ignore cache files not created by this process
	r.m.Lock()
	if r.invalidated && !r.ourFiles.Has(filename) {
		r.m.Unlock()
		return nil, errors.New("cache invalidated")
	}
	r.m.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if time.Now().After(fileInfo.ModTime().Add(r.ttl)) {
		return nil, errors.New("cache expired")
	}

	// the cache is present and its valid.  Try to read and use it.
	cachedBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	r.m.Lock()
	defer r.m.Unlock()

	return cachedBytes, nil
}

func (r *CachedDiscoveryClient) writeCachedFile(filename string, obj runtime.Object) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0750); err != nil {
		return err
	}

	bytes, err := encode(obj)
	if err != nil {
		return err
	}

	f, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(bytes)
	if err != nil {
		return err
	}

	err = os.Chmod(f.Name(), 0660)
	if err != nil {
		return err
	}

	name := f.Name()
	err = f.Close()
	if err != nil {
		return err
	}

	// atomic rename
	r.m.Lock()
	defer r.m.Unlock()
	err = os.Rename(name, filename)
	if err == nil {
		r.ourFiles.Insert(filename)
	}
	return err
}

func decode(b []byte) (*choreov1alpha1.APIResources, error) {
	apiResources := &choreov1alpha1.APIResources{}
	if err := json.Unmarshal(b, apiResources); err != nil {
		return nil, err
	}
	return apiResources, nil
}

func encode(obj runtime.Object) ([]byte, error) {
	return json.Marshal(obj)

}
