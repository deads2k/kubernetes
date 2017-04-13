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
	"bytes"
	"fmt"
	"net/http"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/endpoints/handlers"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/client-go/discovery"

	listers "k8s.io/kube-apiextension-server/pkg/client/listers/apiextensions/internalversion"
	"k8s.io/kube-apiextension-server/pkg/registry/thirdpartystorage"
)

// apisHandler serves the `/apis` endpoint.
// This is registered as a filter so that it never collides with any explictly registered endpoints
type tprHandler struct {
	versionDiscoveryHandler *tprVersionDiscoveryHandler
	groupDiscoveryHandler   *tprGroupDiscoveryHandler

	requestContextMapper apirequest.RequestContextMapper

	thirdPartyLister listers.ThirdPartyLister

	delegate          http.Handler
	restOptionsGetter generic.RESTOptionsGetter
	admission         admission.Interface
}

func (r *tprHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, ok := r.requestContextMapper.Get(req)
	if !ok {
		http.Error(w, "missing context", http.StatusInternalServerError)
		return
	}
	requestInfo, ok := apirequest.RequestInfoFrom(ctx)
	if !ok {
		http.Error(w, "missing requestInfo", http.StatusInternalServerError)
		return
	}
	if !requestInfo.IsResourceRequest {
		pathParts := splitPath(requestInfo.Path)
		// only match /apis/<group>/<version>
		if len(pathParts) == 3 {
			r.versionDiscoveryHandler.ServeHTTP(w, req)
			return
		}
		// only match /apis/<group>
		if len(pathParts) == 2 {
			r.groupDiscoveryHandler.ServeHTTP(w, req)
			return
		}

		r.delegate.ServeHTTP(w, req)
		return
	}
	if len(requestInfo.Subresource) > 0 {
		http.NotFound(w, req)
		return
	}

	thirdPartyName := requestInfo.Resource + "." + requestInfo.APIGroup
	thirdParty, err := r.thirdPartyLister.Get(thirdPartyName)
	if apierrors.IsNotFound(err) {
		r.delegate.ServeHTTP(w, req)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if thirdParty.Spec.Version != requestInfo.APIVersion {
		r.delegate.ServeHTTP(w, req)
		return
	}

	storage := thirdpartystorage.NewREST(
		schema.GroupResource{Group: thirdParty.Spec.Group, Resource: thirdParty.Spec.Name},
		schema.GroupVersionKind{Group: thirdParty.Spec.Group, Version: thirdParty.Spec.Version, Kind: thirdParty.Spec.ListKind},
		UnstructuredCopier{},
		thirdpartystorage.NewStrategy(discovery.NewUnstructuredObjectTyper(nil), thirdParty.Spec.ClusterScoped),
		r.restOptionsGetter,
	)

	parameterScheme := runtime.NewScheme()
	parameterCodec := runtime.NewParameterCodec(parameterScheme)
	parameterScheme.AddUnversionedTypes(schema.GroupVersion{Group: thirdParty.Spec.Group, Version: thirdParty.Spec.Version},
		&metav1.ListOptions{},
		&metav1.ExportOptions{},
		&metav1.GetOptions{},
		&metav1.DeleteOptions{},
	)
	parameterScheme.AddGeneratedDeepCopyFuncs(metav1.GetGeneratedDeepCopyFuncs()...)

	requestScope := handlers.RequestScope{
		Namer: handlers.ContextBasedNaming{
			GetContext: func(req *http.Request) apirequest.Context {
				ret, _ := r.requestContextMapper.Get(req)
				return ret
			},
			SelfLinker:    meta.NewAccessor(),
			ClusterScoped: thirdParty.Spec.ClusterScoped,
		},
		ContextFunc: func(req *http.Request) apirequest.Context {
			ret, _ := r.requestContextMapper.Get(req)
			return ret
		},

		Serializer:     UnstructuredNegotiatedSerializer{},
		ParameterCodec: parameterCodec,

		Creater:         UnstructuredCreator{},
		Convertor:       unstructured.UnstructuredObjectConverter{},
		Defaulter:       UnstructuredDefaulter{},
		Copier:          UnstructuredCopier{},
		Typer:           discovery.NewUnstructuredObjectTyper(nil),
		UnsafeConvertor: unstructured.UnstructuredObjectConverter{},

		Resource:    schema.GroupVersionResource{Group: requestInfo.APIGroup, Version: requestInfo.APIVersion, Resource: requestInfo.Resource},
		Kind:        schema.GroupVersionKind{Group: requestInfo.APIGroup, Version: requestInfo.APIVersion, Kind: thirdParty.Spec.Kind},
		Subresource: "",

		MetaGroupVersion: metav1.SchemeGroupVersion,
	}

	minRequestTimeout := 1 * time.Minute

	switch requestInfo.Verb {
	case "get":
		handler := handlers.GetResource(storage, storage, requestScope)
		handler(w, req)
		return
	case "list":
		forceWatch := false
		handler := handlers.ListResource(storage, storage, requestScope, forceWatch, minRequestTimeout)
		handler(w, req)
		return
	case "watch":
		forceWatch := true
		handler := handlers.ListResource(storage, storage, requestScope, forceWatch, minRequestTimeout)
		handler(w, req)
		return
	case "create":
		handler := handlers.CreateResource(storage, requestScope, discovery.NewUnstructuredObjectTyper(nil), r.admission)
		handler(w, req)
		return
	case "update":
		handler := handlers.UpdateResource(storage, requestScope, discovery.NewUnstructuredObjectTyper(nil), r.admission)
		handler(w, req)
		return
	case "patch":
		handler := handlers.PatchResource(storage, requestScope, r.admission, unstructured.UnstructuredObjectConverter{})
		handler(w, req)
		return
	case "delete":
		allowsOptions := true
		handler := handlers.DeleteResource(storage, allowsOptions, requestScope, r.admission)
		handler(w, req)
		return

	default:
		http.Error(w, fmt.Sprintf("unhandled verb %q", requestInfo.Verb), http.StatusMethodNotAllowed)
		return
	}

	w.Write([]byte("GOT ME\n"))
	w.WriteHeader(200)
}

type UnstructuredNegotiatedSerializer struct{}

func (s UnstructuredNegotiatedSerializer) SupportedMediaTypes() []runtime.SerializerInfo {
	return []runtime.SerializerInfo{
		{
			MediaType:        "application/json",
			EncodesAsText:    true,
			Serializer:       json.NewSerializer(json.DefaultMetaFactory, UnstructuredCreator{}, discovery.NewUnstructuredObjectTyper(nil), false),
			PrettySerializer: json.NewSerializer(json.DefaultMetaFactory, UnstructuredCreator{}, discovery.NewUnstructuredObjectTyper(nil), true),
			StreamSerializer: &runtime.StreamSerializerInfo{
				EncodesAsText: true,
				Serializer:    json.NewSerializer(json.DefaultMetaFactory, UnstructuredCreator{}, discovery.NewUnstructuredObjectTyper(nil), false),
				Framer:        json.Framer,
			},
		},
	}
}

func (s UnstructuredNegotiatedSerializer) EncoderForVersion(serializer runtime.Encoder, gv runtime.GroupVersioner) runtime.Encoder {
	return unstructured.UnstructuredJSONScheme
}

func (s UnstructuredNegotiatedSerializer) DecoderToVersion(serializer runtime.Decoder, gv runtime.GroupVersioner) runtime.Decoder {
	return unstructured.UnstructuredJSONScheme
}

type UnstructuredCreator struct{}

func (UnstructuredCreator) New(kind schema.GroupVersionKind) (runtime.Object, error) {
	ret := &unstructured.Unstructured{}
	ret.SetGroupVersionKind(kind)
	return ret, nil
}

type UnstructuredCopier struct{}

func (UnstructuredCopier) Copy(obj runtime.Object) (runtime.Object, error) {
	// serialize and deserialize to ensure a clean copy
	buf := &bytes.Buffer{}
	err := unstructured.UnstructuredJSONScheme.Encode(obj, buf)
	if err != nil {
		return nil, err
	}
	out := &unstructured.Unstructured{}
	result, _, err := unstructured.UnstructuredJSONScheme.Decode(buf.Bytes(), nil, out)
	return result, err
}

type UnstructuredDefaulter struct{}

func (UnstructuredDefaulter) Default(in runtime.Object) {}

type TPRRESTOptionsGetter struct {
	StorageConfig           storagebackend.Config
	StoragePrefix           string
	EnableWatchCache        bool
	EnableGarbageCollection bool
	DeleteCollectionWorkers int
}

func (t TPRRESTOptionsGetter) GetRESTOptions(resource schema.GroupResource) (generic.RESTOptions, error) {
	ret := generic.RESTOptions{
		StorageConfig:           &t.StorageConfig,
		Decorator:               generic.UndecoratedStorage,
		EnableGarbageCollection: t.EnableGarbageCollection,
		DeleteCollectionWorkers: t.DeleteCollectionWorkers,
		ResourcePrefix:          t.StoragePrefix + "/" + resource.Group + "/" + resource.Resource,
	}
	if t.EnableWatchCache {
		ret.Decorator = genericregistry.StorageWithCacher
	}
	return ret, nil
}
