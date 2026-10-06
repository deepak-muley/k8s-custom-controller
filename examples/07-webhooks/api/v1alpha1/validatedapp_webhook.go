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

package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var validatedapplog = logf.Log.WithName("validatedapp-resource")

// Allowed registries for container images
var allowedRegistries = []string{
	"docker.io",
	"gcr.io",
	"ghcr.io",
	"quay.io",
	"registry.k8s.io",
}

// SetupWebhookWithManager sets up the webhook with the Manager
func (r *ValidatedApp) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-webhooks-examples-k8s-io-v1alpha1-validatedapp,mutating=true,failurePolicy=fail,sideEffects=None,groups=webhooks.examples.k8s.io,resources=validatedapps,verbs=create;update,versions=v1alpha1,name=mvalidatedapp.kb.io,admissionReviewVersions=v1

var _ admission.CustomDefaulter = &ValidatedApp{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (r *ValidatedApp) Default(ctx context.Context, obj runtime.Object) error {
	r = obj.(*ValidatedApp)
	validatedapplog.Info("default", "name", r.Name)

	// Set default replicas to 1 if not specified
	if r.Spec.Replicas == nil {
		defaultReplicas := int32(1)
		r.Spec.Replicas = &defaultReplicas
		validatedapplog.Info("setting default replicas", "replicas", 1)
	}

	// Set default port to 8080 if not specified
	if r.Spec.Port == nil {
		defaultPort := int32(8080)
		r.Spec.Port = &defaultPort
		validatedapplog.Info("setting default port", "port", 8080)
	}

	// Add docker.io prefix if image doesn't have a registry
	if !strings.Contains(r.Spec.Image, "/") {
		r.Spec.Image = "docker.io/library/" + r.Spec.Image
		validatedapplog.Info("adding default registry prefix", "image", r.Spec.Image)
	} else if strings.Count(r.Spec.Image, "/") == 1 && !strings.Contains(strings.Split(r.Spec.Image, "/")[0], ".") {
		// If it's like "user/image" without registry, add docker.io
		r.Spec.Image = "docker.io/" + r.Spec.Image
		validatedapplog.Info("adding docker.io prefix", "image", r.Spec.Image)
	}

	return nil
}

// +kubebuilder:webhook:path=/validate-webhooks-examples-k8s-io-v1alpha1-validatedapp,mutating=false,failurePolicy=fail,sideEffects=None,groups=webhooks.examples.k8s.io,resources=validatedapps,verbs=create;update;delete,versions=v1alpha1,name=vvalidatedapp.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &ValidatedApp{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *ValidatedApp) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	r = obj.(*ValidatedApp)
	validatedapplog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	// Validate the spec
	if err := r.validateValidatedAppSpec(); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, allErrs.ToAggregate()
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *ValidatedApp) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	r = newObj.(*ValidatedApp)
	validatedapplog.Info("validate update", "name", r.Name)

	var allErrs field.ErrorList
	var warnings admission.Warnings

	oldApp, ok := oldObj.(*ValidatedApp)
	if !ok {
		return nil, fmt.Errorf("expected ValidatedApp but got %T", oldObj)
	}

	// Validate the spec
	if err := r.validateValidatedAppSpec(); err != nil {
		allErrs = append(allErrs, err...)
	}

	// Check if replicas are being scaled down dramatically
	if oldApp.Spec.Replicas != nil && r.Spec.Replicas != nil {
		if *r.Spec.Replicas < *oldApp.Spec.Replicas/2 && *oldApp.Spec.Replicas > 2 {
			warnings = append(warnings, fmt.Sprintf(
				"Warning: Scaling down from %d to %d replicas (>50%% reduction)",
				*oldApp.Spec.Replicas, *r.Spec.Replicas))
		}
	}

	// Warn if image is changing
	if oldApp.Spec.Image != r.Spec.Image {
		warnings = append(warnings, fmt.Sprintf(
			"Warning: Changing image from %s to %s",
			oldApp.Spec.Image, r.Spec.Image))
	}

	// Warn if privileged mode is being enabled
	if !oldApp.Spec.AllowPrivileged && r.Spec.AllowPrivileged {
		warnings = append(warnings, "Warning: Enabling privileged mode - ensure this is intended")
	}

	if len(allErrs) == 0 {
		return warnings, nil
	}

	return warnings, allErrs.ToAggregate()
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *ValidatedApp) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	r = obj.(*ValidatedApp)
	validatedapplog.Info("validate delete", "name", r.Name)

	var warnings admission.Warnings

	// Warn if deleting an app with many replicas
	if r.Spec.Replicas != nil && *r.Spec.Replicas > 3 {
		warnings = append(warnings, fmt.Sprintf(
			"Warning: Deleting app with %d replicas", *r.Spec.Replicas))
	}

	// You could add additional checks here, for example:
	// - Check if app is in critical namespace
	// - Check if app has certain labels indicating it's important
	// - Check if app is referenced by other resources

	return warnings, nil
}

// validateValidatedAppSpec validates the ValidatedApp spec
func (r *ValidatedApp) validateValidatedAppSpec() field.ErrorList {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	// Validate image registry
	if err := r.validateImageRegistry(); err != nil {
		allErrs = append(allErrs, field.Invalid(
			specPath.Child("image"),
			r.Spec.Image,
			err.Error()))
	}

	// Validate image tag is not 'latest' in production
	if strings.HasSuffix(r.Spec.Image, ":latest") {
		allErrs = append(allErrs, field.Invalid(
			specPath.Child("image"),
			r.Spec.Image,
			"image tag 'latest' is not allowed for production deployments"))
	}

	// Validate image has a tag
	if !strings.Contains(r.Spec.Image, ":") {
		allErrs = append(allErrs, field.Invalid(
			specPath.Child("image"),
			r.Spec.Image,
			"image must include a tag (e.g., nginx:1.21)"))
	}

	// Validate privileged mode
	if r.Spec.AllowPrivileged {
		// Check if there's a specific annotation allowing privileged mode
		if r.Annotations["security.k8s.io/allow-privileged"] != "true" {
			allErrs = append(allErrs, field.Forbidden(
				specPath.Child("allowPrivileged"),
				"privileged mode requires annotation 'security.k8s.io/allow-privileged: true'"))
		}
	}

	// Validate environment variables don't contain secrets
	for key, value := range r.Spec.Env {
		if strings.Contains(strings.ToLower(key), "password") ||
			strings.Contains(strings.ToLower(key), "secret") ||
			strings.Contains(strings.ToLower(key), "token") {
			allErrs = append(allErrs, field.Invalid(
				specPath.Child("env").Key(key),
				value,
				"sensitive data should be stored in Secrets, not in environment variables"))
		}
	}

	// Validate replicas with privileged mode
	if r.Spec.AllowPrivileged && r.Spec.Replicas != nil && *r.Spec.Replicas > 1 {
		allErrs = append(allErrs, field.Invalid(
			specPath.Child("replicas"),
			*r.Spec.Replicas,
			"privileged containers should not be scaled beyond 1 replica"))
	}

	return allErrs
}

// validateImageRegistry checks if the image is from an allowed registry
func (r *ValidatedApp) validateImageRegistry() error {
	image := r.Spec.Image

	// Extract registry from image
	parts := strings.Split(image, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid image format: %s", image)
	}

	registry := parts[0]

	// Check if registry is in allowed list
	for _, allowed := range allowedRegistries {
		if strings.HasPrefix(registry, allowed) {
			return nil
		}
	}

	return fmt.Errorf("image registry '%s' is not in allowed list: %v", registry, allowedRegistries)
}
