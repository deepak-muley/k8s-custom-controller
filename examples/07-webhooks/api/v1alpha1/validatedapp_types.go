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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidatedAppSpec defines the desired state of ValidatedApp
type ValidatedAppSpec struct {
	// Image is the container image to run
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// Replicas is the number of replicas to run
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	Replicas *int32 `json:"replicas,omitempty"`

	// Port is the container port to expose
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port *int32 `json:"port,omitempty"`

	// Environment variables for the container
	// +optional
	Env map[string]string `json:"env,omitempty"`

	// AllowPrivileged indicates whether privileged mode is allowed
	// +optional
	AllowPrivileged bool `json:"allowPrivileged,omitempty"`
}

// ValidatedAppStatus defines the observed state of ValidatedApp
type ValidatedAppStatus struct {
	// Phase represents the current phase of the application
	// +optional
	Phase string `json:"phase,omitempty"`

	// Replicas is the number of replicas currently running
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Conditions represent the latest available observations of the app's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastValidated is the timestamp of the last successful validation
	// +optional
	LastValidated *metav1.Time `json:"lastValidated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=vapp
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ValidatedApp is the Schema for the validatedapps API
type ValidatedApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ValidatedAppSpec   `json:"spec,omitempty"`
	Status ValidatedAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ValidatedAppList contains a list of ValidatedApp
type ValidatedAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ValidatedApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ValidatedApp{}, &ValidatedAppList{})
}
