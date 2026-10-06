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

// FilteredConfigSpec defines the desired state of FilteredConfig
type FilteredConfigSpec struct {
	// Data holds the configuration data
	// +optional
	Data map[string]string `json:"data,omitempty"`

	// WatchLabels defines which labels to filter on
	// +optional
	WatchLabels map[string]string `json:"watchLabels,omitempty"`
}

// FilteredConfigStatus defines the observed state of FilteredConfig
type FilteredConfigStatus struct {
	// ProcessedCount tracks how many times this config was processed
	// +optional
	ProcessedCount int32 `json:"processedCount,omitempty"`

	// LastProcessed is the timestamp when last processed
	// +optional
	LastProcessed metav1.Time `json:"lastProcessed,omitempty"`

	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=fc
// +kubebuilder:printcolumn:name="Processed",type=integer,JSONPath=`.status.processedCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// FilteredConfig is the Schema for the filteredconfigs API
type FilteredConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FilteredConfigSpec   `json:"spec,omitempty"`
	Status FilteredConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FilteredConfigList contains a list of FilteredConfig
type FilteredConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FilteredConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FilteredConfig{}, &FilteredConfigList{})
}
