/*
Copyright 2022.

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

package controllers

import (
	"context"

	apps "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatorv1 "demo.dw.io/operator/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type OwnerReferenceConfig struct {
	Obj    metav1.Object
	Scheme *runtime.Scheme
}

// DwOperatorReconciler reconciles a DwOperator object
type DwOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Recorder record.EventRecorder
}

func buildDeployment(dwOPerator operatorv1.DwOperator) *apps.Deployment {
	deployment := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dwOPerator.Spec.DeploymentName,
			Namespace: dwOPerator.Namespace,
			//OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&dwOPerator, operatorv1.GroupVersion.WithKind("DwOperator"))},
		},
		Spec: apps.DeploymentSpec{
			Replicas: &dwOPerator.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"demo.dw.io/deployment-name": dwOPerator.Spec.DeploymentName,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"demo.dw.io/deployment-name": dwOPerator.Spec.DeploymentName,
						"k8s-app":                    "dw-server",
						"availability-zone":          "",
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:    "dwdeployment",
							Image:   "bobbyho/dwserver:latest",
							Command: []string{"dwserver", "pod-controller", "watch-endpoints"},
							Env: []apiv1.EnvVar{
								{
									Name: "POD_IP",
									ValueFrom: &core.EnvVarSource{
										FieldRef: &core.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "status.podIP",
										},
									},
								},
								{
									Name: "POD_NAME",
									ValueFrom: &core.EnvVarSource{
										FieldRef: &core.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.name",
										},
									},
								},
								{
									Name: "NODE_NAME",
									ValueFrom: &core.EnvVarSource{
										FieldRef: &core.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
								{
									Name: "NAMESPACE",
									ValueFrom: &core.EnvVarSource{
										FieldRef: &core.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
								{
									Name: "AVAILABILITY_ZONE",
									ValueFrom: &core.EnvVarSource{
										FieldRef: &core.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.labels['availability-zone']",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return &deployment
}

//+kubebuilder:rbac:groups=operator.demo.dw.io,resources=dwoperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.demo.dw.io,resources=dwoperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.demo.dw.io,resources=dwoperators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DwOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *DwOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Fetching DwOperatorReconciler")
	dwOperator := operatorv1.DwOperator{}
	if err := r.Client.Get(ctx, req.NamespacedName, &dwOperator); err != nil {
		logger.Error(err, "failed to get DwOperator resource")
		// Ignore NotFound errors as they will be retried automatically if the
		// resource is created in future.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("checking if an existing Deployment exists for this resource")
	deployment := apps.Deployment{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: dwOperator.Namespace, Name: dwOperator.Spec.DeploymentName}, &deployment)
	if apierrors.IsNotFound(err) {
		logger.Info("could not find existing Deployment for dwOperator, creating one...")

		deployment = *buildDeployment(dwOperator)
		err = controllerutil.SetControllerReference(&dwOperator, &deployment, r.Scheme)
		if err != nil {
			logger.Error(err, "failed to set Deployment owner reference")
			return ctrl.Result{}, err
		}
		if err := r.Client.Create(ctx, &deployment); err != nil {
			logger.Error(err, "failed to create Deployment resource")
			return ctrl.Result{}, err
		}

		//r.Recorder.Eventf(&dwOperator, core.EventTypeNormal, "Created", "Created deployment %q", deployment.Name)
		logger.Info("created Deployment resource for DwOperator")
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "failed to get Deployment for DwOperator resource")
		return ctrl.Result{}, err
	}

	// update the proxy deployment
	expectedReplicas := int32(1)
	if dwOperator.Spec.Replicas != 0 {
		expectedReplicas = *&dwOperator.Spec.Replicas
	}
	if *deployment.Spec.Replicas != expectedReplicas {
		logger.Info("updating replica count", "old_count", *deployment.Spec.Replicas, "new_count", expectedReplicas)

		deployment.Spec.Replicas = &expectedReplicas
		if err := r.Client.Update(ctx, &deployment); err != nil {
			logger.Error(err, "failed to Deployment update replica count")
			return ctrl.Result{}, err
		}

		//r.Recorder.Eventf(&dwOperator, core.EventTypeNormal, "Scaled", "Scaled deployment %q to %d replicas", deployment.Name, expectedReplicas)

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DwOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1.DwOperator{}).
		Complete(r)
}
