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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	statusv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/05-status-conditions/api/v1alpha1"
)

const (
	// Condition types
	TypeAvailable   = "Available"
	TypeProgressing = "Progressing"
	TypeDegraded    = "Degraded"
)

// ManagedServiceReconciler reconciles a ManagedService object
type ManagedServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=status.examples.k8s.io,resources=managedservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=status.examples.k8s.io,resources=managedservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=status.examples.k8s.io,resources=managedservices/finalizers,verbs=update

func (r *ManagedServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the ManagedService instance
	service := &statusv1alpha1.ManagedService{}
	err := r.Get(ctx, req.NamespacedName, service)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ManagedService resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get ManagedService")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling ManagedService",
		"name", service.Name,
		"namespace", service.Namespace,
		"serviceType", service.Spec.ServiceType,
		"version", service.Spec.Version)

	// Set Progressing condition
	meta.SetStatusCondition(&service.Status.Conditions, metav1.Condition{
		Type:               TypeProgressing,
		Status:             metav1.ConditionTrue,
		Reason:             "Reconciling",
		Message:            "Starting reconciliation",
		ObservedGeneration: service.Generation,
	})

	service.Status.Phase = "Provisioning"
	if err := r.Status().Update(ctx, service); err != nil {
		logger.Error(err, "Failed to update ManagedService status")
		return ctrl.Result{}, err
	}

	// Simulate provisioning work
	if err := r.provisionService(ctx, service); err != nil {
		logger.Error(err, "Failed to provision service")

		// Set Degraded condition
		meta.SetStatusCondition(&service.Status.Conditions, metav1.Condition{
			Type:               TypeDegraded,
			Status:             metav1.ConditionTrue,
			Reason:             "ProvisioningFailed",
			Message:            fmt.Sprintf("Failed to provision: %v", err),
			ObservedGeneration: service.Generation,
		})
		meta.SetStatusCondition(&service.Status.Conditions, metav1.Condition{
			Type:               TypeAvailable,
			Status:             metav1.ConditionFalse,
			Reason:             "ProvisioningFailed",
			Message:            "Service is not available",
			ObservedGeneration: service.Generation,
		})
		meta.SetStatusCondition(&service.Status.Conditions, metav1.Condition{
			Type:               TypeProgressing,
			Status:             metav1.ConditionFalse,
			Reason:             "ProvisioningFailed",
			Message:            "Provisioning failed",
			ObservedGeneration: service.Generation,
		})

		service.Status.Phase = "Failed"
		service.Status.ObservedGeneration = service.Generation
		r.Status().Update(ctx, service)
		return ctrl.Result{}, err
	}

	// Update final status - service is ready
	service.Status.ReadyReplicas = service.Spec.Replicas
	service.Status.Phase = "Ready"
	service.Status.ObservedGeneration = service.Generation

	// Set Available condition
	meta.SetStatusCondition(&service.Status.Conditions, metav1.Condition{
		Type:               TypeAvailable,
		Status:             metav1.ConditionTrue,
		Reason:             "ServiceReady",
		Message:            "Service is available",
		ObservedGeneration: service.Generation,
	})

	// Clear Progressing condition
	meta.SetStatusCondition(&service.Status.Conditions, metav1.Condition{
		Type:               TypeProgressing,
		Status:             metav1.ConditionFalse,
		Reason:             "ReconcileComplete",
		Message:            "Reconciliation complete",
		ObservedGeneration: service.Generation,
	})

	// Clear Degraded condition
	meta.SetStatusCondition(&service.Status.Conditions, metav1.Condition{
		Type:               TypeDegraded,
		Status:             metav1.ConditionFalse,
		Reason:             "ServiceHealthy",
		Message:            "Service is healthy",
		ObservedGeneration: service.Generation,
	})

	if err := r.Status().Update(ctx, service); err != nil {
		logger.Error(err, "Failed to update ManagedService status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled ManagedService")
	return ctrl.Result{}, nil
}

func (r *ManagedServiceReconciler) provisionService(ctx context.Context, service *statusv1alpha1.ManagedService) error {
	logger := log.FromContext(ctx)

	logger.Info("Provisioning service",
		"type", service.Spec.ServiceType,
		"version", service.Spec.Version,
		"replicas", service.Spec.Replicas)

	// Simulate provisioning time
	time.Sleep(100 * time.Millisecond)

	// Simulate some validation
	if service.Spec.Replicas < 1 {
		return fmt.Errorf("replicas must be at least 1")
	}

	logger.Info("Service provisioned successfully")
	return nil
}

func (r *ManagedServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&statusv1alpha1.ManagedService{}).
		Named("managedservice").
		Complete(r)
}
