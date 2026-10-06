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
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	eventsourcev1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/10-event-source-chaining/api/v1alpha1"
)

const (
	// Condition types
	ConditionTypeActive    = "Active"
	ConditionTypeReady     = "Ready"
	ConditionTypeProcessing = "Processing"

	// Condition reasons
	ReasonActivated       = "Activated"
	ReasonDeactivated     = "Deactivated"
	ReasonEventProcessed  = "EventProcessed"
	ReasonProcessingError = "ProcessingError"

	// Maximum number of recent events to keep
	MaxRecentEvents = 5
)

// ExternalEvent represents an event from an external source
type ExternalEvent struct {
	EventID   string
	EventType string
	Message   string
	Source    string
	// Namespace and Name identify which ExternalHandler should process this
	Namespace string
	Name      string
}

// ExternalHandlerReconciler reconciles an ExternalHandler object
type ExternalHandlerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// EventChannel receives external events
	EventChannel chan event.GenericEvent
}

// +kubebuilder:rbac:groups=eventsource.examples.k8s.io,resources=externalhandlers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventsource.examples.k8s.io,resources=externalhandlers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=eventsource.examples.k8s.io,resources=externalhandlers/finalizers,verbs=update

// Reconcile handles ExternalHandler resources
func (r *ExternalHandlerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling ExternalHandler", "namespace", req.Namespace, "name", req.Name)

	// Fetch the ExternalHandler instance
	handler := &eventsourcev1alpha1.ExternalHandler{}
	if err := r.Get(ctx, req.NamespacedName, handler); err != nil {
		if errors.IsNotFound(err) {
			log.Info("ExternalHandler resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get ExternalHandler")
		return ctrl.Result{}, err
	}

	// Create a copy for status updates
	original := handler.DeepCopy()

	// Check if external event data is present in annotations
	if eventID, hasEvent := handler.Annotations["event.id"]; hasEvent {
		externalEvent := &ExternalEvent{
			EventID:   eventID,
			EventType: handler.Annotations["event.type"],
			Message:   handler.Annotations["event.message"],
			Source:    handler.Annotations["event.source"],
			Namespace: handler.Namespace,
			Name:      handler.Name,
		}

		// Process the external event
		if err := r.HandleExternalEvent(ctx, externalEvent, handler); err != nil {
			log.Error(err, "Failed to handle external event")
			r.setCondition(handler, ConditionTypeProcessing, metav1.ConditionFalse, ReasonProcessingError, err.Error())
			r.updateStatus(ctx, original, handler)
			return ctrl.Result{}, err
		}

		log.Info("Processed external event",
			"eventID", eventID,
			"eventType", externalEvent.EventType,
			"totalProcessed", handler.Status.TotalEventsProcessed)
	}

	// Process the handler
	if err := r.processHandler(ctx, handler); err != nil {
		log.Error(err, "Failed to process ExternalHandler")
		r.setCondition(handler, ConditionTypeProcessing, metav1.ConditionFalse, ReasonProcessingError, err.Error())
		r.updateStatus(ctx, original, handler)
		return ctrl.Result{}, err
	}

	// Update status
	now := metav1.Now()
	handler.Status.LastReconcileTime = &now
	handler.Status.IsActive = true

	// Set conditions
	r.setCondition(handler, ConditionTypeReady, metav1.ConditionTrue, ReasonActivated, "Handler is ready")
	r.setCondition(handler, ConditionTypeActive, metav1.ConditionTrue, ReasonActivated, "Handler is processing events")

	// Update status if changed
	if err := r.updateStatus(ctx, original, handler); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled ExternalHandler",
		"totalEvents", handler.Status.TotalEventsProcessed,
		"isActive", handler.Status.IsActive)

	return ctrl.Result{}, nil
}

// processHandler handles the main logic for processing external events
func (r *ExternalHandlerReconciler) processHandler(ctx context.Context, handler *eventsourcev1alpha1.ExternalHandler) error {
	log := log.FromContext(ctx)

	// Check if we've reached max events
	if handler.Spec.MaxEvents > 0 && handler.Status.TotalEventsProcessed >= int64(handler.Spec.MaxEvents) {
		log.Info("Maximum events reached, deactivating handler",
			"maxEvents", handler.Spec.MaxEvents,
			"processed", handler.Status.TotalEventsProcessed)
		handler.Status.IsActive = false
		r.setCondition(handler, ConditionTypeActive, metav1.ConditionFalse, ReasonDeactivated,
			fmt.Sprintf("Reached maximum events limit of %d", handler.Spec.MaxEvents))
		return nil
	}

	// This reconciliation was triggered by an external event
	// The event details would be in the context or retrieved from a queue
	log.Info("Processing reconciliation triggered by event source",
		"processingMode", handler.Spec.ProcessingMode,
		"eventFilter", handler.Spec.EventFilter)

	return nil
}

