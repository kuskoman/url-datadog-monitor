//go:build !ignore_autogenerated

/*
Code generated by controller-gen. DO NOT EDIT.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificateStatus) DeepCopyInto(out *CertificateStatus) {
	*out = *in
	in.NotAfter.DeepCopyInto(&out.NotAfter)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificateStatus.
func (in *CertificateStatus) DeepCopy() *CertificateStatus {
	if in == nil {
		return nil
	}
	out := new(CertificateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *URLMonitor) DeepCopyInto(out *URLMonitor) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new URLMonitor.
func (in *URLMonitor) DeepCopy() *URLMonitor {
	if in == nil {
		return nil
	}
	out := new(URLMonitor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *URLMonitor) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *URLMonitorList) DeepCopyInto(out *URLMonitorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]URLMonitor, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new URLMonitorList.
func (in *URLMonitorList) DeepCopy() *URLMonitorList {
	if in == nil {
		return nil
	}
	out := new(URLMonitorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *URLMonitorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *URLMonitorSpec) DeepCopyInto(out *URLMonitorSpec) {
	*out = *in
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.CheckCert != nil {
		in, out := &in.CheckCert, &out.CheckCert
		*out = new(bool)
		**out = **in
	}
	if in.VerifyCert != nil {
		in, out := &in.VerifyCert, &out.VerifyCert
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new URLMonitorSpec.
func (in *URLMonitorSpec) DeepCopy() *URLMonitorSpec {
	if in == nil {
		return nil
	}
	out := new(URLMonitorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *URLMonitorStatus) DeepCopyInto(out *URLMonitorStatus) {
	*out = *in
	in.LastCheckTime.DeepCopyInto(&out.LastCheckTime)
	if in.Certificate != nil {
		in, out := &in.Certificate, &out.Certificate
		*out = new(CertificateStatus)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new URLMonitorStatus.
func (in *URLMonitorStatus) DeepCopy() *URLMonitorStatus {
	if in == nil {
		return nil
	}
	out := new(URLMonitorStatus)
	in.DeepCopyInto(out)
	return out
}
