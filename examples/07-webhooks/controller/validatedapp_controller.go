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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webhooksv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/07-webhooks/api/v1alpha1"
)

const (
	// Finalizer for cleanup
	validatedAppFinalizer = "webhooks.examples.k8s.io/finalizer"

	// Condition types
	ConditionTypeReady      = "Ready"
	ConditionTypeDeployment = "DeploymentReady"
	ConditionTypeService    = "ServiceReady"
)

// ValidatedAppReconciler reconciles a ValidatedApp object
type ValidatedAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webhooks.examples.k8s.io,resources=validatedapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webhooks.examples.k8s.io,resources=validatedapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webhooks.examples.k8s.io,resources=validatedapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *ValidatedAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the ValidatedApp instance
	app := &webhooksv1alpha1.ValidatedApp{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if errors.IsNotFound(err) {
			log.Info("ValidatedApp resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get ValidatedApp")
		return ctrl.Result{}, err
	}

	log.Info("Reconciling ValidatedApp", "name", app.Name, "namespace", app.Namespace)

	// Check if the instance is marked to be deleted
	if app.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(app, validatedAppFinalizer) {
			// Run finalization logic
			if err := r.finalizeValidatedApp(ctx, app); err != nil {
				return ctrl.Result{}, err
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(app, validatedAppFinalizer)
			if err := r.Update(ctx, app); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(app, validatedAppFinalizer) {
		controllerutil.AddFinalizer(app, validatedAppFinalizer)
		if err := r.Update(ctx, app); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Reconcile Deployment
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		// Create new Deployment
		dep := r.deploymentForValidatedApp(app)
		log.Info("Creating new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		if err := r.Create(ctx, dep); err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			r.updateCondition(ctx, app, ConditionTypeDeployment, metav1.ConditionFalse, "DeploymentFailed", err.Error())
			return ctrl.Result{}, err
		}
		r.updateCondition(ctx, app, ConditionTypeDeployment, metav1.ConditionTrue, "DeploymentCreated", "Deployment created successfully")
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Update Deployment if spec has changed
	if r.shouldUpdateDeployment(app, deployment) {
		log.Info("Updating Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		dep := r.deploymentForValidatedApp(app)
		deployment.Spec = dep.Spec
		if err := r.Update(ctx, deployment); err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Reconcile Service if port is specified
	if app.Spec.Port != nil {
		service := &corev1.Service{}
		err := r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, service)
		if err != nil && errors.IsNotFound(err) {
			// Create new Service
			svc := r.serviceForValidatedApp(app)
			log.Info("Creating new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			if err := r.Create(ctx, svc); err != nil {
				log.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
				r.updateCondition(ctx, app, ConditionTypeService, metav1.ConditionFalse, "ServiceFailed", err.Error())
				return ctrl.Result{}, err
			}
			r.updateCondition(ctx, app, ConditionTypeService, metav1.ConditionTrue, "ServiceCreated", "Service created successfully")
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to get Service")
			return ctrl.Result{}, err
		}

		// Update Service if port has changed
		if r.shouldUpdateService(app, service) {
			log.Info("Updating Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			svc := r.serviceForValidatedApp(app)
			service.Spec.Ports = svc.Spec.Ports
			if err := r.Update(ctx, service); err != nil {
				log.Error(err, "Failed to update Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
				return ctrl.Result{}, err
			}
		}
	}

	// Update status
	if err := r.updateStatus(ctx, app, deployment); err != nil {
		log.Error(err, "Failed to update ValidatedApp status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled ValidatedApp")
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// deploymentForValidatedApp returns a Deployment object for the ValidatedApp
func (r *ValidatedAppReconciler) deploymentForValidatedApp(app *webhooksv1alpha1.ValidatedApp) *appsv1.Deployment {
	replicas := int32(1)
	if app.Spec.Replicas != nil {
		replicas = *app.Spec.Replicas
	}

	labels := map[string]string{
		"app":                          app.Name,
		"webhooks.examples.k8s.io/app": app.Name,
	}

	// Build environment variables
	env := []corev1.EnvVar{}
	for key, value := range app.Spec.Env {
		env = append(env, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	// Build security context
	securityContext := &corev1.SecurityContext{}
	if app.Spec.AllowPrivileged {
		securityContext.Privileged = &app.Spec.AllowPrivileged
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            "app",
						Image:           app.Spec.Image,
						Env:             env,
						SecurityContext: securityContext,
					}},
				},
			},
		},
	}

	// Add container port if specified
	if app.Spec.Port != nil {
		deployment.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{{
			ContainerPort: *app.Spec.Port,
			Name:          "http",
		}}
	}

	// Set ValidatedApp instance as the owner
	controllerutil.SetControllerReference(app, deployment, r.Scheme)

	return deployment
}

// serviceForValidatedApp returns a Service object for the ValidatedApp
func (r *ValidatedAppReconciler) serviceForValidatedApp(app *webhooksv1alpha1.ValidatedApp) *corev1.Service {
	labels := map[string]string{
		"app":                          app.Name,
		"webhooks.examples.k8s.io/app": app.Name,
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       *app.Spec.Port,
				TargetPort: intstr.FromInt(int(*app.Spec.Port)),
				Name:       "http",
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Set ValidatedApp instance as the owner
	controllerutil.SetControllerReference(app, service, r.Scheme)

	return service
}

// shouldUpdateDeployment checks if the Deployment needs to be updated
func (r *ValidatedAppReconciler) shouldUpdateDeployment(app *webhooksv1alpha1.ValidatedApp, deployment *appsv1.Deployment) bool {
	expectedReplicas := int32(1)
	if app.Spec.Replicas != nil {
		expectedReplicas = *app.Spec.Replicas
	}

	if *deployment.Spec.Replicas != expectedReplicas {
		return true
	}

	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		container := deployment.Spec.Template.Spec.Containers[0]
		if container.Image != app.Spec.Image {
			return true
		}
	}

	return false
}

// shouldUpdateService checks if the Service needs to be updated
func (r *ValidatedAppReconciler) shouldUpdateService(app *webhooksv1alpha1.ValidatedApp, service *corev1.Service) bool {
	if app.Spec.Port == nil {
		return false
	}

	if len(service.Spec.Ports) == 0 {
		return true
	}

	if service.Spec.Ports[0].Port != *app.Spec.Port {
		return true
	}

	return false
}

// updateStatus updates the ValidatedApp status
func (r *ValidatedAppReconciler) updateStatus(ctx context.Context, app *webhooksv1alpha1.ValidatedApp, deployment *appsv1.Deployment) error {
	// Update replicas count
	app.Status.Replicas = deployment.Status.ReadyReplicas

	// Update phase
	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
		app.Status.Phase = "Running"
		r.updateCondition(ctx, app, ConditionTypeReady, metav1.ConditionTrue, "AllReplicasReady", "All replicas are ready")
	} else {
		app.Status.Phase = "Pending"
		r.updateCondition(ctx, app, ConditionTypeReady, metav1.ConditionFalse, "WaitingForReplicas", fmt.Sprintf("Waiting for replicas: %d/%d ready", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas))
	}

	// Update last validated timestamp
	now := metav1.Now()
	app.Status.LastValidated = &now

	return r.Status().Update(ctx, app)
}

// updateCondition updates a condition in the ValidatedApp status
func (r *ValidatedAppReconciler) updateCondition(ctx context.Context, app *webhooksv1alpha1.ValidatedApp, conditionType string, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		ObservedGeneration: app.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}

	meta.SetStatusCondition(&app.Status.Conditions, condition)
}

// finalizeValidatedApp performs cleanup before the resource is deleted
func (r *ValidatedAppReconciler) finalizeValidatedApp(ctx context.Context, app *webhooksv1alpha1.ValidatedApp) error {
	log := log.FromContext(ctx)
	log.Info("Finalizing ValidatedApp", "name", app.Name, "namespace", app.Namespace)

	// Cleanup logic here (if needed)
	// For example, clean up external resources, notify other services, etc.

	log.Info("Successfully finalized ValidatedApp")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ValidatedAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webhooksv1alpha1.ValidatedApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
