package sampleconfig_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"kubevirt.io/controller-lifecycle-operator-sdk/examples/sample-operator/pkg/apis/sample/v1alpha1"
	"kubevirt.io/controller-lifecycle-operator-sdk/examples/sample-operator/pkg/controller/sampleconfig"
	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/api"
)

var _ = Describe("Controller CR manager", func() {
	manager := sampleconfig.CrManager{}
	It("should create CR", func() {
		obj := manager.Create()

		Expect(obj).To(BeAssignableToTypeOf(&v1alpha1.SampleConfig{}))
	})

	It("should return status", func() {
		status := sdkapi.Status{Phase: sdkapi.PhaseDeploying}
		config := v1alpha1.SampleConfig{Status: v1alpha1.SampleConfigStatus{Status: status}}

		statusFromCR := manager.Status(&config)

		Expect(*statusFromCR).To(BeEquivalentTo(status))
	})
})
