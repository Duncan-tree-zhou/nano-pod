/*
Copyright 2022.

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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NanoPodStatus defines the observed state of NanoPod
type NanoPodStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NanoPod is the Schema for the nanopods API
type NanoPod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NanoPodSpec   `json:"spec,omitempty"`
	Status NanoPodStatus `json:"status,omitempty"`
}

type NanoPodSpec struct {
	Template      v1.PodTemplateSpec `json:"template,omitempty" protobuf:"bytes,3,opt,name=template"`
	PatchStrategy PatchStrategy      `json:"patchStrategy,omitempty"`
}

// StrategyType is the strategy how NanoPod patch to target Pod.
// +enum
type PatchStrategy string

const (
	// AppendPatch means the NanoPod would path all it's attributes on matched Pod if it's not defined in matched Pod
	AppendPatch PatchStrategy = "AppendPatch"
	// OverWritePatch means the NanoPod would patch or overwrite all it's attributes on matched Pod.
	OverWritePatch PatchStrategy = "OverWritePatch"
)

//+kubebuilder:object:root=true

// NanoPodList contains a list of NanoPod
type NanoPodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NanoPod `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NanoPod{}, &NanoPodList{})
}
