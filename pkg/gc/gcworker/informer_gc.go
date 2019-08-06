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
	"fmt"
	"strings"
	"time"

	//	"github.com/WanLinghao/fujitsu-coredump/pkg/k8sclient"
	"github.com/WanLinghao/fujitsu-coredump/pkg/backend/types"
	"k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	//"k8s.io/kubernetes/pkg/util/metrics"
)

// This gc worker follows the logic:
// 1. If any namespace was deleted, it would clean its directory in backendstorage immediately
// 2. If any pod was deleted, it would log its deletion timestamp but not delete it immediately
// 3. Other gc worker like period_gc worker would clean those deleted pods core files after a while

type InformerGC struct {
	//kubeClient clientset.Interface
	workers        int
	backendStorage types.Storage
	queue          workqueue.RateLimitingInterface
	gcThreshold    time.Duration
}

func NewInformerGC(kubeClient clientset.Interface, backendStorage types.Storage, gt time.Duration) (*InformerGC, error) {
	ig := &InformerGC{
		//kubeClient : kubeClient,
		workers:        5,
		backendStorage: backendStorage,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "core_file_cleaner"),
		gcThreshold:    gt,
	}
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)

	nsInformer := kubeInformerFactory.Core().V1().Namespaces()
	podInformer := kubeInformerFactory.Core().V1().Pods()
	// if kubeClient.CoreV1().RESTClient().GetRateLimiter() != nil {
	// 	if err := metrics.RegisterMetricAndTrackRateLimiterUsage("core_file_cleaner", kubeClient.CoreV1().RESTClient().GetRateLimiter()); err != nil {
	// 		return nil, err
	// 	}
	// }
	nsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: ig.namespaceDeleted,
	})
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: ig.podDeleted,
	})

	return ig, nil
}

func (ig *InformerGC) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer ig.queue.ShutDown()

	klog.Infof("Starting core dump file cleaner")
	defer klog.Infof("Shutting down core dump file cleaner")

	// if !controller.WaitForCacheSync("core dump file cleaner", stopCh, c.cmListerSynced) {
	// 	return
	// }

	for i := 0; i < ig.workers; i++ {
		go wait.Until(ig.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (ig *InformerGC) runWorker() {
	for ig.processNextWorkItem() {
	}
}

func (ig *InformerGC) processNextWorkItem() bool {
	key, quit := ig.queue.Get()
	if quit {
		return false
	}
	defer ig.queue.Done(key)

	if err := ig.syncHandler(key.(string)); err != nil {
		utilruntime.HandleError(fmt.Errorf("syncing %q failed: %v", key, err))
		ig.queue.AddRateLimited(key)
		return true
	}

	ig.queue.Forget(key)
	return true
}

func (ig *InformerGC) syncHandler(key string) error {
	currentTime := time.Now()
	defer func() {
		klog.V(4).Infof("Finished syncing for key %s (%v)", key, time.Since(currentTime))
	}()

	ns, podUID, err := ig.splitKey(key)
	if err != nil {
		return err
	}

	if ns == "" {
		return fmt.Errorf("unexpected empty namespace")
	} else if podUID == "" {
		// indicates a namespace was deleted, and we should clean related core files
		return ig.backendStorage.CleanNamespace(ns)
	} else {
		// indicates a pod was deleted, and we should log its deletion time and clean its core files after a while
		return ig.backendStorage.LogPodDeletion(ns, podUID, currentTime.Add(ig.gcThreshold))
	}
}

func (ig *InformerGC) namespaceDeleted(obj interface{}) {
	ns := obj.(*v1.Namespace)
	if ns.DeletionTimestamp == nil {
		// double check
		return
	}
	ig.queue.Add(ig.generateKey(ns.Name, ""))
}

func (ig *InformerGC) podDeleted(obj interface{}) {
	pod := obj.(*v1.Pod)
	if pod.DeletionTimestamp == nil {
		// double check
		return
	}
	ig.queue.Add(ig.generateKey(pod.Namespace, pod.Name))
}

func (ig *InformerGC) generateKey(ns, pod string) string {
	return ns + "/" + pod
}

func (ig *InformerGC) splitKey(key string) (string, string, error) {
	fileds := strings.Split(key, "/")
	if len(fileds) != 2 {
		return "", "", fmt.Errorf("unexpected key:%s", fileds)
	}
	return fileds[0], fileds[1], nil
}
