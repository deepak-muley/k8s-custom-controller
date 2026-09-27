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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	examplesv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/01-basic-reconciler/api/v1alpha1"
)

// GuestbookReconciler reconciles a Guestbook object
type GuestbookReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=examples.k8s.io,resources=guestbooks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=examples.k8s.io,resources=guestbooks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=examples.k8s.io,resources=guestbooks/finalizers,verbs=update

// Reconcile is the main reconciliation loop
// The Reconcile function compares the state specified by the Guestbook object
// against the actual cluster state, and then performs operations to make the
// cluster state reflect the state specified by the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *GuestbookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Guestbook instance
	guestbook := &examplesv1alpha1.Guestbook{}
	err := r.Get(ctx, req.NamespacedName, guestbook)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			logger.Info("Guestbook resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get Guestbook")
		return ctrl.Result{}, err
	}

	// Log the reconciliation
	logger.Info("Reconciling Guestbook",
		"name", guestbook.Name,
		"namespace", guestbook.Namespace,
		"message", guestbook.Spec.Message,
		"replicas", guestbook.Spec.Replicas)

	// Process the Guestbook
	if err := r.processGuestbook(ctx, guestbook); err != nil {
		logger.Error(err, "Failed to process Guestbook")
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateStatus(ctx, guestbook); err != nil {
		logger.Error(err, "Failed to update Guestbook status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled Guestbook")
	return ctrl.Result{}, nil
}

// processGuestbook performs the main business logic
func (r *GuestbookReconciler) processGuestbook(ctx context.Context, guestbook *examplesv1alpha1.Guestbook) error {
	logger := log.FromContext(ctx)

	// Validate replicas
	if guestbook.Spec.Replicas < 1 {
		return fmt.Errorf("replicas must be at least 1, got %d", guestbook.Spec.Replicas)
	}

	// Simulate processing
	logger.Info("Processing guestbook message",
		"message", guestbook.Spec.Message,
		"replicas", guestbook.Spec.Replicas)

	// In a real controller, you would create/update child resources here
	// For example: Deployments, Services, ConfigMaps, etc.

	return nil
}

// updateStatus updates the Guestbook status
func (r *GuestbookReconciler) updateStatus(ctx context.Context, guestbook *examplesv1alpha1.Guestbook) error {
	// Update status fields
	guestbook.Status.State = "Active"
	guestbook.Status.LastUpdated = metav1.Now()
	guestbook.Status.ObservedGeneration = guestbook.Generation

	// Update the status subresource
	// This is separate from updating the spec to avoid unnecessary reconciliations
	if err := r.Status().Update(ctx, guestbook); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GuestbookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplesv1alpha1.Guestbook{}).
		Named("guestbook").
		Complete(r)
}
