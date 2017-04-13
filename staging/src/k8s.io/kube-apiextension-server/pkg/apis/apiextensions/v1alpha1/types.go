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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ThirdPartySpec describe how a user wants their resource to appear
type ThirdPartySpec struct {
	// Group is the group this resource belongs in
	Group string `json:"group" protobuf:"bytes,1,opt,name=group"`
	// Version is the version this resource belongs in
	Version string `json:"version" protobuf:"bytes,2,opt,name=version"`
	// Name is the plural name of the resource
	Name string `json:"name" protobuf:"bytes,3,opt,name=name"`
	// Singular is the singular name of the resource.  Defaults to lowercased <kind>
	Singular string `json:"singular,omitempty" protobuf:"bytes,4,opt,name=singular"`
	// ShortNames are short names for the resource.
	ShortNames []string `json:"shortNames,omitempty" protobuf:"bytes,5,opt,name=shortNames"`
	// Kind is the serialized kind of the resource
	Kind string `json:"kind" protobuf:"bytes,6,opt,name=kind"`
	// ListKind is the serialized kind of the list for this resource.  Defaults to <kind>List
	ListKind string `json:"listKind,omitempty" protobuf:"bytes,7,opt,name=listKind"`

	// ClusterScoped indicates that this resource is cluster scoped as opposed to namespace scoped
	ClusterScoped bool `json:"clusterScoped" protobuf:"bytes,8,opt,name=clusterScoped"`
}

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// ThirdPartyConditionType is a valid value for ThirdPartyCondition.Type
type ThirdPartyConditionType string

const (
	// NameConflict means the names chosen for this ThirdParty conflict with others in the group.
	NameConflict ThirdPartyConditionType = "NameConflict"
	// Terminating means that the ThirdParty has been deleted and is cleaning up.
	Terminating ThirdPartyConditionType = "Terminating"
)

// ThirdPartyCondition contains details for the current condition of this pod.
type ThirdPartyCondition struct {
	// Type is the type of the condition.
	Type ThirdPartyConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ThirdPartyConditionType"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// ThirdPartyStatus indicates the state of the ThirdParty
type ThirdPartyStatus struct {
	// Conditions indicate state for particular aspects of a ThirdParty
	Conditions []ThirdPartyCondition `json:"conditions" protobuf:"bytes,1,opt,name=conditions"`
}

// +genclient=true
// +nonNamespaced=true

// ThirdParty represents a resource that should be exposed on the API server.  Its name MUST be in the format
// <.spec.name>.<.spec.group>.
type ThirdParty struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec describes how the user wants the resources to appear
	Spec ThirdPartySpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	// Status indicates the actual state of the ThirdParty
	Status ThirdPartyStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ThirdPartyList is a list of ThirdParty objects.
type ThirdPartyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items individual ThirdParties
	Items []ThirdParty `json:"items" protobuf:"bytes,2,rep,name=items"`
}
