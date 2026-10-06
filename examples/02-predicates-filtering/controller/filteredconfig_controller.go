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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	predicatesv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/02-predicates-filtering/api/v1alpha1"
)

// FilteredConfigReconciler reconciles a FilteredConfig object
type FilteredConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=predicates.examples.k8s.io,resources=filteredconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=predicates.examples.k8s.io,resources=filteredconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=predicates.examples.k8s.io,resources=filteredconfigs/finalizers,verbs=update

func (r *FilteredConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	config := &predicatesv1alpha1.FilteredConfig{}
	err := r.Get(ctx, req.NamespacedName, config)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("FilteredConfig resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get FilteredConfig")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling FilteredConfig - this should only happen for meaningful changes",
		"name", config.Name,
		"namespace", config.Namespace,
		"generation", config.Generation)

	// Update status - increment processed count
	config.Status.ProcessedCount++
	config.Status.LastProcessed = metav1.Now()
	config.Status.ObservedGeneration = config.Generation

	if err := r.Status().Update(ctx, config); err != nil {
		logger.Error(err, "Failed to update FilteredConfig status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled FilteredConfig", "processedCount", config.Status.ProcessedCount)
	return ctrl.Result{}, nil
}

func (r *FilteredConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&predicatesv1alpha1.FilteredConfig{}).
		WithEventFilter(predicate.Funcs{
			// Only reconcile on create events
			CreateFunc: func(e event.CreateEvent) bool {
				return true
			},
			// Only reconcile on updates when generation changes (spec changes)
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldGen := e.ObjectOld.GetGeneration()
				newGen := e.ObjectNew.GetGeneration()

				// Filter: only process if generation changed (spec updated)
				if oldGen != newGen {
					log.Log.Info("Generation changed - will reconcile",
						"name", e.ObjectNew.GetName(),
						"oldGen", oldGen,
						"newGen", newGen)
					return true
				}

				// Filter out status-only updates
				log.Log.V(1).Info("Generation unchanged - skipping reconcile",
					"name", e.ObjectNew.GetName(),
					"generation", newGen)
				return false
			},
			// Only reconcile deletes if object has specific label
			DeleteFunc: func(e event.DeleteEvent) bool {
				labels := e.Object.GetLabels()
				if labels != nil && labels["reconcile-on-delete"] == "true" {
					log.Log.Info("Resource has reconcile-on-delete label - will process",
						"name", e.Object.GetName())
					return true
				}
				log.Log.Info("Resource does not have reconcile-on-delete label - skipping",
					"name", e.Object.GetName())
				return false
			},
			// Generic events - filter based on labels
			GenericFunc: func(e event.GenericEvent) bool {
				labels := e.Object.GetLabels()
				if labels != nil && labels["watch"] == "true" {
					return true
				}
				return false
			},
		}).
		Named("filteredconfig").
		Complete(r)
}
