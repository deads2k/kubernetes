package kubernetes_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/davecgh/go-spew/spew"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	manualfake "k8s.io/client-go/rest/fake"
)

func TestListTimeout(t *testing.T) {
	fakeClient := &manualfake.RESTClient{
		GroupVersion:         appsv1.SchemeGroupVersion,
		NegotiatedSerializer: scheme.Codecs,
		Client: manualfake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			if req.URL.Query().Get("timeout") != "21s" {
				t.Fatal(spew.Sdump(req.URL.Query()))
			}
			return &http.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(&bytes.Buffer{})}, nil
		}),
	}
	clientConfig := &rest.Config{
		APIPath: "/apis",
		ContentConfig: rest.ContentConfig{
			NegotiatedSerializer: scheme.Codecs,
			GroupVersion:         &appsv1.SchemeGroupVersion,
		},
	}
	restClient, _ := rest.RESTClientFor(clientConfig)
	restClient.Client = fakeClient.Client
	realClient := kubernetes.New(restClient)

	timeout := int64(21)
	realClient.AppsV1().DaemonSets("").List(metav1.ListOptions{TimeoutSeconds: &timeout})
	realClient.AppsV1().DaemonSets("").Watch(metav1.ListOptions{TimeoutSeconds: &timeout})
}
