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

// MonitoredSpec defines the desired state of Monitored
type MonitoredSpec struct {
	// Target is the resource to monitor
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Target string `json:"target"`

	// Threshold defines the alert threshold
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=80
	// +optional
	Threshold int32 `json:"threshold,omitempty"`

	// Interval defines the monitoring interval in seconds
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=3600
	// +kubebuilder:default=60
	// +optional
	Interval int32 `json:"interval,omitempty"`
}

// MonitoredStatus defines the observed state of Monitored
type MonitoredStatus struct {
	// State represents the current state of monitoring
	// +kubebuilder:validation:Enum=Monitoring;Alerting;Healthy;Failed
	// +optional
	State string `json:"state,omitempty"`

	// LastCheckTime is the timestamp of the last health check
	// +optional
	LastCheckTime metav1.Time `json:"lastCheckTime,omitempty"`

	// ReconcileCount tracks the number of reconciliations
	// +optional
	ReconcileCount int64 `json:"reconcileCount,omitempty"`

	// CurrentValue is the current monitored value
	// +optional
	CurrentValue int32 `json:"currentValue,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Monitored
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=mon
// +kubebuilder:printcolumn:name="Target",type=string,JSONPath=`.spec.target`
// +kubebuilder:printcolumn:name="Threshold",type=integer,JSONPath=`.spec.threshold`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Reconciles",type=integer,JSONPath=`.status.reconcileCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Monitored is the Schema for the monitoreds API
type Monitored struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MonitoredSpec   `json:"spec,omitempty"`
	Status MonitoredStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MonitoredList contains a list of Monitored
type MonitoredList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Monitored `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Monitored{}, &MonitoredList{})
}
