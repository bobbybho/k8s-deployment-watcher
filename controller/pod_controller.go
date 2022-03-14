package controller

import (
	"fmt"
	"sync"
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

	pb "github.com/bobbybho/k8s-deployment-watcher/proto"
)

// PodController ...
type PodController struct {
	controller
	PQ   map[string]chan pb.PodStatReply
	lock sync.RWMutex
}

func NewPodController(clientset kubernetes.Interface, namespace string) *PodController {
	pc := &PodController{}

	q := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), common.PodQueue)

	pw := watcher.NewPodWatcher(clientset, namespace, q)
	pc.informer = pw.GetShareIndexInformer()
	pc.queue = q

	pc.client = clientset

	pc.PQ = make(map[string]chan pb.PodStatReply)

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

	podStatReply := pb.PodStatReply{}
	podStatReply.Message = e.EventType
	podStatReply.Podstat = &pb.PodStat{
		Podstate: string(pod.Status.Phase),
		Podip:    pod.Status.PodIP,
		Nodename: pod.Status.NominatedNodeName,
		Podname:  pod.Name,
		Hostip:   pod.Status.HostIP,
	}

	pc.lock.RLock()
	defer pc.lock.RUnlock()

	for _, podStatusChan := range pc.PQ {
		podStatusChan <- podStatReply
	}

	return nil
}

// OpenChannel ...
func (pc *PodController) OpenChannel(clientID string) chan pb.PodStatReply {
	pc.lock.Lock()
	defer pc.lock.Unlock()

	if _, ok := pc.PQ[clientID]; !ok {
		pc.PQ[clientID] = make(chan pb.PodStatReply, 1)
	}

	return pc.PQ[clientID]
}

// OpenChannel ...
func (pc *PodController) CloseChannel(clientID string) {
	pc.lock.Lock()
	defer pc.lock.Unlock()

	if _, ok := pc.PQ[clientID]; ok {
		close(pc.PQ[clientID])
		delete(pc.PQ, clientID)
	}
}
