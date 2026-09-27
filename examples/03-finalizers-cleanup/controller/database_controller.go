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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	finalizersv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/03-finalizers-cleanup/api/v1alpha1"
)

const (
	databaseFinalizer = "finalizers.examples.k8s.io/database-cleanup"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// ExternalClient would be a real client to your database service
	// For this example, we'll simulate it with in-memory state
	externalDatabases map[string]bool
}

// +kubebuilder:rbac:groups=finalizers.examples.k8s.io,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=finalizers.examples.k8s.io,resources=databases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=finalizers.examples.k8s.io,resources=databases/finalizers,verbs=update

// Reconcile handles Database resources with finalizers for cleanup
//
// The reconciliation logic follows this pattern:
// 1. Check if resource is being deleted (DeletionTimestamp set)
// 2. If being deleted and has our finalizer, perform cleanup and remove finalizer
// 3. If not being deleted and doesn't have our finalizer, add it
// 4. Perform normal reconciliation logic
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Database instance
	database := &finalizersv1alpha1.Database{}
	err := r.Get(ctx, req.NamespacedName, database)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Database resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Database")
		return ctrl.Result{}, err
	}

	// Examine DeletionTimestamp to determine if object is under deletion
	if database.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so register our finalizer if it's not present
		if !controllerutil.ContainsFinalizer(database, databaseFinalizer) {
			logger.Info("Adding finalizer to Database")
			controllerutil.AddFinalizer(database, databaseFinalizer)
			if err := r.Update(ctx, database); err != nil {
				logger.Error(err, "Failed to add finalizer")
				return ctrl.Result{}, err
			}
			logger.Info("Finalizer added successfully")
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(database, databaseFinalizer) {
			logger.Info("Database is being deleted, performing cleanup")

			// Update status to show we're deleting
			database.Status.State = "Deleting"
			database.Status.Message = "Cleaning up external resources"
			if err := r.Status().Update(ctx, database); err != nil {
				logger.Error(err, "Failed to update status during deletion")
				// Continue with cleanup even if status update fails
			}

			// Perform cleanup of external resources
			if err := r.cleanupExternalResources(ctx, database); err != nil {
				logger.Error(err, "Failed to cleanup external resources")
				// Return error to retry cleanup
				return ctrl.Result{}, err
			}

			logger.Info("External resources cleaned up successfully")

			// Remove our finalizer from the list and update it
			controllerutil.RemoveFinalizer(database, databaseFinalizer)
			if err := r.Update(ctx, database); err != nil {
				logger.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, err
			}

			logger.Info("Finalizer removed, Database will be deleted")
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Normal reconciliation logic
	logger.Info("Reconciling Database",
		"name", database.Name,
		"namespace", database.Namespace,
		"databaseName", database.Spec.DatabaseName,
		"storageSizeMB", database.Spec.StorageSizeMB)

	// Provision the database if needed
	if err := r.provisionDatabase(ctx, database); err != nil {
		logger.Error(err, "Failed to provision database")
		database.Status.State = "Failed"
		database.Status.Message = err.Error()
		r.Status().Update(ctx, database)
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateStatus(ctx, database); err != nil {
		logger.Error(err, "Failed to update Database status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled Database")
	return ctrl.Result{}, nil
}

// provisionDatabase simulates provisioning a database in an external system
func (r *DatabaseReconciler) provisionDatabase(ctx context.Context, database *finalizersv1alpha1.Database) error {
	logger := log.FromContext(ctx)

	// Initialize the map if nil (first use)
	if r.externalDatabases == nil {
		r.externalDatabases = make(map[string]bool)
	}

	// Check if already provisioned
	externalID := fmt.Sprintf("db-%s-%s", database.Namespace, database.Name)
	if r.externalDatabases[externalID] {
		logger.Info("Database already provisioned", "externalID", externalID)
		return nil
	}

	// Simulate provisioning
	logger.Info("Provisioning database in external system",
		"databaseName", database.Spec.DatabaseName,
		"storageSizeMB", database.Spec.StorageSizeMB,
		"backupEnabled", database.Spec.BackupEnabled)

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Mark as provisioned
	r.externalDatabases[externalID] = true
	database.Status.ExternalID = externalID

	logger.Info("Database provisioned successfully", "externalID", externalID)
	return nil
}

// cleanupExternalResources cleans up the external database
func (r *DatabaseReconciler) cleanupExternalResources(ctx context.Context, database *finalizersv1alpha1.Database) error {
	logger := log.FromContext(ctx)

	externalID := database.Status.ExternalID
	if externalID == "" {
		logger.Info("No external ID found, nothing to cleanup")
		return nil
	}

	logger.Info("Cleaning up external database", "externalID", externalID)

	// Simulate cleanup work
	time.Sleep(100 * time.Millisecond)

	// Remove from our simulated external system
	if r.externalDatabases != nil {
		delete(r.externalDatabases, externalID)
	}

	logger.Info("External database deleted successfully", "externalID", externalID)
	return nil
}

// updateStatus updates the Database status
func (r *DatabaseReconciler) updateStatus(ctx context.Context, database *finalizersv1alpha1.Database) error {
	database.Status.State = "Ready"
	database.Status.Message = "Database is ready"
	database.Status.LastUpdated = metav1.Now()
	database.Status.ObservedGeneration = database.Generation

	if err := r.Status().Update(ctx, database); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&finalizersv1alpha1.Database{}).
		Named("database").
		Complete(r)
}
