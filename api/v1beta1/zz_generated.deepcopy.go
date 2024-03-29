//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RtlSdrReceiver) DeepCopyInto(out *RtlSdrReceiver) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RtlSdrReceiver.
func (in *RtlSdrReceiver) DeepCopy() *RtlSdrReceiver {
	if in == nil {
		return nil
	}
	out := new(RtlSdrReceiver)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RtlSdrReceiver) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RtlSdrReceiverList) DeepCopyInto(out *RtlSdrReceiverList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RtlSdrReceiver, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RtlSdrReceiverList.
func (in *RtlSdrReceiverList) DeepCopy() *RtlSdrReceiverList {
	if in == nil {
		return nil
	}
	out := new(RtlSdrReceiverList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RtlSdrReceiverList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RtlSdrReceiverSpec) DeepCopyInto(out *RtlSdrReceiverSpec) {
	*out = *in
	if in.Frequency != nil {
		in, out := &in.Frequency, &out.Frequency
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.ContainerPort != nil {
		in, out := &in.ContainerPort, &out.ContainerPort
		*out = new(v1.ContainerPort)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RtlSdrReceiverSpec.
func (in *RtlSdrReceiverSpec) DeepCopy() *RtlSdrReceiverSpec {
	if in == nil {
		return nil
	}
	out := new(RtlSdrReceiverSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RtlSdrReceiverStatus) DeepCopyInto(out *RtlSdrReceiverStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Pod != nil {
		in, out := &in.Pod, &out.Pod
		*out = new(v1.ObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RtlSdrReceiverStatus.
func (in *RtlSdrReceiverStatus) DeepCopy() *RtlSdrReceiverStatus {
	if in == nil {
		return nil
	}
	out := new(RtlSdrReceiverStatus)
	in.DeepCopyInto(out)
	return out
}
