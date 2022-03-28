package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/api"
)

// SampleConfigSpec defines the desired state of SampleConfig
type SampleConfigSpec struct {
	// Rules on which nodes controller pod(s) will be scheduled
	// +optional
	Infra sdkapi.NodePlacement `json:"infra,omitempty"`
}

// SampleConfigStatus defines the observed state of SampleConfig
type SampleConfigStatus struct {
	sdkapi.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SampleConfig is the Schema for the sampleconfigs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sampleconfigs,scope=Cluster
// +genclient
// +genclient:nonNamespaced
type SampleConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SampleConfigSpec   `json:"spec,omitempty"`
	Status SampleConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// SampleConfigList contains a list of SampleConfig
type SampleConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SampleConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SampleConfig{}, &SampleConfigList{})
}
