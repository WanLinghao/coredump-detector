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
	"os"
	"strings"
	"time"

	"github.com/WanLinghao/coredump-detector/pkg/backend/types"
	coredumpclientset "github.com/WanLinghao/coredump-detector/pkg/client/clientset_generated/clientset"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// This gc worker follows the logic:
// 1. If any namespace was deleted, it would clean its directory in backendstorage immediately
// 2. If any pod was deleted, it would log its deletion timestamp but not delete it immediately
// 3. Other gc worker like period_gc worker would clean those deleted pods core files after a while

type InformerGC struct {
	workers                    int
	backendStorage             types.Storage
	backendCleanQueue          workqueue.RateLimitingInterface
	coredumpEndpointCleanQueue workqueue.RateLimitingInterface
	gcThreshold                time.Duration
	kubeClient                 clientset.Interface
	coredumpEndpointClient     coredumpclientset.Interface
	kubeInformerFactory        kubeinformers.SharedInformerFactory
	podListerSynced            cache.InformerSynced
	nsListerSynced             cache.InformerSynced
}

func NewInformerGC(kubeClient clientset.Interface, backendStorage types.Storage, gt time.Duration) (*InformerGC, error) {
	ig := &InformerGC{
		workers:                    5,
		backendStorage:             backendStorage,
		backendCleanQueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "backend_cleaner"),
		coredumpEndpointCleanQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "coredumpendpoint_cleaner"),

		gcThreshold: gt,
		kubeClient:  kubeClient,
	}
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)

	cfg := config.GetConfigOrDie()
	cdeClient, err := coredumpclientset.NewForConfig(cfg)
	if err != nil {
		//clientLog.Error(err, "unable to set up client config")
		os.Exit(1)
	}
	ig.coredumpEndpointClient = cdeClient

	nsInformer := kubeInformerFactory.Core().V1().Namespaces()
	podInformer := kubeInformerFactory.Core().V1().Pods()

	ig.podListerSynced = podInformer.Informer().HasSynced
	ig.nsListerSynced = nsInformer.Informer().HasSynced

	nsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: ig.namespaceDeleted,
	})
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: ig.podDeleted,
	})
	ig.kubeInformerFactory = kubeInformerFactory
	return ig, nil
}

