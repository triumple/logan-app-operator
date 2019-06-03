package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PythonBoot is the Schema for the pythonboots API
// +k8s:openapi-gen=true
type PythonBoot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BootSpec   `json:"spec,omitempty"`
	Status BootStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PythonBootList contains a list of PythonBoot
type PythonBootList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PythonBoot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PythonBoot{}, &PythonBootList{})
}
