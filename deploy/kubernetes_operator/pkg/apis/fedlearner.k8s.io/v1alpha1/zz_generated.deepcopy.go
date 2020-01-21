// +build !ignore_autogenerated

/* Copyright 2020 The FedLearner Authors. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
	sets "k8s.io/apimachinery/pkg/util/sets"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FLApp) DeepCopyInto(out *FLApp) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FLApp.
func (in *FLApp) DeepCopy() *FLApp {
	if in == nil {
		return nil
	}
	out := new(FLApp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FLApp) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FLAppList) DeepCopyInto(out *FLAppList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FLApp, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FLAppList.
func (in *FLAppList) DeepCopy() *FLAppList {
	if in == nil {
		return nil
	}
	out := new(FLAppList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FLAppList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FLAppSpec) DeepCopyInto(out *FLAppSpec) {
	*out = *in
	in.PS.DeepCopyInto(&out.PS)
	in.Master.DeepCopyInto(&out.Master)
	in.Worker.DeepCopyInto(&out.Worker)
	if in.FLReplicaSpecs != nil {
		in, out := &in.FLReplicaSpecs, &out.FLReplicaSpecs
		*out = make(map[FLReplicaType]*ReplicaSpec, len(*in))
		for key, val := range *in {
			var outVal *ReplicaSpec
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(ReplicaSpec)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FLAppSpec.
func (in *FLAppSpec) DeepCopy() *FLAppSpec {
	if in == nil {
		return nil
	}
	out := new(FLAppSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FLAppStatus) DeepCopyInto(out *FLAppStatus) {
	*out = *in
	if in.PSAddress != nil {
		in, out := &in.PSAddress, &out.PSAddress
		*out = make(sets.String, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PairStatus != nil {
		in, out := &in.PairStatus, &out.PairStatus
		*out = make(map[FLReplicaType]*Pair, len(*in))
		for key, val := range *in {
			var outVal *Pair
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Pair)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FLAppStatus.
func (in *FLAppStatus) DeepCopy() *FLAppStatus {
	if in == nil {
		return nil
	}
	out := new(FLAppStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MasterSpec) DeepCopyInto(out *MasterSpec) {
	*out = *in
	in.Template.DeepCopyInto(&out.Template)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MasterSpec.
func (in *MasterSpec) DeepCopy() *MasterSpec {
	if in == nil {
		return nil
	}
	out := new(MasterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PSSpec) DeepCopyInto(out *PSSpec) {
	*out = *in
	in.Template.DeepCopyInto(&out.Template)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PSSpec.
func (in *PSSpec) DeepCopy() *PSSpec {
	if in == nil {
		return nil
	}
	out := new(PSSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Pair) DeepCopyInto(out *Pair) {
	*out = *in
	if in.Local != nil {
		in, out := &in.Local, &out.Local
		*out = make(sets.String, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Remote != nil {
		in, out := &in.Remote, &out.Remote
		*out = make(sets.String, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Mapping != nil {
		in, out := &in.Mapping, &out.Mapping
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Pair.
func (in *Pair) DeepCopy() *Pair {
	if in == nil {
		return nil
	}
	out := new(Pair)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReplicaSpec) DeepCopyInto(out *ReplicaSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	in.Template.DeepCopyInto(&out.Template)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReplicaSpec.
func (in *ReplicaSpec) DeepCopy() *ReplicaSpec {
	if in == nil {
		return nil
	}
	out := new(ReplicaSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkerSpec) DeepCopyInto(out *WorkerSpec) {
	*out = *in
	in.Template.DeepCopyInto(&out.Template)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkerSpec.
func (in *WorkerSpec) DeepCopy() *WorkerSpec {
	if in == nil {
		return nil
	}
	out := new(WorkerSpec)
	in.DeepCopyInto(out)
	return out
}
