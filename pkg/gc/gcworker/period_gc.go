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

package gcworker

import (
	"time"

	"github.com/WanLinghao/fujitsu-coredump/pkg/backend/types"
	"k8s.io/klog"
)

// This gc worker do gc periodically, it mainly invokes the GC() function in backendStorage

type periodGC struct {
	period         time.Duration
	backendStorage types.Storage
}

func NewPeriodGC(period time.Duration, backendStorage types.Storage) *periodGC {
	return &periodGC{
		period:         period,
		backendStorage: backendStorage,
	}
}

func (pgc *periodGC) Run(stopCh <-chan struct{}) {
	klog.Infof("Starting core dump file period cleaner")
	defer klog.Infof("Shutting down core dump file period cleaner")

	tick := time.NewTicker(pgc.period)
	for {
		select {
		case <-stopCh:
			return
		case <-tick.C:
			pgc.gc()
		}
	}
}

func (pgc *periodGC) gc() {
	err := pgc.backendStorage.GC()
	if err != nil {
		klog.Errorf("error happens when do period gc: %v", err)
	}
}
