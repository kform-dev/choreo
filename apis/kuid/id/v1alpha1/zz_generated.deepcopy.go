//go:build !ignore_autogenerated

/*
Copyright 2024 Nokia.

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

package v1alpha1

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdaptorID) DeepCopyInto(out *AdaptorID) {
	*out = *in
	out.NodeID = in.NodeID
	if in.ModuleBay != nil {
		in, out := &in.ModuleBay, &out.ModuleBay
		*out = new(int)
		**out = **in
	}
	if in.Module != nil {
		in, out := &in.Module, &out.Module
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdaptorID.
func (in *AdaptorID) DeepCopy() *AdaptorID {
	if in == nil {
		return nil
	}
	out := new(AdaptorID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterID) DeepCopyInto(out *ClusterID) {
	*out = *in
	out.SiteID = in.SiteID
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterID.
func (in *ClusterID) DeepCopy() *ClusterID {
	if in == nil {
		return nil
	}
	out := new(ClusterID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointID) DeepCopyInto(out *EndpointID) {
	*out = *in
	out.NodeID = in.NodeID
	if in.ModuleBay != nil {
		in, out := &in.ModuleBay, &out.ModuleBay
		*out = new(int)
		**out = **in
	}
	if in.Module != nil {
		in, out := &in.Module, &out.Module
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointID.
func (in *EndpointID) DeepCopy() *EndpointID {
	if in == nil {
		return nil
	}
	out := new(EndpointID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeID) DeepCopyInto(out *NodeID) {
	*out = *in
	out.SiteID = in.SiteID
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeID.
func (in *NodeID) DeepCopy() *NodeID {
	if in == nil {
		return nil
	}
	out := new(NodeID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PartitionAttachmentID) DeepCopyInto(out *PartitionAttachmentID) {
	*out = *in
	out.SiteID = in.SiteID
	if in.Cluster != nil {
		in, out := &in.Cluster, &out.Cluster
		*out = new(string)
		**out = **in
	}
	if in.Node != nil {
		in, out := &in.Node, &out.Node
		*out = new(string)
		**out = **in
	}
	if in.NodeSet != nil {
		in, out := &in.NodeSet, &out.NodeSet
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PartitionAttachmentID.
func (in *PartitionAttachmentID) DeepCopy() *PartitionAttachmentID {
	if in == nil {
		return nil
	}
	out := new(PartitionAttachmentID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PartitionClusterID) DeepCopyInto(out *PartitionClusterID) {
	*out = *in
	out.SiteID = in.SiteID
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PartitionClusterID.
func (in *PartitionClusterID) DeepCopy() *PartitionClusterID {
	if in == nil {
		return nil
	}
	out := new(PartitionClusterID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PartitionEndpointID) DeepCopyInto(out *PartitionEndpointID) {
	*out = *in
	out.NodeID = in.NodeID
	if in.ModuleBay != nil {
		in, out := &in.ModuleBay, &out.ModuleBay
		*out = new(int)
		**out = **in
	}
	if in.Module != nil {
		in, out := &in.Module, &out.Module
		*out = new(int)
		**out = **in
	}
	if in.Adaptor != nil {
		in, out := &in.Adaptor, &out.Adaptor
		*out = new(string)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PartitionEndpointID.
func (in *PartitionEndpointID) DeepCopy() *PartitionEndpointID {
	if in == nil {
		return nil
	}
	out := new(PartitionEndpointID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PartitionNodeID) DeepCopyInto(out *PartitionNodeID) {
	*out = *in
	out.SiteID = in.SiteID
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PartitionNodeID.
func (in *PartitionNodeID) DeepCopy() *PartitionNodeID {
	if in == nil {
		return nil
	}
	out := new(PartitionNodeID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PortID) DeepCopyInto(out *PortID) {
	*out = *in
	out.NodeID = in.NodeID
	if in.ModuleBay != nil {
		in, out := &in.ModuleBay, &out.ModuleBay
		*out = new(int)
		**out = **in
	}
	if in.Module != nil {
		in, out := &in.Module, &out.Module
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PortID.
func (in *PortID) DeepCopy() *PortID {
	if in == nil {
		return nil
	}
	out := new(PortID)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SiteID) DeepCopyInto(out *SiteID) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SiteID.
func (in *SiteID) DeepCopy() *SiteID {
	if in == nil {
		return nil
	}
	out := new(SiteID)
	in.DeepCopyInto(out)
	return out
}
