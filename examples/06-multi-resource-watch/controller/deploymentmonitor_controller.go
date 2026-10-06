/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	watchesv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/06-multi-resource-watch/api/v1alpha1"
)

// DeploymentMonitorReconciler reconciles a DeploymentMonitor object
type DeploymentMonitorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=watches.examples.k8s.io,resources=deploymentmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=watches.examples.k8s.io,resources=deploymentmonitors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=watches.examples.k8s.io,resources=deploymentmonitors/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// This controller watches Deployments and their ReplicaSets to provide monitoring
// information about deployment status and replica set details.
func (r *DeploymentMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the DeploymentMonitor instance
	monitor := &watchesv1alpha1.DeploymentMonitor{}
	err := r.Get(ctx, req.NamespacedName, monitor)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("DeploymentMonitor resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get DeploymentMonitor")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling DeploymentMonitor",
		"name", monitor.Name,
		"namespace", monitor.Namespace,
		"targetDeployment", monitor.Spec.DeploymentName,
		"targetNamespace", monitor.Spec.Namespace)

	// Fetch the Deployment being monitored
	deployment := &appsv1.Deployment{}
	deploymentKey := types.NamespacedName{
		Name:      monitor.Spec.DeploymentName,
		Namespace: monitor.Spec.Namespace,
	}

	err = r.Get(ctx, deploymentKey, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Deployment not found",
				"deployment", monitor.Spec.DeploymentName,
				"namespace", monitor.Spec.Namespace)
			return r.updateStatusDeploymentNotFound(ctx, monitor)
		}
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Fetch ReplicaSets owned by the Deployment
	replicaSets := &appsv1.ReplicaSetList{}
	err = r.List(ctx, replicaSets,
		client.InNamespace(monitor.Spec.Namespace),
		client.MatchingLabels(deployment.Spec.Selector.MatchLabels))
	if err != nil {
		logger.Error(err, "Failed to list ReplicaSets")
		return ctrl.Result{}, err
	}

	// Filter ReplicaSets that are actually owned by this Deployment
	ownedReplicaSets := filterOwnedReplicaSets(deployment, replicaSets.Items)

	logger.Info("Found ReplicaSets for Deployment",
		"deployment", deployment.Name,
		"totalReplicaSets", len(ownedReplicaSets))

	// Update status with deployment and replicaset information
	if err := r.updateStatus(ctx, monitor, deployment, ownedReplicaSets); err != nil {
		logger.Error(err, "Failed to update DeploymentMonitor status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled DeploymentMonitor")
	return ctrl.Result{}, nil
}

// filterOwnedReplicaSets filters ReplicaSets that are owned by the given Deployment
func filterOwnedReplicaSets(deployment *appsv1.Deployment, replicaSets []appsv1.ReplicaSet) []appsv1.ReplicaSet {
	var owned []appsv1.ReplicaSet
	for _, rs := range replicaSets {
		for _, ownerRef := range rs.OwnerReferences {
			if ownerRef.UID == deployment.UID {
				owned = append(owned, rs)
				break
			}
		}
	}
	return owned
}

