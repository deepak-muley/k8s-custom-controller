package controller

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ratelimitv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/11-rate-limiting-backoff/api/v1alpha1"
)

// RateLimitedReconciler reconciles a RateLimited object
type RateLimitedReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ratelimit.examples.k8s.io,resources=ratelimiteds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ratelimit.examples.k8s.io,resources=ratelimiteds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ratelimit.examples.k8s.io,resources=ratelimiteds/finalizers,verbs=update

// Reconcile implements the reconciliation loop for RateLimited resources
func (r *RateLimitedReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting reconciliation", "resource", req.NamespacedName)

	// Fetch the RateLimited instance
	rateLimited := &ratelimitv1alpha1.RateLimited{}
	if err := r.Get(ctx, req.NamespacedName, rateLimited); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("RateLimited resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get RateLimited resource")
		return ctrl.Result{}, err
	}

	// Create a copy for status updates
	original := rateLimited.DeepCopy()

	// Update processing count
	rateLimited.Status.ProcessingCount++
	now := metav1.Now()
	rateLimited.Status.LastProcessedTime = &now

	// Set initial phase if not set
	if rateLimited.Status.Phase == "" {
		rateLimited.Status.Phase = "Pending"
	}

	// Simulate processing time if configured
	if rateLimited.Spec.ProcessingTime > 0 {
		rateLimited.Status.Phase = "Processing"
		rateLimited.Status.Message = fmt.Sprintf("Processing for %d seconds", rateLimited.Spec.ProcessingTime)
		logger.Info("Simulating processing time", "seconds", rateLimited.Spec.ProcessingTime)

		// Update status before processing
		if err := r.Status().Patch(ctx, rateLimited, client.MergeFrom(original)); err != nil {
			logger.Error(err, "Failed to update status before processing")
			return ctrl.Result{}, err
		}

		time.Sleep(time.Duration(rateLimited.Spec.ProcessingTime) * time.Second)
	}

	// Simulate failures based on failure rate percentage
	shouldFail := false
	if rateLimited.Spec.FailureRatePercent > 0 {
		random := rand.Intn(100)
		shouldFail = random < rateLimited.Spec.FailureRatePercent
		logger.Info("Failure simulation", "failureRatePercent", rateLimited.Spec.FailureRatePercent,
			"random", random, "shouldFail", shouldFail)
	}

	// Check if max retries exceeded
	maxRetries := rateLimited.Spec.MaxRetries
	if maxRetries == 0 {
		maxRetries = 5 // Default max retries
	}

	if shouldFail {
		rateLimited.Status.FailureCount++
		rateLimited.Status.LastFailureTime = &now
		rateLimited.Status.LastFailureReason = "Simulated failure based on failure rate"

		// Calculate exponential backoff duration
		backoffSeconds := calculateBackoff(rateLimited.Status.FailureCount)
		rateLimited.Status.BackoffDuration = backoffSeconds

		if rateLimited.Status.FailureCount >= maxRetries {
			// Max retries exceeded
			rateLimited.Status.Phase = "Failed"
			rateLimited.Status.Message = fmt.Sprintf("Maximum retries (%d) exceeded", maxRetries)
			logger.Info("Maximum retries exceeded", "failureCount", rateLimited.Status.FailureCount,
				"maxRetries", maxRetries)

			// Update status
			if err := r.Status().Patch(ctx, rateLimited, client.MergeFrom(original)); err != nil {
				logger.Error(err, "Failed to update status after max retries")
				return ctrl.Result{}, err
			}

			// Don't requeue - processing has permanently failed
			return ctrl.Result{}, nil
		}

		// Set rate limited phase
		rateLimited.Status.Phase = "RateLimited"
		rateLimited.Status.Message = fmt.Sprintf("Processing failed (attempt %d/%d), backing off for %d seconds",
			rateLimited.Status.FailureCount, maxRetries, backoffSeconds)

		logger.Info("Processing failed, will retry with backoff",
			"failureCount", rateLimited.Status.FailureCount,
			"backoffSeconds", backoffSeconds)

		// Update status
		if err := r.Status().Patch(ctx, rateLimited, client.MergeFrom(original)); err != nil {
			logger.Error(err, "Failed to update status after failure")
			return ctrl.Result{}, err
		}

		// Return with RequeueAfter for exponential backoff
		return ctrl.Result{RequeueAfter: time.Duration(backoffSeconds) * time.Second}, nil
	}

	// Processing succeeded
	rateLimited.Status.Phase = "Completed"
	rateLimited.Status.Message = "Processing completed successfully"
	rateLimited.Status.BackoffDuration = 0

	// Reset failure count on success
	if rateLimited.Status.FailureCount > 0 {
		logger.Info("Processing succeeded after failures", "previousFailures", rateLimited.Status.FailureCount)
		rateLimited.Status.FailureCount = 0
		rateLimited.Status.LastFailureReason = ""
	}

	// Update status
	if err := r.Status().Patch(ctx, rateLimited, client.MergeFrom(original)); err != nil {
		logger.Error(err, "Failed to update status after success")
		return ctrl.Result{}, err
	}

	// Check if requeue is configured
	if rateLimited.Spec.RequeueAfterSeconds > 0 {
		logger.Info("Scheduling requeue", "seconds", rateLimited.Spec.RequeueAfterSeconds)
		return ctrl.Result{RequeueAfter: time.Duration(rateLimited.Spec.RequeueAfterSeconds) * time.Second}, nil
	}

	logger.Info("Reconciliation completed successfully")
	return ctrl.Result{}, nil
}

// calculateBackoff calculates exponential backoff duration with jitter
// Base delay: 1 second, multiplier: 2, max: 300 seconds (5 minutes)
func calculateBackoff(failureCount int) int {
	const (
		baseDelay  = 1
		maxDelay   = 300
		multiplier = 2
	)

	// Calculate exponential backoff: baseDelay * (multiplier ^ (failureCount - 1))
	delay := baseDelay
	for i := 1; i < failureCount; i++ {
		delay *= multiplier
		if delay >= maxDelay {
			delay = maxDelay
			break
		}
	}

	// Add jitter (0-25% of delay)
	jitter := rand.Intn(delay/4 + 1)
	delay += jitter

	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// SetupWithManager sets up the controller with the Manager
func (r *RateLimitedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// The workqueue uses default rate limiting:
	// - Exponential backoff when items are requeued due to errors
	// - Per-item rate limiting based on the default controller rate limiter
	//
	// Our application-level backoff (using Result{RequeueAfter: duration}) provides
	// fine-grained control over retry timing based on business logic

	return ctrl.NewControllerManagedBy(mgr).
		For(&ratelimitv1alpha1.RateLimited{}).
		Complete(r)
}
