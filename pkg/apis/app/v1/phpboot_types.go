package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PhpBoot is the Schema for the phpboots API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type PhpBoot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BootSpec   `json:"spec,omitempty"`
	Status BootStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PhpBootList contains a list of PhpBoot
type PhpBootList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PhpBoot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PhpBoot{}, &PhpBootList{})
}
