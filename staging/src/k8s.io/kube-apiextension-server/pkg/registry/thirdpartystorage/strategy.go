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

package thirdpartystorage

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
)

type ThirdPartyStorageStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator

	clusterScoped bool
}

func NewStrategy(typer runtime.ObjectTyper, clusterScoped bool) ThirdPartyStorageStrategy {
	return ThirdPartyStorageStrategy{typer, names.SimpleNameGenerator, clusterScoped}
}

func (a ThirdPartyStorageStrategy) NamespaceScoped() bool {
	return !a.clusterScoped
}

func (ThirdPartyStorageStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
}

func (ThirdPartyStorageStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
}

func (a ThirdPartyStorageStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("metadata"), nil, err.Error())}
	}

	return validation.ValidateObjectMetaAccessor(accessor, !a.clusterScoped, validation.NameIsDNSSubdomain, field.NewPath("metadata"))
}

func (ThirdPartyStorageStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (ThirdPartyStorageStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (ThirdPartyStorageStrategy) Canonicalize(obj runtime.Object) {
}

func (ThirdPartyStorageStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	objAccessor, err := meta.Accessor(obj)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("metadata"), nil, err.Error())}
	}
	oldAccessor, err := meta.Accessor(old)
	if err != nil {
		return field.ErrorList{field.Invalid(field.NewPath("metadata"), nil, err.Error())}
	}

	return validation.ValidateObjectMetaAccessorUpdate(objAccessor, oldAccessor, field.NewPath("metadata"))

	return field.ErrorList{}
}

func (a ThirdPartyStorageStrategy) GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, nil, err
	}
	return labels.Set(accessor.GetLabels()), objectMetaFieldsSet(accessor, !a.clusterScoped), nil
}

// objectMetaFieldsSet returns a fields that represent the ObjectMeta.
func objectMetaFieldsSet(objectMeta metav1.Object, clusterScoped bool) fields.Set {
	if clusterScoped {
		return fields.Set{
			"metadata.name": objectMeta.GetName(),
		}
	}
	return fields.Set{
		"metadata.name":      objectMeta.GetName(),
		"metadata.namespace": objectMeta.GetNamespace(),
	}
}

func (a ThirdPartyStorageStrategy) MatchThirdPartyStorage(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: a.GetAttrs,
	}
}
