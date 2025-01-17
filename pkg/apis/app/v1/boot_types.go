package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Boot is the common Schema for the all boot types API
// +k8s:openapi-gen=true
type Boot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec contains the desired behavior of the Boot
	Spec BootSpec `json:"spec,omitempty"`
	// status contains the last observed state of the BootStatus
	Status BootStatus `json:"status,omitempty"`

	BootType string `json:"bootType"`
	AppKey   string `json:"appKey"`
}

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BootSpec defines the desired state of Boot for specified types, as JavaBoot/PhpBoot/PythonBoot/NodeJSBoot
// +k8s:openapi-gen=true
type BootSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// Image is the app container' image. Image must not have a tag version.
	Image string `json:"image"`
	// Version is the app container's image version.
	Version string `json:"version"`
	// Replicas is the number of desired replicas.
	// This is a pointer to distinguish between explicit zero and unspecified.
	// Defaults to 1.
	Replicas *int32 `json:"replicas,omitempty"`
	// Env is list of environment variables to set in the app container.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// Port that are exposed by the app container
	Port int32 `json:"port,omitempty"`
	// Reserved, not used. for latter use
	SubDomain string `json:"subDomain,omitempty"`
	// Health is check path for the app container.
	Health *string `json:"health,omitempty"`
	// Readiness is a readiness check path for the app container.
	Readiness *string `json:"readiness,omitempty"`
	// Prometheus will scrape metrics from the service, default is `true`
	Prometheus string `json:"prometheus"`
	// Resources is the compute resource requirements for the app container
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Command is command for boot's container. If empty, will use image's ENTRYPOINT, specified here if needed override.
	Command []string `json:"command,omitempty"`
	// SessionAffinity is SessionAffinity for boot's created service. If empty, will not set
	SessionAffinity string `json:"sessionAffinity,omitempty"`
	// NodePort will expose the service on each node’s IP at a random port, default is ``
	NodePort string `json:"nodePort,omitempty"`
	// pvc is list of PersistentVolumeClaim to set in the app container.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Pvc []PersistentVolumeClaimMount `json:"pvc,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// BootStatus defines the observed state of Boot for specified types, as JavaBoot/PhpBoot/PythonBoot/NodeJSBoot
// +k8s:openapi-gen=true
type BootStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	Type     string `json:"type,omitempty"`
	Deploy   string `json:"deploy,omitempty"`
	Services string `json:"services,omitempty"`
}

// PersistentVolumeClaimMount defines the Boot match a PersistentVolumeClaim
// +k8s:openapi-gen=true
type PersistentVolumeClaimMount struct {
	// This must match the Name of a PersistentVolumeClaim.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Mounted read-only if true, read-write otherwise (false or unspecified).
	// Defaults to false.
	// +optional
	ReadOnly bool `json:"readOnly,omitempty" protobuf:"varint,2,opt,name=readOnly"`
	// Path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	MountPath string `json:"mountPath" protobuf:"bytes,3,opt,name=mountPath"`
}
