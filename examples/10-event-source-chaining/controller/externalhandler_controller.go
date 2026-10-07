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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	eventsourcev1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/10-event-source-chaining/api/v1alpha1"
)

const MaxRecentEvents = 5

// ExternalEvent represents an event from an external source
type ExternalEvent struct {
	EventID   string
	EventType string
	Message   string
	Source    string
	Namespace string
	Name      string
}

// ExternalHandlerReconciler reconciles an ExternalHandler object
type ExternalHandlerReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	EventChannel chan ExternalEvent
}

// +kubebuilder:rbac:groups=eventsource.examples.k8s.io,resources=externalhandlers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventsource.examples.k8s.io,resources=externalhandlers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=eventsource.examples.k8s.io,resources=externalhandlers/finalizers,verbs=update

func (r *ExternalHandlerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	handler := &eventsourcev1alpha1.ExternalHandler{}
	err := r.Get(ctx, req.NamespacedName, handler)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ExternalHandler resource not found")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get ExternalHandler")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling ExternalHandler", "name", handler.Name, "namespace", handler.Namespace)

	// Mark as active
	handler.Status.IsActive = true

	// Process any pending events from channel (non-blocking)
	select {
	case externalEvent := <-r.EventChannel:
		if externalEvent.Name == handler.Name && externalEvent.Namespace == handler.Namespace {
			if err := r.processExternalEvent(ctx, handler, externalEvent); err != nil {
				logger.Error(err, "Failed to process external event")
				return ctrl.Result{}, err
			}
		}
	default:
		// No events pending
	}

	if err := r.Status().Update(ctx, handler); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled ExternalHandler")
	// Requeue after 30 seconds to check for more events
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *ExternalHandlerReconciler) processExternalEvent(ctx context.Context, handler *eventsourcev1alpha1.ExternalHandler, evt ExternalEvent) error {
	logger := log.FromContext(ctx)

	// Filter by event type if configured
	if handler.Spec.EventFilter != "" && evt.EventType != handler.Spec.EventFilter {
		logger.V(1).Info("Event filtered out", "eventType", evt.EventType, "filter", handler.Spec.EventFilter)
		return nil
	}

	// Check max events limit
	if handler.Spec.MaxEvents > 0 && handler.Status.TotalEventsProcessed >= int64(handler.Spec.MaxEvents) {
		logger.Info("Max events limit reached", "limit", handler.Spec.MaxEvents)
		return nil
	}

	logger.Info("Processing external event",
		"eventID", evt.EventID,
		"eventType", evt.EventType,
		"source", evt.Source,
		"message", evt.Message)

	// Update status
	handler.Status.TotalEventsProcessed++
	handler.Status.LastProcessedEvent = &eventsourcev1alpha1.ExternalEvent{
		EventID:   evt.EventID,
		EventType: evt.EventType,
		Source:    evt.Source,
		Message:   evt.Message,
		Timestamp: metav1.Now(),
	}

	// Add to recent events
	eventInfo := eventsourcev1alpha1.ExternalEvent{
		EventID:   evt.EventID,
		EventType: evt.EventType,
		Source:    evt.Source,
		Message:   evt.Message,
		Timestamp: metav1.Now(),
	}

	handler.Status.RecentEvents = append([]eventsourcev1alpha1.ExternalEvent{eventInfo}, handler.Status.RecentEvents...)
	if len(handler.Status.RecentEvents) > MaxRecentEvents {
		handler.Status.RecentEvents = handler.Status.RecentEvents[:MaxRecentEvents]
	}

	logger.Info("Event processed successfully", "totalProcessed", handler.Status.TotalEventsProcessed)
	return nil
}

func (r *ExternalHandlerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize event channel if not set
	if r.EventChannel == nil {
		r.EventChannel = make(chan ExternalEvent, 100)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&eventsourcev1alpha1.ExternalHandler{}).
		Named("externalhandler").
		Complete(r)
}

// StartExternalEventSimulator starts a goroutine that simulates external events
func StartExternalEventSimulator(ctx context.Context, client client.Client, eventChan chan ExternalEvent) {
	logger := ctrl.Log.WithName("external-event-simulator")
	logger.Info("Starting external event simulator")

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		eventID := 0
		for {
			select {
			case <-ctx.Done():
				logger.Info("Stopping external event simulator")
				return
			case <-ticker.C:
				// List all ExternalHandlers
				handlerList := &eventsourcev1alpha1.ExternalHandlerList{}
				if err := client.List(ctx, handlerList); err != nil {
					logger.Error(err, "Failed to list ExternalHandlers")
					continue
				}

				// Generate events for each handler
				for _, handler := range handlerList.Items {
					eventID++
					evt := ExternalEvent{
						EventID:   fmt.Sprintf("evt-%d", eventID),
						EventType: pickEventType(eventID),
						Message:   fmt.Sprintf("Simulated event #%d", eventID),
						Source:    "external-simulator",
						Namespace: handler.Namespace,
						Name:      handler.Name,
					}

					select {
					case eventChan <- evt:
						logger.Info("Generated event",
							"eventID", evt.EventID,
							"type", evt.EventType,
							"for", handler.Name)

						// Trigger reconciliation
						namespacedName := types.NamespacedName{
							Namespace: handler.Namespace,
							Name:      handler.Name,
						}
						updatedHandler := &eventsourcev1alpha1.ExternalHandler{}
						if err := client.Get(ctx, namespacedName, updatedHandler); err == nil {
							// Just update to trigger reconcile
							client.Update(ctx, updatedHandler)
						}
					default:
						logger.Info("Event channel full, skipping event")
					}
				}
			}
		}
	}()
}

func pickEventType(id int) string {
	types := []string{"webhook", "message", "alert", "notification"}
	return types[id%len(types)]
}
