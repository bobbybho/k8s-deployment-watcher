package watcher

import (
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/bobbyho10/k8s-deployment-watcher/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

// PodWatcher ...
type PodWatcher struct {
	informerFactory informers.SharedInformerFactory
	podInformer     corev1.PodInformer
	queue           workqueue.RateLimitingInterface
}

// NewDeploymentWatcher ...
func NewPodWatcher(clientset kubernetes.Interface, namespace string, queue workqueue.RateLimitingInterface) *PodWatcher {
	pw := &PodWatcher{}

	pw.informerFactory = informers.NewSharedInformerFactory(clientset, time.Second*30)
	pw.informerFactory = informers.NewSharedInformerFactoryWithOptions(clientset, time.Second*30, informers.WithNamespace(namespace))
	pw.podInformer = pw.informerFactory.Core().V1().Pods()
	pw.queue = queue

	pw.podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    pw.podAdd,
		UpdateFunc: pw.podUpdate,
		DeleteFunc: pw.podDelete,
	})

	klog.Infof("New POD Watcher in namespace %v", namespace)

	return pw
}

// GetShareIndexInformer ...
func (n *PodWatcher) GetShareIndexInformer() cache.SharedIndexInformer {
	return n.podInformer.Informer()
}

// Run ...
func (n *PodWatcher) Run(stopCh chan struct{}) error {

	// Starts all the shared informers that have been created by the factory so
	// far.
	n.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, n.podInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to sync")
	}
	return nil
}

func (n *PodWatcher) podAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	klog.Infof("POD CREATED: %s/%s %v", pod.Name, pod.Namespace, pod.Labels)

	var event common.Event
	var err error
	event.Key, err = cache.MetaNamespaceKeyFunc(obj)
	event.EventType = "create"
	if err != nil {
		n.queue.Add(event)
	}
}

func (n *PodWatcher) podUpdate(old, new interface{}) {
	oldPod := old.(*v1.Pod)
	newPod := new.(*v1.Pod)
	klog.Infof(
		"POD UPDATED. %s/%s",
		oldPod.Namespace, oldPod.Name,
	)

	var event common.Event
	var err error
	event.Key, err = cache.MetaNamespaceKeyFunc(old)
	event.EventType = "update"
	if err != nil {
		n.queue.Add(event)
	}

	klog.Infof("%s", cmp.Diff(oldPod, newPod))
}

func (n *PodWatcher) podDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	klog.Infof("POD DELETED: %s/%s", pod.Name, pod.Namespace)

	var event common.Event
	var err error
	event.Key, err = cache.MetaNamespaceKeyFunc(obj)
	event.EventType = "update"
	if err != nil {
		n.queue.Add(event)
	}
}
