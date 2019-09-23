package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BootRevision is the Schema for the bootrevisions API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type BootRevision struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec contains the desired behavior of the Boot
	Spec BootSpec `json:"spec,omitempty"`
	// status contains the last observed state of the BootStatus
	Status BootStatus `json:"status,omitempty"`

	BootType string `json:"bootType"`
	AppKey   string `json:"appKey"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BootRevisionList contains a list of BootRevision
type BootRevisionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BootRevision `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BootRevision{}, &BootRevisionList{})
}
