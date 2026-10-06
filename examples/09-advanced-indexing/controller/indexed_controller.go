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
	"sigs.k8s.io/controller-runtime/pkg/log"

	indexingv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/09-advanced-indexing/api/v1alpha1"
)

const (
	// Field index names
	categoryIndexField = "spec.category"
	ownerIndexField    = "spec.owner"
	priorityIndexField = "spec.priority"
)

// IndexedResourceReconciler reconciles an IndexedResource object
type IndexedResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=indexing.examples.k8s.io,resources=indexedresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=indexing.examples.k8s.io,resources=indexedresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=indexing.examples.k8s.io,resources=indexedresources/finalizers,verbs=update

// Reconcile handles the reconciliation logic for IndexedResource
func (r *IndexedResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Reconciling IndexedResource")

	// Fetch the IndexedResource instance
	resource := &indexingv1alpha1.IndexedResource{}
	if err := r.Get(ctx, req.NamespacedName, resource); err != nil {
		if errors.IsNotFound(err) {
			log.Info("IndexedResource not found, likely deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get IndexedResource")
		return ctrl.Result{}, err
	}

	// Demonstrate index usage: Find all resources with the same category
	relatedResources, err := r.findResourcesByCategory(ctx, resource.Spec.Category, resource.Namespace)
	if err != nil {
		log.Error(err, "Failed to find related resources by category")
		return ctrl.Result{}, err
	}

	// Filter out the current resource and build list of names
	var relatedNames []string
	for _, related := range relatedResources {
		if related.Name != resource.Name {
			relatedNames = append(relatedNames, related.Name)
		}
	}

	// Demonstrate another index: Find resources by owner if specified
	if resource.Spec.Owner != "" {
		ownerResources, err := r.findResourcesByOwner(ctx, resource.Spec.Owner, resource.Namespace)
		if err != nil {
			log.Error(err, "Failed to find resources by owner")
			return ctrl.Result{}, err
		}
		log.Info("Resources found by owner", "owner", resource.Spec.Owner, "count", len(ownerResources))
	}

	// Demonstrate priority-based lookup
	priorityResources, err := r.findResourcesByPriority(ctx, resource.Spec.Priority, resource.Namespace)
	if err != nil {
		log.Error(err, "Failed to find resources by priority")
		return ctrl.Result{}, err
	}
	log.Info("Resources found by priority", "priority", resource.Spec.Priority, "count", len(priorityResources))

	// Update status with indexed lookup results
	resource.Status.Phase = "Active"
	resource.Status.LastProcessed = metav1.NewTime(time.Now())
	resource.Status.RelatedResources = relatedNames
	resource.Status.Message = fmt.Sprintf("Found %d related resources in category '%s'",
		len(relatedNames), resource.Spec.Category)

	if err := r.Status().Update(ctx, resource); err != nil {
		log.Error(err, "Failed to update IndexedResource status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled IndexedResource",
		"category", resource.Spec.Category,
		"relatedCount", len(relatedNames))

	return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
}

// findResourcesByCategory uses the category index for fast lookups
func (r *IndexedResourceReconciler) findResourcesByCategory(ctx context.Context, category, namespace string) ([]indexingv1alpha1.IndexedResource, error) {
	var resourceList indexingv1alpha1.IndexedResourceList

	// Use MatchingFields to leverage the index - this is much faster than listing all
	// resources and filtering in memory
	if err := r.List(ctx, &resourceList,
		client.InNamespace(namespace),
		client.MatchingFields{categoryIndexField: category},
	); err != nil {
		return nil, err
	}

	return resourceList.Items, nil
}

// findResourcesByOwner uses the owner index for fast lookups
func (r *IndexedResourceReconciler) findResourcesByOwner(ctx context.Context, owner, namespace string) ([]indexingv1alpha1.IndexedResource, error) {
	var resourceList indexingv1alpha1.IndexedResourceList

	if err := r.List(ctx, &resourceList,
		client.InNamespace(namespace),
		client.MatchingFields{ownerIndexField: owner},
	); err != nil {
		return nil, err
	}

	return resourceList.Items, nil
}

// findResourcesByPriority uses the priority index for fast lookups
func (r *IndexedResourceReconciler) findResourcesByPriority(ctx context.Context, priority int, namespace string) ([]indexingv1alpha1.IndexedResource, error) {
	var resourceList indexingv1alpha1.IndexedResourceList

	// For integer fields, we convert to string for indexing
	priorityStr := fmt.Sprintf("%d", priority)
	if err := r.List(ctx, &resourceList,
		client.InNamespace(namespace),
		client.MatchingFields{priorityIndexField: priorityStr},
	); err != nil {
		return nil, err
	}

	return resourceList.Items, nil
}

// SetupWithManager sets up the controller with the Manager and configures field indexes
func (r *IndexedResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Setup field index for category field
	// This allows fast lookups using client.MatchingFields
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&indexingv1alpha1.IndexedResource{},
		categoryIndexField,
		func(rawObj client.Object) []string {
			resource := rawObj.(*indexingv1alpha1.IndexedResource)
			return []string{resource.Spec.Category}
		},
	); err != nil {
		return err
	}

	// Setup field index for owner field
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&indexingv1alpha1.IndexedResource{},
		ownerIndexField,
		func(rawObj client.Object) []string {
			resource := rawObj.(*indexingv1alpha1.IndexedResource)
			if resource.Spec.Owner == "" {
				return nil
			}
			return []string{resource.Spec.Owner}
		},
	); err != nil {
		return err
	}

	// Setup field index for priority field
	// Note: For integer fields, we index the string representation
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&indexingv1alpha1.IndexedResource{},
		priorityIndexField,
		func(rawObj client.Object) []string {
			resource := rawObj.(*indexingv1alpha1.IndexedResource)
			return []string{fmt.Sprintf("%d", resource.Spec.Priority)}
		},
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&indexingv1alpha1.IndexedResource{}).
		Complete(r)
}
