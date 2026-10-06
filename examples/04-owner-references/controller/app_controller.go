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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ownershipv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/04-owner-references/api/v1alpha1"
)

// AppReconciler reconciles an App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ownership.examples.k8s.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ownership.examples.k8s.io,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ownership.examples.k8s.io,resources=apps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	app := &ownershipv1alpha1.App{}
	err := r.Get(ctx, req.NamespacedName, app)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("App resource not found. Child resources will be garbage collected")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get App")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling App", "name", app.Name, "namespace", app.Namespace)

	// Create/Update ConfigMap
	if err := r.reconcileConfigMap(ctx, app); err != nil {
		return ctrl.Result{}, err
	}

	// Create/Update Deployment
	if err := r.reconcileDeployment(ctx, app); err != nil {
		return ctrl.Result{}, err
	}

	// Create/Update Service
	if err := r.reconcileService(ctx, app); err != nil {
		return ctrl.Result{}, err
	}

	// Update status
	app.Status.ObservedGeneration = app.Generation
	app.Status.DeploymentReady = true
	app.Status.ServiceCreated = true

	if err := r.Status().Update(ctx, app); err != nil {
		logger.Error(err, "Failed to update App status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled App")
	return ctrl.Result{}, nil
}

func (r *AppReconciler) reconcileConfigMap(ctx context.Context, app *ownershipv1alpha1.App) error {
	logger := log.FromContext(ctx)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name + "-config",
			Namespace: app.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error {
		cm.Data = map[string]string{
			"app.name": app.Name,
			"port":     string(rune(app.Spec.Port)),
		}
		return controllerutil.SetControllerReference(app, cm, r.Scheme)
	})

	if err != nil {
		logger.Error(err, "Failed to reconcile ConfigMap")
		return err
	}

	logger.Info("ConfigMap reconciled", "name", cm.Name)
	return nil
}

func (r *AppReconciler) reconcileDeployment(ctx context.Context, app *ownershipv1alpha1.App) error {
	logger := log.FromContext(ctx)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		replicas := app.Spec.Replicas
		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": app.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": app.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: app.Spec.Image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: app.Spec.Port},
							},
						},
					},
				},
			},
		}
		return controllerutil.SetControllerReference(app, deployment, r.Scheme)
	})

	if err != nil {
		logger.Error(err, "Failed to reconcile Deployment")
		return err
	}

	logger.Info("Deployment reconciled", "name", deployment.Name)
	return nil
}

func (r *AppReconciler) reconcileService(ctx context.Context, app *ownershipv1alpha1.App) error {
	logger := log.FromContext(ctx)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		service.Spec = corev1.ServiceSpec{
			Selector: map[string]string{"app": app.Name},
			Ports: []corev1.ServicePort{
				{
					Port:       app.Spec.Port,
					TargetPort: intstr.FromInt(int(app.Spec.Port)),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		}
		return controllerutil.SetControllerReference(app, service, r.Scheme)
	})

	if err != nil {
		logger.Error(err, "Failed to reconcile Service")
		return err
	}

	logger.Info("Service reconciled", "name", service.Name)
	return nil
}

func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ownershipv1alpha1.App{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Named("app").
		Complete(r)
}
