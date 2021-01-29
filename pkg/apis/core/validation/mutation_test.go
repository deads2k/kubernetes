package validation

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	"k8s.io/apimachinery/pkg/api/apitesting/validationtesting"
	metafuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/diff"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	coreapi "k8s.io/kubernetes/pkg/apis/core"
	corefuzzer "k8s.io/kubernetes/pkg/apis/core/fuzzer"
)

func getScheme() *runtime.Scheme {
	scheme := &runtime.Scheme{}
	utilruntime.Must(coreapi.AddToScheme(scheme))
	return scheme
}

func getValidators() *validationtesting.RuntimeObjectsValidator {
	validator := validationtesting.NewRuntimeObjectsValidator()

	validator.MustRegister(&coreapi.Node{}, false, ValidateNode, ValidateNodeUpdate)
	return validator
}

type validateUpdateCheck struct {
	obj    runtime.Object
	oldObj runtime.Object
}

func newClusterScopedObjectMeta(name string) metav1.ObjectMeta {
	seed := time.Now().UnixNano()
	fuzzer := fuzzer.FuzzerFor(metafuzzer.Funcs, rand.NewSource(seed), legacyscheme.Codecs)

	objMeta := metav1.ObjectMeta{}
	fuzzer.Fuzz(&objMeta)
	objMeta.Name = name
	objMeta.Namespace = ""
	// we don't validate these in this method
	objMeta.ManagedFields = nil

	return objMeta
}

func newClusterScopedObjectMetaUpdate(oldObjMeta metav1.ObjectMeta) metav1.ObjectMeta {
	oldObjMetaCopy := oldObjMeta.DeepCopy()
	seed := time.Now().UnixNano()
	fuzzer := fuzzer.FuzzerFor(metafuzzer.Funcs, rand.NewSource(seed), legacyscheme.Codecs)

	objMeta := metav1.ObjectMeta{}
	fuzzer.Fuzz(&objMeta)

	// many fields are immutable
	objMeta.Name = oldObjMetaCopy.Name
	objMeta.Namespace = oldObjMetaCopy.Namespace
	objMeta.UID = oldObjMetaCopy.UID
	objMeta.CreationTimestamp = oldObjMetaCopy.CreationTimestamp
	objMeta.DeletionTimestamp = oldObjMetaCopy.DeletionTimestamp
	objMeta.DeletionGracePeriodSeconds = oldObjMetaCopy.DeletionGracePeriodSeconds
	objMeta.ClusterName = oldObjMetaCopy.ClusterName
	objMeta.Generation = oldObjMetaCopy.Generation
	objMeta.ManagedFields = oldObjMetaCopy.ManagedFields

	return objMeta
}

func newNode(name string) *coreapi.Node {
	ret := &coreapi.Node{}
	seed := time.Now().UnixNano()
	fuzzer := fuzzer.FuzzerFor(
		fuzzer.MergeFuzzerFuncs(metafuzzer.Funcs, corefuzzer.Funcs),
		rand.NewSource(seed), legacyscheme.Codecs)
	fuzzer.Fuzz(ret)

	ret.ObjectMeta = newClusterScopedObjectMeta(name)

	return ret
}

func newNodeUpdate(node *coreapi.Node) *coreapi.Node {
	ret := &coreapi.Node{}
	seed := time.Now().UnixNano()
	fuzzer := fuzzer.FuzzerFor(
		fuzzer.MergeFuzzerFuncs(metafuzzer.Funcs, corefuzzer.Funcs),
		rand.NewSource(seed), legacyscheme.Codecs)
	fuzzer.Fuzz(ret)

	nodeCopy := node.DeepCopy()
	ret.ObjectMeta = newClusterScopedObjectMetaUpdate(nodeCopy.ObjectMeta)
	ret.Spec.ProviderID = nodeCopy.Spec.ProviderID
	ret.Spec.PodCIDRs = nodeCopy.Spec.PodCIDRs
	ret.Spec.ConfigSource = nodeCopy.Spec.ConfigSource
	ret.Spec.DoNotUseExternalID = nodeCopy.Spec.DoNotUseExternalID

	return ret
}

func newNodeValidationUpdateCheck() validateUpdateCheck {
	old := newNode("the-node")
	return validateUpdateCheck{
		obj:    newNodeUpdate(old),
		oldObj: old,
	}
}

// this test checks to see if validateUpdate mutates its arguments
func TestMutationValidateUpdate(t *testing.T) {
	validators := getValidators()

	// only test node now, but this is a proof of concept for overall enforcement
	mutationObjects := []validateUpdateCheck{
		newNodeValidationUpdateCheck(),
	}

	for _, tt := range mutationObjects {
		typeName := reflect.TypeOf(tt.obj)
		for i := 0; i < 20; i++ {
			t.Run(typeName.Name(), func(t *testing.T) {
				validator, exists := validators.GetInfo(tt.obj)
				if !exists {
					t.Fatal("missing validation func")
				}

				originalObj := tt.obj.DeepCopyObject()
				originalOldObj := tt.oldObj.DeepCopyObject()
				errors := validator.Validator.ValidateUpdate(tt.obj, tt.oldObj)
				if len(errors) > 0 {
					t.Fatal(errors)
				}
				if !reflect.DeepEqual(tt.oldObj, originalOldObj) {
					t.Errorf("mutation of oldObject:\n%v", diff.ObjectGoPrintDiff(originalOldObj, tt.oldObj))
				}
				if !reflect.DeepEqual(tt.obj, originalObj) {
					t.Errorf("mutation of oldObject:\n%v", diff.ObjectGoPrintDiff(originalObj, tt.obj))
				}
			})
		}
	}
}
