package controller

import (
	"fmt"
	"time"

	"github.com/bobbybho/k8s-deployment-watcher/common"
	"github.com/bobbybho/k8s-deployment-watcher/watcher"
	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

// PodController ...
type PodController struct {
	controller
}

func NewPodController(clientset kubernetes.Interface, namespace string) *PodController {
	pc := &PodController{}

	q := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), common.PodQueue)

	pw := watcher.NewPodWatcher(clientset, namespace, q)
	pc.informer = pw.GetShareIndexInformer()
	pc.queue = q

	pc.client = clientset
	return pc
}

func (pc *PodController) Run(stopper <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer pc.queue.ShutDown()

	klog.Infof("Starting PodController...")

	go pc.informer.Run(stopper)

	klog.Info("Synchronizing events...")

	//synchronize the cache before starting to process events
	if !cache.WaitForCacheSync(stopper, pc.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		klog.Info("synchronization failed...")
		return
	}

	klog.Info("Synchronizing completed")

	wait.Until(pc.runWorker, time.Second, stopper)
}

func (pc *PodController) runWorker() {
	for pc.processNextItem() {
		// continue looping
	}
}

func (pc *PodController) processNextItem() bool {
	item, stop := pc.queue.Get()

	if stop {
		return false
	}

	err := pc.processItem(item.(common.Event))
	if err == nil {
		pc.queue.Forget(item)
		return true
	}

	return true
}

func (pc *PodController) processItem(e common.Event) error {
	obj, _, err := pc.informer.GetIndexer().GetByKey(e.Key)
	if err != nil {
		return fmt.Errorf("failted to fetch object with key %s from store: %v", e.Key, err)
	}

	pod := obj.(*v1.Pod)

	klog.Infof("processed item %v for pod %v labels: %v", pod.Name, pod.Labels)

	return nil
}
