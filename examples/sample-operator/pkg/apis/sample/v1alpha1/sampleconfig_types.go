package v1alpha1

import (
	sdkapi "github.com/kubevirt/controller-lifecycle-operator-sdk/pkg/sdk/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SampleConfigSpec defines the desired state of SampleConfig
type SampleConfigSpec struct {
}

// SampleConfigStatus defines the observed state of SampleConfig
type SampleConfigStatus struct {
	sdkapi.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SampleConfig is the Schema for the sampleconfigs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sampleconfigs,scope=Namespaced
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
