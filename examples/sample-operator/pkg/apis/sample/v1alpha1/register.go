// NOTE: Boilerplate only.  Ignore this file.

// Package v1alpha1 contains API Schema definitions for the sample v1alpha1 API group
// +k8s:deepcopy-gen=package,register
// +groupName=sample.lifecycle.kubevirt.io
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "sample.lifecycle.kubevirt.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme may be used to add all resources defined in the project to a Scheme
	AddToScheme = SchemeBuilder.AddToScheme
)
