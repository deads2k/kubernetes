package resourcequota

import (
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/admission"
	resourcequotaapi "k8s.io/apiserver/pkg/admission/plugin/resourcequota/apis/resourcequota"
	v1 "k8s.io/apiserver/pkg/quota/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/quota/v1/evaluator/core"
	testing2 "k8s.io/utils/clock/testing"
	"k8s.io/utils/ptr"
	"reflect"
	"testing"
	"time"
)

func newTerminatingPod(namespace, name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: corev1.PodSpec{
			ActiveDeadlineSeconds: ptr.To[int64](10),
		},
		Status: corev1.PodStatus{},
	}
}

func newNonTerminatingPod(namespace, name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec:   corev1.PodSpec{},
		Status: corev1.PodStatus{},
	}
}

func newScopedQuota(namespace, name string, scope corev1.ResourceQuotaScope, used int64) *corev1.ResourceQuota {
	return &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourcePods: *resource.NewQuantity(2, resource.BinarySI),
			},
			Scopes: []corev1.ResourceQuotaScope{
				scope,
			},
		},
		Status: corev1.ResourceQuotaStatus{
			Hard: corev1.ResourceList{
				corev1.ResourcePods: *resource.NewQuantity(2, resource.BinarySI),
			},
			Used: corev1.ResourceList{
				corev1.ResourcePods: *resource.NewQuantity(used, resource.BinarySI),
			},
		},
	}
}

type testPodLister struct {
	pods []*corev1.Pod
}

// List will return all objects across namespaces
func (q *testPodLister) List(selector labels.Selector) (ret []runtime.Object, err error) {
	for i := range q.pods {
		ret = append(ret, q.pods[i])
	}
	return ret, nil
}

// Get will attempt to retrieve assuming that name==key
func (q *testPodLister) Get(name string) (runtime.Object, error) {
	for _, curr := range q.pods {
		if curr.Name == name {
			return curr, nil
		}
	}
	return nil, apierrors.NewNotFound(corev1.Resource("resourcequota"), name)
}

// ByNamespace will give you a GenericNamespaceLister for one namespace
func (q *testPodLister) ByNamespace(namespace string) cache.GenericNamespaceLister {
	namespacedPods := []*corev1.Pod{}
	for i, curr := range q.pods {
		if curr.Namespace == namespace {
			namespacedPods = append(namespacedPods, q.pods[i])
		}
	}
	return &testPodLister{namespacedPods}
}

func TestCheckRequest(t *testing.T) {
	firstTestStartingPodState := []*corev1.Pod{
		newTerminatingPod("foo", "one"),
	}
	podEvaluator := core.NewPodEvaluator(func(schema.GroupVersionResource) (cache.GenericLister, error) {
		return &testPodLister{firstTestStartingPodState}, nil
	}, testing2.NewFakeClock(time.Now()))

	type args struct {
		quotas    []corev1.ResourceQuota
		a         admission.Attributes
		evaluator v1.Evaluator
		limited   []resourcequotaapi.LimitedResource
	}
	tests := []struct {
		name    string
		args    args
		want    []corev1.ResourceQuota
		wantErr bool
	}{
		{
			name: "standard",
			args: args{
				quotas: []corev1.ResourceQuota{
					*newScopedQuota("foo", "terminating", corev1.ResourceQuotaScopeTerminating, 1),
					*newScopedQuota("foo", "nonterminating", corev1.ResourceQuotaScopeNotTerminating, 1),
				},
				a: admission.NewAttributesRecord(
					newTerminatingPod("foo", "one"),
					newNonTerminatingPod("foo", "one"),
					schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
					"foo",
					"one",
					schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
					"",
					admission.Update,
					nil,
					false,
					nil,
				),
				evaluator: podEvaluator,
				limited:   nil,
			},
			want: []corev1.ResourceQuota{
				*newScopedQuota("foo", "terminating", corev1.ResourceQuotaScopeTerminating, 2),
				*newScopedQuota("foo", "nonterminating", corev1.ResourceQuotaScopeTerminating, 0),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckRequest(tt.args.quotas, tt.args.a, tt.args.evaluator, tt.args.limited)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckRequest() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