// updateStatus updates the DeploymentMonitor status with deployment and replicaset information
func (r *DeploymentMonitorReconciler) updateStatus(
	ctx context.Context,
	monitor *watchesv1alpha1.DeploymentMonitor,
	deployment *appsv1.Deployment,
	replicaSets []appsv1.ReplicaSet,
) error {
	// Build ReplicaSet info
	rsInfos := make([]watchesv1alpha1.ReplicaSetInfo, 0, len(replicaSets))
	for _, rs := range replicaSets {
		revision := ""
		if rs.Annotations != nil {
			revision = rs.Annotations["deployment.kubernetes.io/revision"]
		}

		rsInfo := watchesv1alpha1.ReplicaSetInfo{
			Name:              rs.Name,
			Replicas:          *rs.Spec.Replicas,
			ReadyReplicas:     rs.Status.ReadyReplicas,
			AvailableReplicas: rs.Status.AvailableReplicas,
			Revision:          revision,
		}
		rsInfos = append(rsInfos, rsInfo)
	}

	// Update status fields
	monitor.Status.DeploymentFound = true
	monitor.Status.DesiredReplicas = *deployment.Spec.Replicas
	monitor.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	monitor.Status.ReadyReplicas = deployment.Status.ReadyReplicas
	monitor.Status.ReplicaSets = rsInfos
	monitor.Status.TotalReplicaSets = len(replicaSets)
	monitor.Status.LastUpdated = metav1.Now()
	monitor.Status.ObservedGeneration = monitor.Generation

	// Set appropriate message
	if deployment.Status.AvailableReplicas == *deployment.Spec.Replicas {
		monitor.Status.Message = fmt.Sprintf("Deployment %s is ready with %d/%d replicas available",
			deployment.Name, deployment.Status.AvailableReplicas, *deployment.Spec.Replicas)
	} else {
		monitor.Status.Message = fmt.Sprintf("Deployment %s is not fully ready: %d/%d replicas available",
			deployment.Name, deployment.Status.AvailableReplicas, *deployment.Spec.Replicas)
	}

	// Update the status subresource
	if err := r.Status().Update(ctx, monitor); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// updateStatusDeploymentNotFound updates the status when the deployment is not found
func (r *DeploymentMonitorReconciler) updateStatusDeploymentNotFound(
	ctx context.Context,
	monitor *watchesv1alpha1.DeploymentMonitor,
) (ctrl.Result, error) {
	monitor.Status.DeploymentFound = false
	monitor.Status.DesiredReplicas = 0
	monitor.Status.AvailableReplicas = 0
	monitor.Status.ReadyReplicas = 0
	monitor.Status.ReplicaSets = nil
	monitor.Status.TotalReplicaSets = 0
	monitor.Status.LastUpdated = metav1.Now()
	monitor.Status.ObservedGeneration = monitor.Generation
	monitor.Status.Message = fmt.Sprintf("Deployment %s not found in namespace %s",
		monitor.Spec.DeploymentName, monitor.Spec.Namespace)

	if err := r.Status().Update(ctx, monitor); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// This example demonstrates multi-resource watching:
// - For() watches the primary resource (DeploymentMonitor)
// - Watches() watches Deployments with a custom event handler to trigger reconciliation
// - Watches() watches ReplicaSets with a custom event handler to trigger reconciliation
func (r *DeploymentMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&watchesv1alpha1.DeploymentMonitor{}).
		Watches(
			&appsv1.Deployment{},
			handler.EnqueueRequestsFromMapFunc(r.findDeploymentMonitorsForDeployment),
		).
		Watches(
			&appsv1.ReplicaSet{},
			handler.EnqueueRequestsFromMapFunc(r.findDeploymentMonitorsForReplicaSet),
		).
		Named("deploymentmonitor").
		Complete(r)
}

// findDeploymentMonitorsForDeployment finds all DeploymentMonitors that reference a given Deployment
func (r *DeploymentMonitorReconciler) findDeploymentMonitorsForDeployment(ctx context.Context, deployment client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	monitorList := &watchesv1alpha1.DeploymentMonitorList{}
	if err := r.List(ctx, monitorList); err != nil {
		logger.Error(err, "Failed to list DeploymentMonitors")
		return []reconcile.Request{}
	}

	var requests []reconcile.Request
	for _, monitor := range monitorList.Items {
		if monitor.Spec.DeploymentName == deployment.GetName() &&
			monitor.Spec.Namespace == deployment.GetNamespace() {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      monitor.Name,
					Namespace: monitor.Namespace,
				},
			})
			logger.Info("Enqueueing DeploymentMonitor due to Deployment change",
				"monitor", monitor.Name,
				"deployment", deployment.GetName())
		}
	}

	return requests
}

// findDeploymentMonitorsForReplicaSet finds all DeploymentMonitors that should be reconciled when a ReplicaSet changes
func (r *DeploymentMonitorReconciler) findDeploymentMonitorsForReplicaSet(ctx context.Context, replicaSet client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	// Find the Deployment that owns this ReplicaSet
	rs := replicaSet.(*appsv1.ReplicaSet)
	var deploymentName string
	for _, ownerRef := range rs.OwnerReferences {
		if ownerRef.Kind == "Deployment" {
			deploymentName = ownerRef.Name
			break
		}
	}

	if deploymentName == "" {
		// ReplicaSet is not owned by a Deployment
		return []reconcile.Request{}
	}

	// Find all DeploymentMonitors that monitor this Deployment
	monitorList := &watchesv1alpha1.DeploymentMonitorList{}
	if err := r.List(ctx, monitorList); err != nil {
		logger.Error(err, "Failed to list DeploymentMonitors")
		return []reconcile.Request{}
	}

	var requests []reconcile.Request
	for _, monitor := range monitorList.Items {
		if monitor.Spec.DeploymentName == deploymentName &&
			monitor.Spec.Namespace == replicaSet.GetNamespace() {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      monitor.Name,
					Namespace: monitor.Namespace,
				},
			})
			logger.Info("Enqueueing DeploymentMonitor due to ReplicaSet change",
				"monitor", monitor.Name,
				"replicaSet", replicaSet.GetName(),
				"deployment", deploymentName)
		}
	}

	return requests
}
