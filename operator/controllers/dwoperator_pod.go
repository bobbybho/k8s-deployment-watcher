package controllers

import (
	"context"
	"encoding/json"

	"demo.dw.io/operator/controllers/podstate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// DwPodReconciler reconciles a DwPod object
type DwPodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *DwPodReconciler) labelPodWithNodeAZ(ctx context.Context, pod *corev1.Pod) error {
	logger := log.FromContext(ctx)

	logger.Info("Updating Pod AZ label", "namespace", pod.Namespace, "target_pod", pod.Name, "labels", pod.Labels)

	scheduled, nodeName := podstate.IsPodScheduled(pod)
	if !scheduled {
		logger.Info("target_pod is not scheduled yet", "namespace", pod.Namespace, "target_pod", pod.Name, "labels", pod.Labels)
		return nil
	}
	node := &corev1.Node{}
	if err := r.Get(context.Background(), types.NamespacedName{Name: nodeName}, node); err != nil {
		return err
	}
	// Get the current Pod labels.
	podLabels := pod.Labels

	zone, exists := podLabels["availability-zone"]
	if exists && zone != "" {
		return nil
	}

	zone = node.Labels["topology.kubernetes.io/zone"]

	podLabels["availability-zone"] = zone

	logger.Info("Setting Pod AZ label from node labels", "namespace", pod.Namespace, "pod", pod.Name, "labels", podLabels)
	mergePatch, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": podLabels,
		},
	})
	if err != nil {
		return err
	}
	if err := r.Patch(context.Background(), pod, client.RawPatch(types.StrategicMergePatchType, mergePatch)); err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func (r *DwPodReconciler) reconcileDwPodLabel(ctx context.Context, pods *corev1.PodList) error {
	logger := log.FromContext(ctx)

	var err error
	for _, pod := range pods.Items {
		if err = r.labelPodWithNodeAZ(ctx, &pod); err != nil {
			logger.Error(err, "failed to add AZ label to Pod")
			return err
		}
	}
	return nil
}

func (r *DwPodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Fetching Pod Resource")
	/*
		pod := corev1.Pod{}
		if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
			logger.Error(err, "failed to get pod resource")
			// Ignore NotFound errors as they will be retried automatically if the
			// resource is created in future.
			if !apierrors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	*/

	// kubernetes uses level trigger so we want to reconcile all the targed Pod events in a single reoncile loop
	pods := &corev1.PodList{}
	listOps := []client.ListOption{
		client.InNamespace(req.Namespace),
		client.MatchingLabels{
			"k8s-app": "dw-server",
		},
	}

	if err := r.List(ctx, pods, listOps...); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileDwPodLabel(ctx, pods); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
/*
func (r *DwPodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
*/

func (r *DwPodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create a new Controller
	c, err := controller.New("dwpod-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	p := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {

			if _, ok := e.ObjectNew.GetLabels()["demo.dw.io/deployment-name"]; !ok {
				return false
			}
			return e.ObjectOld != e.ObjectNew
		},
		CreateFunc: func(e event.CreateEvent) bool {
			if _, ok := e.Object.GetLabels()["demo.dw.io/deployment-name"]; !ok {
				return false
			}
			return true
		},
	}

	// Watch for Pod events, and enqueue a reconcile.Request for namespace
	err = c.Watch(
		&source.Kind{Type: &corev1.Pod{}},
		handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Namespace: object.GetNamespace(),
					},
				},
			}
		}),
		p)
	if err != nil {
		return err
	}
	return nil
}
