package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExternalHandlerSpec defines the desired state of ExternalHandler
type ExternalHandlerSpec struct {
	// EventFilter defines which external events to process
	// +optional
	EventFilter string `json:"eventFilter,omitempty"`

	// MaxEvents is the maximum number of events to process
	// +optional
	// +kubebuilder:default=100
	MaxEvents int `json:"maxEvents,omitempty"`

	// ProcessingMode defines how to process events (sequential or parallel)
	// +optional
	// +kubebuilder:default="sequential"
	// +kubebuilder:validation:Enum=sequential;parallel
	ProcessingMode string `json:"processingMode,omitempty"`
}

// ExternalEvent represents an event received from external source
type ExternalEvent struct {
	// EventID is the unique identifier for the event
	EventID string `json:"eventID"`

	// EventType is the type of the event
	EventType string `json:"eventType"`

	// Timestamp when the event was received
	Timestamp metav1.Time `json:"timestamp"`

	// Message contains event details
	Message string `json:"message"`

	// Source is the origin of the event
	Source string `json:"source"`
}

// ExternalHandlerStatus defines the observed state of ExternalHandler
type ExternalHandlerStatus struct {
	// TotalEventsProcessed is the count of events processed
	TotalEventsProcessed int64 `json:"totalEventsProcessed"`

	// LastProcessedEvent contains details of the most recent event
	// +optional
	LastProcessedEvent *ExternalEvent `json:"lastProcessedEvent,omitempty"`

	// RecentEvents contains the last 5 events processed
	// +optional
	RecentEvents []ExternalEvent `json:"recentEvents,omitempty"`

	// IsActive indicates if the handler is actively processing events
	IsActive bool `json:"isActive"`

	// LastReconcileTime is the last time reconciliation occurred
	// +optional
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`

	// Conditions represent the latest available observations of the handler's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=eh
// +kubebuilder:printcolumn:name="Events",type=integer,JSONPath=`.status.totalEventsProcessed`
// +kubebuilder:printcolumn:name="Active",type=boolean,JSONPath=`.status.isActive`
// +kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.spec.processingMode`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ExternalHandler is the Schema for the externalhandlers API
type ExternalHandler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExternalHandlerSpec   `json:"spec,omitempty"`
	Status ExternalHandlerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ExternalHandlerList contains a list of ExternalHandler
type ExternalHandlerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalHandler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExternalHandler{}, &ExternalHandlerList{})
}
