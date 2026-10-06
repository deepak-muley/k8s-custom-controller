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

// IndexedResourceSpec defines the desired state of IndexedResource
type IndexedResourceSpec struct {
	// Category is a custom field that will be indexed for fast lookups
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Category string `json:"category"`

	// Owner is another field that can be indexed
	// +optional
	Owner string `json:"owner,omitempty"`

	// Priority defines the importance of this resource
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:default=5
	Priority int `json:"priority,omitempty"`

	// Tags are additional metadata that can be indexed
	// +optional
	Tags []string `json:"tags,omitempty"`

	// Data is arbitrary payload data
	// +optional
	Data string `json:"data,omitempty"`
}

// IndexedResourceStatus defines the observed state of IndexedResource
type IndexedResourceStatus struct {
	// Phase indicates the current lifecycle phase of the resource
	// +optional
	Phase string `json:"phase,omitempty"`

	// LastProcessed is the timestamp of the last successful reconciliation
	// +optional
	LastProcessed metav1.Time `json:"lastProcessed,omitempty"`

	// RelatedResources is a list of related resources found via index lookups
	// +optional
	RelatedResources []string `json:"relatedResources,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=idxres
// +kubebuilder:printcolumn:name="Category",type=string,JSONPath=`.spec.category`
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.spec.owner`
// +kubebuilder:printcolumn:name="Priority",type=integer,JSONPath=`.spec.priority`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// IndexedResource is the Schema for the indexedresources API
type IndexedResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IndexedResourceSpec   `json:"spec,omitempty"`
	Status IndexedResourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IndexedResourceList contains a list of IndexedResource
type IndexedResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IndexedResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IndexedResource{}, &IndexedResourceList{})
}
