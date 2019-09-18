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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

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
	PodUID types.UID `json:"poduid,omitempty" protobuf:"bytes,1,opt,name=poduid"`
}

// CoredumpEndpointStatus defines the observed state of CoredumpEndpoint
type CoredumpEndpointStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

type CoredumpGetOptions struct {
	metav1.TypeMeta `json:",inline"`
	Container       string `json:"container,omitempty" protobuf:"bytes,1,opt,name=container"`
}
