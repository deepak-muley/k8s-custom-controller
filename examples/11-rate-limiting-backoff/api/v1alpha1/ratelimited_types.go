package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RateLimitedSpec defines the desired state of RateLimited
type RateLimitedSpec struct {
	// ProcessingTime is the time in seconds to simulate processing
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=300
	// +optional
	ProcessingTime int `json:"processingTime,omitempty"`

	// FailureRatePercent is the probability of failure as a percentage (0 to 100)
	// Used to simulate failures and trigger rate limiting
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +optional
	FailureRatePercent int `json:"failureRatePercent,omitempty"`

	// MaxRetries is the maximum number of retries before giving up
	// +kubebuilder:validation:Minimum=0
	// +optional
	MaxRetries int `json:"maxRetries,omitempty"`

	// RequeueAfterSeconds is the time to wait before requeuing
	// +kubebuilder:validation:Minimum=0
	// +optional
	RequeueAfterSeconds int `json:"requeueAfterSeconds,omitempty"`
}

// RateLimitedStatus defines the observed state of RateLimited
type RateLimitedStatus struct {
	// ProcessingCount is the number of times this resource has been processed
	ProcessingCount int `json:"processingCount,omitempty"`

	// FailureCount is the number of times processing has failed
	FailureCount int `json:"failureCount,omitempty"`

	// LastProcessedTime is the last time the resource was processed
	LastProcessedTime *metav1.Time `json:"lastProcessedTime,omitempty"`

	// LastFailureTime is the last time processing failed
	LastFailureTime *metav1.Time `json:"lastFailureTime,omitempty"`

	// LastFailureReason is the reason for the last failure
	LastFailureReason string `json:"lastFailureReason,omitempty"`

	// Phase represents the current phase of processing
	// +kubebuilder:validation:Enum=Pending;Processing;Completed;Failed;RateLimited
	Phase string `json:"phase,omitempty"`

	// BackoffDuration is the current backoff duration in seconds
	BackoffDuration int `json:"backoffDuration,omitempty"`

	// Message provides additional information about the current state
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=rl
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="ProcessingCount",type=integer,JSONPath=`.status.processingCount`
// +kubebuilder:printcolumn:name="FailureCount",type=integer,JSONPath=`.status.failureCount`
// +kubebuilder:printcolumn:name="BackoffDuration",type=integer,JSONPath=`.status.backoffDuration`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// RateLimited is the Schema for the ratelimiteds API
type RateLimited struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RateLimitedSpec   `json:"spec,omitempty"`
	Status RateLimitedStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RateLimitedList contains a list of RateLimited
type RateLimitedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimited `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RateLimited{}, &RateLimitedList{})
}
