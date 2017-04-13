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

package discovery

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/emicklei/go-restful"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/endpoints/handlers/negotiation"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
)

// APIResourceLister should be removed.  It exists because TPRs are currently hacked in by adding
// (and expecting failures) to an existing webservice, hoping successful matches mean success, and
// serving "fake" discovery information which doesn't match the version which was just added.
type APIResourceLister interface {
	ListAPIResources() []metav1.APIResource
}

// APIVersionDiscoveryHandler creates a webservice serving the supported resources for the version
// E.g., such a web service will be registered at /apis/extensions/v1beta1.
type APIVersionDiscoveryHandler struct {
	serializer runtime.NegotiatedSerializer

	groupVersion schema.GroupVersion

	apiResourceListerLock sync.Mutex
	apiResourceLister     atomic.Value
}

// TODO, we can remove this path once we eliminate the old TPRs.  No other flow requires it and removing it makes the code easier to read
// since we'd eliminate the entire concept of a resourcelister
func NewAPIVersionDiscoveryHandler(serializer runtime.NegotiatedSerializer, groupVersion schema.GroupVersion, apiResourceLister APIResourceLister) *APIVersionDiscoveryHandler {
	if keepUnversioned(groupVersion.Group) {
		// Because in release 1.1, /apis/extensions returns response with empty
		// APIVersion, we use stripVersionNegotiatedSerializer to keep the
		// response backwards compatible.
		serializer = stripVersionNegotiatedSerializer{serializer}
	}

	ret := &APIVersionDiscoveryHandler{
		serializer:   serializer,
		groupVersion: groupVersion,
	}
	ret.apiResourceLister.Store(apiResourceLister)
	return ret
}

func NewAPIVersionDiscoveryHandlerForResources(serializer runtime.NegotiatedSerializer, groupVersion schema.GroupVersion, resources []metav1.APIResource) *APIVersionDiscoveryHandler {
	if keepUnversioned(groupVersion.Group) {
		// Because in release 1.1, /apis/extensions returns response with empty
		// APIVersion, we use stripVersionNegotiatedSerializer to keep the
		// response backwards compatible.
		serializer = stripVersionNegotiatedSerializer{serializer}
	}

	ret := &APIVersionDiscoveryHandler{
		serializer:   serializer,
		groupVersion: groupVersion,
	}
	ret.apiResourceLister.Store(staticLister{resources})
	return ret
}

func (s *APIVersionDiscoveryHandler) AddResource(resource metav1.APIResource) {
	s.apiResourceListerLock.Lock()
	defer s.apiResourceListerLock.Unlock()

	lister := s.apiResourceLister.Load().(APIResourceLister)
	switch t := lister.(type) {
	case staticLister:
		t.list = append(t.list, resource)
	default:
		lister = unionLister{
			listers: []APIResourceLister{lister, staticLister{list: []metav1.APIResource{resource}}},
		}
	}

	s.apiResourceLister.Store(lister)
}

func (s *APIVersionDiscoveryHandler) AddToWebService(ws *restful.WebService) {
	mediaTypes, _ := negotiation.MediaTypesForSerializer(s.serializer)
	ws.Route(ws.GET("/").To(s.restfulHandle).
		Doc("get available resources").
		Operation("getAPIResources").
		Produces(mediaTypes...).
		Consumes(mediaTypes...).
		Writes(metav1.APIResourceList{}))
}

// restfulHandle returns a handler which will return the api.VersionAndVersion of the group.
func (s *APIVersionDiscoveryHandler) restfulHandle(req *restful.Request, resp *restful.Response) {
	s.ServeHTTP(resp.ResponseWriter, req.Request)
}

func (s *APIVersionDiscoveryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	lister := s.apiResourceLister.Load().(APIResourceLister)
	responsewriters.WriteObjectNegotiated(s.serializer, schema.GroupVersion{}, w, req, http.StatusOK,
		&metav1.APIResourceList{GroupVersion: s.groupVersion.String(), APIResources: lister.ListAPIResources()})
}

// staticLister implements the APIResourceLister interface for normal versions
type staticLister struct {
	list []metav1.APIResource
}

func (s staticLister) ListAPIResources() []metav1.APIResource {
	return s.list
}

// unionLister is uncommon.  It's slow an memory alloc-y, but it only happens for TPRs and even then
// it's only used during discovery which is comparatively infrequent.
type unionLister struct {
	listers []APIResourceLister
}

func (s unionLister) ListAPIResources() []metav1.APIResource {
	ret := []metav1.APIResource{}
	for _, lister := range s.listers {
		ret = append(ret, lister.ListAPIResources()...)
	}
	return ret
}