// HandleExternalEvent processes an external event and updates the handler status
func (r *ExternalHandlerReconciler) HandleExternalEvent(ctx context.Context, externalEvent *ExternalEvent, handler *eventsourcev1alpha1.ExternalHandler) error {
	log := log.FromContext(ctx)

	// Create event record
	evt := eventsourcev1alpha1.ExternalEvent{
		EventID:   externalEvent.EventID,
		EventType: externalEvent.EventType,
		Timestamp: metav1.Now(),
		Message:   externalEvent.Message,
		Source:    externalEvent.Source,
	}

	// Check event filter if specified
	if handler.Spec.EventFilter != "" && externalEvent.EventType != handler.Spec.EventFilter {
		log.Info("Event filtered out", "eventType", externalEvent.EventType, "filter", handler.Spec.EventFilter)
		return nil
	}

	// Update status
	handler.Status.TotalEventsProcessed++
	handler.Status.LastProcessedEvent = &evt

	// Add to recent events (keep only last 5)
	handler.Status.RecentEvents = append([]eventsourcev1alpha1.ExternalEvent{evt}, handler.Status.RecentEvents...)
	if len(handler.Status.RecentEvents) > MaxRecentEvents {
		handler.Status.RecentEvents = handler.Status.RecentEvents[:MaxRecentEvents]
	}

	// Set condition
	r.setCondition(handler, ConditionTypeProcessing, metav1.ConditionTrue, ReasonEventProcessed,
		fmt.Sprintf("Processed event %s of type %s", externalEvent.EventID, externalEvent.EventType))

	log.Info("External event processed",
		"eventID", externalEvent.EventID,
		"eventType", externalEvent.EventType,
		"totalProcessed", handler.Status.TotalEventsProcessed)

	return nil
}

// setCondition sets or updates a condition
func (r *ExternalHandlerReconciler) setCondition(handler *eventsourcev1alpha1.ExternalHandler, condType string, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               condType,
		Status:             status,
		ObservedGeneration: handler.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}

	meta.SetStatusCondition(&handler.Status.Conditions, condition)
}

// updateStatus updates the handler status if it has changed
func (r *ExternalHandlerReconciler) updateStatus(ctx context.Context, original, updated *eventsourcev1alpha1.ExternalHandler) error {
	if updated.Status.TotalEventsProcessed != original.Status.TotalEventsProcessed ||
		updated.Status.IsActive != original.Status.IsActive ||
		!conditionsEqual(original.Status.Conditions, updated.Status.Conditions) {
		return r.Status().Update(ctx, updated)
	}
	return nil
}

// conditionsEqual checks if two condition slices are equal
func conditionsEqual(a, b []metav1.Condition) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Type != b[i].Type ||
			a[i].Status != b[i].Status ||
			a[i].Reason != b[i].Reason {
			return false
		}
	}
	return true
}

// SetupWithManager sets up the controller with the Manager
func (r *ExternalHandlerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize event channel if not already set
	if r.EventChannel == nil {
		r.EventChannel = make(chan event.GenericEvent, 100)
	}

	// Create a typed channel for ExternalHandler events
	typedChannel := make(chan event.TypedGenericEvent[*eventsourcev1alpha1.ExternalHandler], 100)

	// Start a goroutine to convert generic events to typed events
	go func() {
		for evt := range r.EventChannel {
			if handler, ok := evt.Object.(*eventsourcev1alpha1.ExternalHandler); ok {
				typedChannel <- event.TypedGenericEvent[*eventsourcev1alpha1.ExternalHandler]{
					Object: handler,
				}
			}
		}
	}()

	return ctrl.NewControllerManagedBy(mgr).
		For(&eventsourcev1alpha1.ExternalHandler{}).
		// Watch for events from the channel source
		WatchesRawSource(
			source.Channel(
				typedChannel,
				&handler.TypedEnqueueRequestForObject[*eventsourcev1alpha1.ExternalHandler]{},
			),
		).
		Complete(r)
}

// StartExternalEventSimulator starts a goroutine that simulates external events
func StartExternalEventSimulator(ctx context.Context, client client.Client, eventChan chan event.GenericEvent) {
	log := ctrl.Log.WithName("external-event-simulator")
	log.Info("Starting external event simulator")

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		eventCounter := 0

		for {
			select {
			case <-ctx.Done():
				log.Info("Stopping external event simulator")
				return
			case <-ticker.C:
				// List all ExternalHandler resources
				handlerList := &eventsourcev1alpha1.ExternalHandlerList{}
				if err := client.List(ctx, handlerList); err != nil {
					log.Error(err, "Failed to list ExternalHandlers")
					continue
				}

				// Generate events for active handlers
				for _, handler := range handlerList.Items {
					// Skip if handler is not active or has reached max events
					if !handler.Status.IsActive && handler.Status.TotalEventsProcessed > 0 {
						continue
					}

					if handler.Spec.MaxEvents > 0 && handler.Status.TotalEventsProcessed >= int64(handler.Spec.MaxEvents) {
						continue
					}

					eventCounter++
					eventID := fmt.Sprintf("evt-%d-%d", time.Now().Unix(), eventCounter)

					// Determine event type
					eventTypes := []string{"webhook", "message", "alert", "notification"}
					eventType := eventTypes[eventCounter%len(eventTypes)]

					log.Info("Generating external event",
						"eventID", eventID,
						"eventType", eventType,
						"namespace", handler.Namespace,
						"name", handler.Name)

					// Create a copy of the handler with event annotations
					eventHandler := handler.DeepCopy()
					if eventHandler.Annotations == nil {
						eventHandler.Annotations = make(map[string]string)
					}
					eventHandler.Annotations["event.id"] = eventID
					eventHandler.Annotations["event.type"] = eventType
					eventHandler.Annotations["event.message"] = fmt.Sprintf("External event %s from simulator", eventID)
					eventHandler.Annotations["event.source"] = "external-simulator"

					// Send event to channel
					select {
					case eventChan <- event.GenericEvent{Object: eventHandler}:
						log.Info("External event sent to channel", "eventID", eventID)
					case <-ctx.Done():
						return
					default:
						log.Info("Event channel full, dropping event", "eventID", eventID)
					}
				}
			}
		}
	}()
}
