package watcher

import (
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	appinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// DeploymentWatcher ...
type DeploymentWatcher struct {
	informerFactory    informers.SharedInformerFactory
	deploymentInformer appinformers.DeploymentInformer
}

// NewDeploymentWatcher ...
func NewDeploymentWatcher(clientset kubernetes.Interface, namespace string) *DeploymentWatcher {
	dw := &DeploymentWatcher{}

	dw.informerFactory = informers.NewSharedInformerFactory(clientset, time.Second*30)
	dw.informerFactory = informers.NewSharedInformerFactoryWithOptions(clientset, time.Second*30, informers.WithNamespace(namespace))
	dw.deploymentInformer = dw.informerFactory.Apps().V1().Deployments()

	dw.deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    dw.deploymentAdd,
		UpdateFunc: dw.deploymentUpdate,
		DeleteFunc: dw.deploymentDelete,
	})

	klog.Infof("New Deployment Watcher in namespace %v", namespace)

	return dw
}

// WatchDeploymentEndpoints ...
func (n *DeploymentWatcher) Run(stopCh chan struct{}) error {

	// Starts all the shared informers that have been created by the factory so
	// far.
	n.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, n.deploymentInformer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return nil
}

func (n *DeploymentWatcher) deploymentAdd(obj interface{}) {
	deployment := obj.(*appv1.Deployment)
	klog.Infof("DEPLOYMENT CREATED: %s/%s %v", deployment.Namespace, deployment.Name, deployment.Status.Replicas)
}

func (n *DeploymentWatcher) deploymentUpdate(old, new interface{}) {
	oldDeployment := old.(*appv1.Deployment)
	newDeployment := new.(*appv1.Deployment)
	klog.Infof(
		"DEPLOYMENT UPDATED. %s/%s (%v)%v",
		oldDeployment.Namespace, oldDeployment.Name, oldDeployment.Status.AvailableReplicas, newDeployment.Status.Replicas,
	)

	klog.Infof("%s", cmp.Diff(oldDeployment, newDeployment))
}

func (n *DeploymentWatcher) deploymentDelete(obj interface{}) {
	deployment := obj.(*appv1.Deployment)
	klog.Infof("DEPLOYMENT DELETED: %s/%s", deployment.Namespace, deployment.Name)
}
