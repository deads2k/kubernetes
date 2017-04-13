/*
Copyright 2017 The Kubernetes Authors.

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

package apiserver

import (
	"net/http"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/discovery"
)

type tprVersionDiscoveryHandler struct {
	// TODO, writing is infrequent, optimize this
	discoveryLock sync.RWMutex
	discovery     map[schema.GroupVersion]*discovery.APIVersionDiscoveryHandler

	delegate http.Handler
}

func (r *tprVersionDiscoveryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pathParts := splitPath(req.URL.Path)
	// only match /apis/<group>/<version>
	if len(pathParts) != 3 {
		r.delegate.ServeHTTP(w, req)
		return
	}
	discovery, ok := r.getDiscovery(schema.GroupVersion{Group: pathParts[1], Version: pathParts[2]})
	if !ok {
		r.delegate.ServeHTTP(w, req)
		return
	}

	discovery.ServeHTTP(w, req)
}

func (r *tprVersionDiscoveryHandler) getDiscovery(version schema.GroupVersion) (*discovery.APIVersionDiscoveryHandler, bool) {
	r.discoveryLock.RLock()
	defer r.discoveryLock.RUnlock()

	ret, ok := r.discovery[version]
	return ret, ok
}

func (r *tprVersionDiscoveryHandler) setDiscovery(version schema.GroupVersion, discovery *discovery.APIVersionDiscoveryHandler) {
	r.discoveryLock.Lock()
	defer r.discoveryLock.Unlock()

	r.discovery[version] = discovery
}

func (r *tprVersionDiscoveryHandler) unsetDiscovery(version schema.GroupVersion) {
	r.discoveryLock.Lock()
	defer r.discoveryLock.Unlock()

	delete(r.discovery, version)
}

type tprGroupDiscoveryHandler struct {
	// TODO, writing is infrequent, optimize this
	discoveryLock sync.RWMutex
	discovery     map[string]*discovery.APIGroupDiscoveryHandler

	delegate http.Handler
}

func (r *tprGroupDiscoveryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pathParts := splitPath(req.URL.Path)
	// only match /apis/<group>
	if len(pathParts) != 2 {
		r.delegate.ServeHTTP(w, req)
		return
	}
	discovery, ok := r.getDiscovery(pathParts[1])
	if !ok {
		r.delegate.ServeHTTP(w, req)
		return
	}

	discovery.ServeHTTP(w, req)
}

func (r *tprGroupDiscoveryHandler) getDiscovery(group string) (*discovery.APIGroupDiscoveryHandler, bool) {
	r.discoveryLock.RLock()
	defer r.discoveryLock.RUnlock()

	ret, ok := r.discovery[group]
	return ret, ok
}

func (r *tprGroupDiscoveryHandler) setDiscovery(group string, discovery *discovery.APIGroupDiscoveryHandler) {
	r.discoveryLock.Lock()
	defer r.discoveryLock.Unlock()

	r.discovery[group] = discovery
}

func (r *tprGroupDiscoveryHandler) unsetDiscovery(group string) {
	r.discoveryLock.Lock()
	defer r.discoveryLock.Unlock()

	delete(r.discovery, group)
}

// splitPath returns the segments for a URL path.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}
