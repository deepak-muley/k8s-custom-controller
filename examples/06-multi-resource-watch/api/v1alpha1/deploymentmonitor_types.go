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

// DeploymentMonitorSpec defines the desired state of DeploymentMonitor
type DeploymentMonitorSpec struct {
	// DeploymentName is the name of the Deployment to monitor
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	DeploymentName string `json:"deploymentName"`

	// Namespace is the namespace of the Deployment to monitor
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Namespace string `json:"namespace"`
}

// ReplicaSetInfo contains information about a ReplicaSet
type ReplicaSetInfo struct {
	// Name of the ReplicaSet
	Name string `json:"name"`

	// Replicas is the desired number of replicas
	Replicas int32 `json:"replicas"`

	// ReadyReplicas is the number of ready replicas
	ReadyReplicas int32 `json:"readyReplicas"`

	// AvailableReplicas is the number of available replicas
	AvailableReplicas int32 `json:"availableReplicas"`

	// Revision is the deployment revision
	Revision string `json:"revision,omitempty"`
}

// DeploymentMonitorStatus defines the observed state of DeploymentMonitor
type DeploymentMonitorStatus struct {
	// DeploymentFound indicates if the deployment was found
	// +optional
	DeploymentFound bool `json:"deploymentFound,omitempty"`

	// DesiredReplicas is the desired number of replicas from the Deployment
	// +optional
	DesiredReplicas int32 `json:"desiredReplicas,omitempty"`

	// AvailableReplicas is the number of available replicas
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// ReadyReplicas is the number of ready replicas
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// ReplicaSets contains information about the ReplicaSets owned by the Deployment
	// +optional
	ReplicaSets []ReplicaSetInfo `json:"replicaSets,omitempty"`

	// TotalReplicaSets is the total number of ReplicaSets
	// +optional
	TotalReplicaSets int `json:"totalReplicaSets,omitempty"`

	// LastUpdated is the timestamp when the status was last updated
	// +optional
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed DeploymentMonitor
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=dm
// +kubebuilder:printcolumn:name="Deployment",type=string,JSONPath=`.spec.deploymentName`
// +kubebuilder:printcolumn:name="Found",type=boolean,JSONPath=`.status.deploymentFound`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.status.desiredReplicas`
// +kubebuilder:printcolumn:name="Available",type=integer,JSONPath=`.status.availableReplicas`
// +kubebuilder:printcolumn:name="ReplicaSets",type=integer,JSONPath=`.status.totalReplicaSets`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// DeploymentMonitor is the Schema for the deploymentmonitors API
type DeploymentMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentMonitorSpec   `json:"spec,omitempty"`
	Status DeploymentMonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DeploymentMonitorList contains a list of DeploymentMonitor
type DeploymentMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeploymentMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeploymentMonitor{}, &DeploymentMonitorList{})
}
