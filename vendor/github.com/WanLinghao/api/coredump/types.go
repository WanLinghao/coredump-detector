package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

type CoredumpGetOptions struct {
	metav1.TypeMeta `json:",inline"`
	Container       string `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CoredumpEndpoint
// +k8s:openapi-gen=true
// +resource:path=coredumpendpoints,strategy=CoredumpEndpointStrategy
// +subresource:request=CoredumpEndpointDump,path=dump,kind=CoredumpEndpointDump
type CoredumpEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   CoredumpEndpointSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status CoredumpEndpointStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// CoredumpEndpointSpec defines the desired state of CoredumpEndpoint
type CoredumpEndpointSpec struct {
	PodUID types.UID `json:"podUID,omitempty" protobuf:"bytes,1,opt,name=podUID,casttype=k8s.io/apimachinery/pkg/types.UID"`
}

// CoredumpEndpointStatus defines the observed state of CoredumpEndpoint
type CoredumpEndpointStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +subresource-request
type CoredumpEndpointDump struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CoredumpEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []CoredumpEndpoint `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CoredumpEndpointDumpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []CoredumpEndpointDump `json:"items" protobuf:"bytes,2,rep,name=items"`
}
