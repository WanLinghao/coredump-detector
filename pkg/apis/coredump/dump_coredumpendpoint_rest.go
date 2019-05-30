
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



package coredump

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

var _ = rest.GetterWithOptions(&CoredumpEndpointDumpREST{})

// +k8s:deepcopy-gen=false
type CoredumpEndpointDumpREST struct {
	Registry CoredumpEndpointRegistry
}

// Get retrieves the object from the storage. It is required to support Patch.
func (r *CoredumpEndpointDumpREST) Get(ctx context.Context, name string, opts runtime.Object) (runtime.Object, error) {
	return nil, nil
}

func (r *CoredumpEndpointDumpREST) New() runtime.Object {
	return &CoredumpEndpointDump{}
}

// NewGetOptions creates a new options object
func (r *CoredumpEndpointDumpREST) NewGetOptions() (runtime.Object, bool, string) {
	return &CoredumpGetOptions{}, false, ""
}
