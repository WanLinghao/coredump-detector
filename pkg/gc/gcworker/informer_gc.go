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

	"github.com/WanLinghao/fujitsu-coredump/pkg/backend/types"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	coredumpclientset "github.com/wanlinghao/fujitsu-coredump/pkg/client/clientset_generated"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/api/errors"
)

// This gc worker follows the logic:
// 1. If any namespace was deleted, it would clean its directory in backendstorage immediately
// 2. If any pod was deleted, it would log its deletion timestamp but not delete it immediately
// 3. Other gc worker like period_gc worker would clean those deleted pods core files after a while

type InformerGC struct {
	workers             int
	backendStorage      types.Storage
	backendCleanQueue               workqueue.RateLimitingInterface
	coredumpEndpointCleanQueue      workqueue.RateLimitingInterface
	gcThreshold         time.Duration
	kubeClient clientset.Interface
	coredumpEndpointClient coredumpclientset.Interface
	kubeInformerFactory kubeinformers.SharedInformerFactory
	podListerSynced     cache.InformerSynced
	nsListerSynced      cache.InformerSynced
}

func NewInformerGC(kubeClient clientset.Interface, backendStorage types.Storage, gt time.Duration) (*InformerGC, error) {
	ig := &InformerGC{
		workers:        5,
		backendStorage: backendStorage,
		backendCleanQueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "backend_cleaner"),
		coredumpEndpointCleanQueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "coredumpendpoint_cleaner"),

		gcThreshold:    gt,
		kubeClient:     kubeClient,
	}
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)

	cfg := config.GetConfigOrDie()
	cdeClient, err := coredumpclientset.NewForConfig(cfg)
	if err != nil {
		clientLog.Error(err, "unable to set up client config")
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

	if !WaitForCacheSync("core dump file cleaner", stopCh, ig.podListerSynced, ig.nsListerSynced) {
		return
	}

	ig.kubeInformerFactory.Start(stopCh)

	for i := 0; i < ig.workers; i++ {
		go wait.Until(ig.runWorker, time.Second, stopCh)
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

func (ig *InformerGC) processCoredumpEndpointItem() {
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
		klog.Infof("Finished syncing for key %s (%v)", key, time.Since(currentTime))
	}()

	ns, _, podUID := ig.splitKey(key)
	if err != nil {
		return err
	}

	if ns == "" {
		return fmt.Errorf("unexpected empty namespace")
	}

	// Backend would handle podUID:
	// 1. podUID is empty, it means delete all the core files of this namespace
	// 2. podUID is NOT empty, it means delete the core files generated by this one specific pod
	return ig.backendStorage.CleanCoreFiles(ns, podUID, "")
}

func (ig *InformerGC) syncCoredumpEndpoint(key string) error {
	currentTime := time.Now()
	defer func() {
		klog.Infof("Finished syncing for key %s (%v)", key, time.Since(currentTime))
	}()

	ns, cdeName, podUID, err := ig.splitKey(key)
	if err != nil {
		return err
	}
	if ns == "" {
		return fmt.Errorf("unexpected empty namespace")
	}

	allErrors := []error{}
	if cdeName != "" {
		// Clean one coredumpendpoint
		currentCde, err := ig.coredumpEndpointClient.CoredumpV1alpha1().CoredumpEndpoints(ns).Get(cdeName)
		if err != nil {
			if !errors.IsNotFound(err) {
				allErrors = append(allErrors, err)
			}
		}

		if string(currentCde.Spec.PodUID) == podUID {
			// check the coredumpendpoint we are handling refers the same pod to the exist one in cluster
			klog.Infof("Clean useless coredumpendpoint object %s/%s refering pod uid %s", ns, cdeName, podUID)
			err = ig.coredumpEndpointClient.CoredumpV1alpha1().CoredumpEndpoints(ns).Delete(cdeName, &metav1.DeleteOptions{})
			if err != nil {
				allErrors = append(allErrors, err)
			}
		}
	} else {
		// Clean whole namespace
		// TODO: this should be done automatically by apiserver
		err = ig.coredumpEndpointClient.CoredumpV1alpha1().CoredumpEndpoints(ns).Delete(nil, metav1.DeleteOptions{})
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}
	return utilerrors.NewAggregate(allErrors)
}

func (ig *InformerGC) namespaceDeleted(obj interface{}) {
	ns := obj.(*v1.Namespace)
	if ns.DeletionTimestamp == nil {
		// double check
		return
	}
	ig.backendCleanQueue.Add(ig.generateKey(ns.Name, "", ""))
	ig.coredumpEndpointCleanQueue.Add(ig.generateKey(ns.Name, "", ""))
}

func (ig *InformerGC) podDeleted(obj interface{}) {
	pod := obj.(*v1.Pod)
	if pod.DeletionTimestamp == nil {
		// double check
		return
	}
	ig.backendCleanQueue.AddAfter(ig.generateKey(pod.Namespace, pod.Name, string(pod.UID)), ig.gcThreshold)
	ig.coredumpEndpointCleanQueue.AddAfter(ig.generateKey(pod.Namespace, pod.Name, string(pod.UID)), ig.gcThreshold)
}

func (ig *InformerGC) generateKey(ns, name, podUID string) string {
	return ns + "/" + podUID
}

func (ig *InformerGC) splitKey(key string) (ns, name, podUID string, error) {
	fileds := strings.Split(key, "/")
	if len(fileds) != 3 {
		return "", "", fmt.Errorf("unexpected key:%s", fileds)
	}
	return fileds[0], fileds[1], fileds[1], nil
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
