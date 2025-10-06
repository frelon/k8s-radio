/*
Copyright 2024.

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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PodFailedReason = "PodFailed"
	ReadyCondition  = "Ready"
)

// RtlSdrReceiverSpec defines the desired state of RtlSdrReceiver
type RtlSdrReceiverSpec struct {
	// +kubebuilder:validation:Default=v4
	Version RtlSdrVersion `json:"version"`

	// Frequency is the radio frequency to tune the receiver to.
	// +kubebuilder:example="101.9M"
	// +optional
	Frequency *resource.Quantity `json:"frequency"`

	// ContainerPort contains the port settings for the Pod.
	// +optional
	ContainerPort *corev1.ContainerPort `json:"port"`
}

// RtlSdrVersion is the major version of the rtl-sdr receiver.
// +kubebuilder:validation:Enum=v3;v4
type RtlSdrVersion string

const (
	V3 RtlSdrVersion = "v3"
	V4 RtlSdrVersion = "v4"
)

// RtlSdrReceiverStatus defines the observed state of RtlSdrReceiver
type RtlSdrReceiverStatus struct {
	// Conditions describe the state of the receiver.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// State describes the current state of the receiver.
	// +optional
	State RtlSdrReceiverState `json:"state"`

	// Pod is a reference to the underlying pod.
	// +optional
	Pod *corev1.ObjectReference `json:"pod,omitempty"`
}

// RtlSdrReceiverState state of the rtl-sdr receiver.
// +kubebuilder:validation:Enum=Waiting;Running;Failed
type RtlSdrReceiverState string

const (
	StateWaiting RtlSdrReceiverState = "Waiting"
	StateRunning RtlSdrReceiverState = "Running"
	StateFailed  RtlSdrReceiverState = "Failed"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RtlSdrReceiver is the Schema for the rtlsdrreceivers API
type RtlSdrReceiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RtlSdrReceiverSpec   `json:"spec,omitempty"`
	Status RtlSdrReceiverStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RtlSdrReceiverList contains a list of RtlSdrReceiver
type RtlSdrReceiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RtlSdrReceiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RtlSdrReceiver{}, &RtlSdrReceiverList{})
}
