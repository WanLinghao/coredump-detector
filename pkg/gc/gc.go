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

package gc

import (
	//	"github.com/WanLinghao/coredump-detector/pkg/k8sclient"
	//	clientset "k8s.io/client-go/kubernetes"
	"github.com/WanLinghao/coredump-detector/pkg/backend"
	"github.com/WanLinghao/coredump-detector/pkg/gc/gcworker"
	"github.com/WanLinghao/coredump-detector/pkg/gc/options"
	"github.com/WanLinghao/coredump-detector/pkg/k8sclient"
)

type GCWorker interface {
	Run(stopCh <-chan struct{})
}

type BackendGC struct {
	gcWorkers map[string]GCWorker
}

func NewBackendGC() (*BackendGC, error) {
	bs := backend.GetBackendStorage()
	workers := map[string]GCWorker{}
	informer, err := gcworker.NewInformerGC(k8sclient.GetClient(), bs, options.GCOpts.GCThreshold)
	if err != nil {
		return nil, err
	}

	workers["informer_gc_worker"] = informer
	//	workers["period_gc_worker"] = gcworker.NewPeriodGC(options.GCOpts.GCPeriod, bs)

	return &BackendGC{workers}, nil
}

func (bgc *BackendGC) GC(stopCh <-chan struct{}) {
	for _, worker := range bgc.gcWorkers {
		go worker.Run(stopCh)
	}
}
