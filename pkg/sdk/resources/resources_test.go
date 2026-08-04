package resources

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("CreateOperatorDeployment", func() {
	It("should use distinct maps for selector and pod template labels", func() {
		builder := NewResourceBuilder(
			map[string]string{"common": "label"},
			map[string]string{"operator": "label"},
		)

		deployment := builder.CreateOperatorDeployment(
			"test-operator", "test-ns",
			"app", "my-operator",
			"sa", 1, corev1.PodSpec{},
		)

		selectorLabels := deployment.Spec.Selector.MatchLabels
		podLabels := deployment.Spec.Template.Labels

		Expect(selectorLabels).To(Equal(podLabels))

		podLabels["extra-pod-label"] = "should-not-leak"

		Expect(selectorLabels).ToNot(HaveKey("extra-pod-label"))
	})
})
