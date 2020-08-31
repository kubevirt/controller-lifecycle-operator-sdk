package api

import (
	conditions "github.com/openshift/custom-resource-status/conditions/v1"
)

// Phase is the current phase of the deployment
type Phase string

const (
	// PhaseDeploying signals that the resources are being deployed
	PhaseDeploying Phase = "Deploying"

	// PhaseDeployed signals that the resources are successfully deployed
	PhaseDeployed Phase = "Deployed"

	// PhaseDeleting signals that the resources are being removed
	PhaseDeleting Phase = "Deleting"

	// PhaseDeleted signals that the resources are deleted
	PhaseDeleted Phase = "Deleted"

	// PhaseError signals that the deployment is in an error state
	PhaseError Phase = "Error"

	// PhaseUpgrading signals that the resources are being deployed
	PhaseUpgrading Phase = "Upgrading"

	// PhaseEmpty is an uninitialized phase
	PhaseEmpty Phase = ""
)

// Status represents status of a operator configuration resource; must be inlined in the operator configuration resource status
type Status struct {
	Phase Phase `json:"phase,omitempty"`
	// A list of current conditions of the resource
	Conditions []conditions.Condition `json:"conditions,omitempty" optional:"true"`
	// The version of the resource as defined by the operator
	OperatorVersion string `json:"operatorVersion,omitempty" optional:"true"`
	// The desired version of the resource
	TargetVersion string `json:"targetVersion,omitempty" optional:"true"`
	// The observed version of the resource
	ObservedVersion string `json:"observedVersion,omitempty" optional:"true"`
}

// DeepCopyInto is copying the receiver, writing into out. in must be non-nil.
func (in *Status) DeepCopyInto(out *Status) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]conditions.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}
