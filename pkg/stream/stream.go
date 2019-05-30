/*
Copyright 2019 The Kubernetes Authors.

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

package stream

import (
	"context"
	"io"
	"bytes"
	"io/ioutil"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
)

// CoredumpStreamer is a resource that streams the contents of a particular
// location URL.
type CoredumpStreamer struct {
	Body string
}

// a CoredumpStreamer must implement a rest.ResourceStreamer
var _ rest.ResourceStreamer = &CoredumpStreamer{}

func (obj *CoredumpStreamer) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}
func (obj *CoredumpStreamer) DeepCopyObject() runtime.Object {
	panic("rest.LocationStreamer does not implement DeepCopyObject")
}

func (s *CoredumpStreamer) InputStream(ctx context.Context, apiVersion, acceptHeader string) (stream io.ReadCloser, flush bool, contentType string, err error) {
	stream = ioutil.NopCloser(bytes.NewReader([]byte(s.Body))) // r type is io.ReadCloser
	flush = true
	contentType = "text/plain"
	err = nil
	return
}
