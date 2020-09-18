package controller

import (
	"kubevirt.io/controller-lifecycle-operator-sdk/examples/sample-operator/pkg/controller/sampleconfig"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, sampleconfig.Add)
}