func (ig *InformerGC) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer ig.backendCleanQueue.ShutDown()
	defer ig.coredumpEndpointCleanQueue.ShutDown()

	klog.Infof("Starting core dump file inform cleaner")
	defer klog.Infof("Shutting down core dump inform cleaner")

	ig.kubeInformerFactory.Start(stopCh)
	if !WaitForCacheSync("core dump file cleaner", stopCh, ig.podListerSynced, ig.nsListerSynced) {
		return
	}

	for i := 0; i < ig.workers; i++ {
		go wait.Until(ig.runBackendWorker, time.Second, stopCh)
		go wait.Until(ig.runCoredumpEndpointWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (ig *InformerGC) runBackendWorker() {
	for ig.processBackendItem() {
	}
}

func (ig *InformerGC) runCoredumpEndpointWorker() {
	for ig.processCoredumpEndpointItem() {
	}
}

func (ig *InformerGC) processBackendItem() bool {
	key, quit := ig.backendCleanQueue.Get()
	if quit {
		return false
	}
	defer ig.backendCleanQueue.Done(key)

	if err := ig.syncBackend(key.(string)); err != nil {
		utilruntime.HandleError(fmt.Errorf("syncing %q failed: %v", key, err))
		ig.backendCleanQueue.AddRateLimited(key)
		return true
	}

	ig.backendCleanQueue.Forget(key)
	return true
}

func (ig *InformerGC) processCoredumpEndpointItem() bool {
	key, quit := ig.coredumpEndpointCleanQueue.Get()
	if quit {
		return false
	}
	defer ig.coredumpEndpointCleanQueue.Done(key)

	if err := ig.syncCoredumpEndpoint(key.(string)); err != nil {
		utilruntime.HandleError(fmt.Errorf("syncing %q failed: %v", key, err))
		ig.coredumpEndpointCleanQueue.AddRateLimited(key)
		return true
	}

	ig.coredumpEndpointCleanQueue.Forget(key)
	return true
}

func (ig *InformerGC) syncBackend(key string) error {
	currentTime := time.Now()
	defer func() {
		klog.Infof("Finished backend syncing for key %s (%v)", key, time.Since(currentTime))
	}()

	ns, podName, podUID, err := ig.splitKey(key)
	if err != nil {
		return err
	}

	if ns == "" {
		return fmt.Errorf("unexpected empty namespace")
	}

	// Backend would handle podUID:
	// 1. podUID is empty, it means delete all the core files of this namespace
	// 2. podUID is NOT empty, it means delete the core files generated by this one specific pod
	if podUID != "" {
		klog.Infof("Clean backend core files for pod %s/%s (%s)", ns, podName, podUID)
	} else {
		klog.Infof("Clean all backend core files in namespace %s", ns)
	}
	return ig.backendStorage.CleanCoreFiles(ns, podUID, "")
}

func (ig *InformerGC) syncCoredumpEndpoint(key string) error {
	currentTime := time.Now()
	defer func() {
		klog.Infof("Finished aggregation api layer syncing for key %s (%v)", key, time.Since(currentTime))
	}()

	ns, cdeName, podUID, err := ig.splitKey(key)
	if err != nil {
		return err
	}

	if ns == "" || cdeName == "" || podUID == "" {
		return fmt.Errorf("unexpected key in syncCoredumpEndpoint: %s", key)
	}

	// Clean coredumpendpoint
	currentCde, err := ig.coredumpEndpointClient.CoredumpV1alpha1().CoredumpEndpoints(ns).Get(cdeName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// The coredumpendpoint has been deleted, so we do nothing here.
			return nil
		} else {
			klog.Warningf("failed to check if coredumpendpoint %s/%s still exist before deletion: %v", ns, cdeName, err)
		}
	}

	if string(currentCde.Spec.PodUID) == podUID {
		// check the coredumpendpoint we are handling refers the same pod to the exist one in cluster
		klog.Infof("Clean coredumpendpoint object %s/%s refering pod uid %s", ns, cdeName, podUID)
		err = ig.coredumpEndpointClient.CoredumpV1alpha1().CoredumpEndpoints(ns).Delete(cdeName, &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete coredumpendpoint %s with pod uid %s in namespace %s: %v", cdeName, podUID, ns, err)
		}
	} else {
		klog.Infof("Give up cleaning coredumpendpint %s in namespace %s since the bound pod uid has changed from %s to %s",
			cdeName, ns, podUID, currentCde.Spec.PodUID)
	}

	return nil
}

func (ig *InformerGC) namespaceDeleted(obj interface{}) {
	ns := obj.(*v1.Namespace)
	if ns.DeletionTimestamp == nil {
		// double check
		return
	}

	ig.backendCleanQueue.Add(ig.generateKey(ns.Name, "", ""))
}

func (ig *InformerGC) podDeleted(obj interface{}) {
	pod := obj.(*v1.Pod)
	if pod.DeletionTimestamp == nil {
		// double check
		return
	}

	ns, err := ig.kubeClient.CoreV1().Namespaces().Get(pod.Namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// The namespace has been deleted, so we do nothing here.
			return
		} else {
			klog.Warningf("failed to check if namespace %s still exist before deletion: %v", pod.Namespace, err)
		}
	}

	if ns.DeletionTimestamp != nil {
		// Namespace has been deleted, the gc job has be done.
		return
	}

	currentTime := time.Now()
	deletionTime := currentTime.Add(ig.gcThreshold)
	klog.Infof("Detect pod deletion(%s/%s), execute gc job at %v (%s later)", pod.Namespace, pod.Name, deletionTime, ig.gcThreshold)
	ig.backendCleanQueue.AddAfter(ig.generateKey(pod.Namespace, pod.Name, string(pod.UID)), ig.gcThreshold)
	ig.coredumpEndpointCleanQueue.AddAfter(ig.generateKey(pod.Namespace, pod.Name, string(pod.UID)), ig.gcThreshold)
}

func (ig *InformerGC) generateKey(ns, name, podUID string) string {
	return ns + "/" + name + "/" + podUID
}

func (ig *InformerGC) splitKey(key string) (ns, name, podUID string, err error) {
	fileds := strings.Split(key, "/")
	if len(fileds) != 3 {
		return "", "", "", fmt.Errorf("unexpected key:%s", fileds)
	}
	return fileds[0], fileds[1], fileds[2], nil
}

// This function was copied from package k8s.io/kubernetes/pkg/controller
func WaitForCacheSync(controllerName string, stopCh <-chan struct{}, cacheSyncs ...cache.InformerSynced) bool {
	klog.Infof("Waiting for caches to sync for %s controller", controllerName)

	if !cache.WaitForCacheSync(stopCh, cacheSyncs...) {
		utilruntime.HandleError(fmt.Errorf("unable to sync caches for %s controller", controllerName))
		return false
	}

	klog.Infof("Caches are synced for %s controller", controllerName)
	return true
}
