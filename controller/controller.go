package controller

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	client   kubernetes.Interface
	informer cache.SharedIndexInformer
	queue    workqueue.RateLimitingInterface
}
