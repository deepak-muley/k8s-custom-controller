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
	"math/rand"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"

	metricsv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/08-metrics-events/api/v1alpha1"
)

var (
	// Custom Prometheus metrics
	reconcileCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "monitored_reconcile_total",
			Help: "Total number of reconciliations for Monitored resources",
		},
		[]string{"namespace", "name", "result"},
	)

	reconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "monitored_reconcile_duration_seconds",
			Help:    "Duration of Monitored reconciliations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace", "name"},
	)

	monitoredGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "monitored_current_value",
			Help: "Current monitored value for each Monitored resource",
		},
		[]string{"namespace", "name", "target"},
	)

	thresholdExceededCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "monitored_threshold_exceeded_total",
			Help: "Total number of times threshold was exceeded",
		},
		[]string{"namespace", "name", "target"},
	)

	activeMonitored = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "monitored_active_total",
			Help: "Total number of active Monitored resources",
		},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(
		reconcileCounter,
		reconcileDuration,
		monitoredGauge,
		thresholdExceededCounter,
		activeMonitored,
	)
}

// MonitoredReconciler reconciles a Monitored object
type MonitoredReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=metrics.examples.k8s.io,resources=monitoreds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=metrics.examples.k8s.io,resources=monitoreds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=metrics.examples.k8s.io,resources=monitoreds/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is the main reconciliation loop
// The Reconcile function compares the state specified by the Monitored object
// against the actual cluster state, and then performs operations to make the
// cluster state reflect the state specified by the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *MonitoredReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()

	// Track reconciliation duration
	defer func() {
		duration := time.Since(startTime).Seconds()
		reconcileDuration.WithLabelValues(req.Namespace, req.Name).Observe(duration)
	}()

	// Fetch the Monitored instance
	monitored := &metricsv1alpha1.Monitored{}
	err := r.Get(ctx, req.NamespacedName, monitored)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, could have been deleted after reconcile request.
			// Decrement active monitored gauge
			activeMonitored.Dec()

			// Record successful reconciliation (deletion)
			reconcileCounter.WithLabelValues(req.Namespace, req.Name, "success").Inc()

			logger.Info("Monitored resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reconcileCounter.WithLabelValues(req.Namespace, req.Name, "error").Inc()
		logger.Error(err, "Failed to get Monitored")
		return ctrl.Result{}, err
	}

	// Log the reconciliation
	logger.Info("Reconciling Monitored",
		"name", monitored.Name,
		"namespace", monitored.Namespace,
		"target", monitored.Spec.Target,
		"threshold", monitored.Spec.Threshold)

	// Record event for reconciliation start
	r.Recorder.Event(monitored, corev1.EventTypeNormal, "ReconcileStarted",
		fmt.Sprintf("Starting reconciliation for target: %s", monitored.Spec.Target))

	// Process the Monitored resource
	if err := r.processMonitored(ctx, monitored); err != nil {
		// Record failure event
		r.Recorder.Event(monitored, corev1.EventTypeWarning, "ProcessingFailed",
			fmt.Sprintf("Failed to process monitored resource: %v", err))

		reconcileCounter.WithLabelValues(req.Namespace, req.Name, "error").Inc()
		logger.Error(err, "Failed to process Monitored")
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateStatus(ctx, monitored); err != nil {
		// Record failure event
		r.Recorder.Event(monitored, corev1.EventTypeWarning, "StatusUpdateFailed",
			fmt.Sprintf("Failed to update status: %v", err))

		reconcileCounter.WithLabelValues(req.Namespace, req.Name, "error").Inc()
		logger.Error(err, "Failed to update Monitored status")
		return ctrl.Result{}, err
	}

	// Record successful reconciliation
	reconcileCounter.WithLabelValues(req.Namespace, req.Name, "success").Inc()

	// Record success event
	r.Recorder.Event(monitored, corev1.EventTypeNormal, "ReconcileSuccess",
		fmt.Sprintf("Successfully reconciled, state: %s, value: %d",
			monitored.Status.State, monitored.Status.CurrentValue))

	// Update active monitored gauge
	activeMonitored.Inc()

	logger.Info("Successfully reconciled Monitored")

	// Requeue after the specified interval to simulate continuous monitoring
	requeueAfter := time.Duration(monitored.Spec.Interval) * time.Second
	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// processMonitored performs the main business logic
func (r *MonitoredReconciler) processMonitored(ctx context.Context, monitored *metricsv1alpha1.Monitored) error {
	logger := log.FromContext(ctx)

	// Simulate monitoring by generating a random value
	// In a real controller, you would query actual metrics from your target system
	currentValue := rand.Int31n(100)

	// Update the gauge metric
	monitoredGauge.WithLabelValues(
		monitored.Namespace,
		monitored.Name,
		monitored.Spec.Target,
	).Set(float64(currentValue))

	logger.Info("Monitoring target",
		"target", monitored.Spec.Target,
		"currentValue", currentValue,
		"threshold", monitored.Spec.Threshold)

	// Check if threshold is exceeded
	if currentValue > monitored.Spec.Threshold {
		// Increment threshold exceeded counter
		thresholdExceededCounter.WithLabelValues(
			monitored.Namespace,
			monitored.Name,
			monitored.Spec.Target,
		).Inc()

		// Record warning event
		r.Recorder.Event(monitored, corev1.EventTypeWarning, "ThresholdExceeded",
			fmt.Sprintf("Current value %d exceeds threshold %d for target %s",
				currentValue, monitored.Spec.Threshold, monitored.Spec.Target))

		monitored.Status.State = "Alerting"
		logger.Info("Threshold exceeded, alerting", "currentValue", currentValue)
	} else {
		// Record normal event
		r.Recorder.Event(monitored, corev1.EventTypeNormal, "HealthCheck",
			fmt.Sprintf("Health check passed: %d/%d for target %s",
				currentValue, monitored.Spec.Threshold, monitored.Spec.Target))

		monitored.Status.State = "Healthy"
		logger.Info("Health check passed", "currentValue", currentValue)
	}

	// Update status fields
	monitored.Status.CurrentValue = currentValue
	monitored.Status.LastCheckTime = metav1.Now()
	monitored.Status.ReconcileCount++

	return nil
}

// updateStatus updates the Monitored status
func (r *MonitoredReconciler) updateStatus(ctx context.Context, monitored *metricsv1alpha1.Monitored) error {
	// Update observed generation
	monitored.Status.ObservedGeneration = monitored.Generation

	// Update the status subresource
	if err := r.Status().Update(ctx, monitored); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MonitoredReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&metricsv1alpha1.Monitored{}).
		Named("monitored").
		Complete(r)
}
