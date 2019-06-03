package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JavaBoot is the Schema for the javaboots API
// +k8s:openapi-gen=true
type JavaBoot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BootSpec   `json:"spec,omitempty"`
	Status BootStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JavaBootList contains a list of JavaBoot
type JavaBootList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JavaBoot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JavaBoot{}, &JavaBootList{})
}
