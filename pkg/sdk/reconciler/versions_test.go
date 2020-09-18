package reconciler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/reconciler"
)

var _ = Describe("Version upgrade", func() {
	DescribeTable("should be detected", func(currentVersion, targetVersion string) {
		upgrade, err := reconciler.ShouldTakeUpdatePath(targetVersion, currentVersion, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(upgrade).To(BeTrue())
	},
		Entry("patch upgrade", "0.0.1", "0.0.2"),
		Entry("minor upgrade", "0.0.1", "0.1.0"),
		Entry("major upgrade", "0.0.1", "1.0.0"),
		Entry("major from minor upgrade", "0.1.0", "1.0.0"),

		Entry("v-prefixed path upgrade", "v0.0.1", "v0.0.2"),
	)

	It("should be detected for empty current version", func() {
		upgrade, err := reconciler.ShouldTakeUpdatePath("0.0.1", "", false)

		Expect(err).ToNot(HaveOccurred())
		Expect(upgrade).To(BeTrue())
	})

	It("should not be detected for deploying", func() {
		upgrade, err := reconciler.ShouldTakeUpdatePath("0.0.1", "0.0.1", true)

		Expect(err).ToNot(HaveOccurred())
		Expect(upgrade).To(BeFalse())
	})

	DescribeTable("should not be detected for same version", func(currentVersion, targetVersion string) {
		upgrade, err := reconciler.ShouldTakeUpdatePath(targetVersion, currentVersion, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(upgrade).To(BeFalse())
	},
		Entry("in same notation", "0.0.1", "0.0.1"),
		Entry("in mixed notation I", "0.0.1", "v0.0.1"),
		Entry("in mixed notation II", "v0.0.1", "0.0.1"),
	)
})

var _ = Describe("Version downgrade", func() {

	It("should be reported as error", func() {
		upgrade, err := reconciler.ShouldTakeUpdatePath("0.0.1", "0.0.2", false)

		Expect(err).To(HaveOccurred())
		Expect(upgrade).To(BeFalse())
	})
})
