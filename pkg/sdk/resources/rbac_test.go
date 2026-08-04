package resources

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	rbacv1 "k8s.io/api/rbac/v1"
)

var _ = Describe("RBAC Resources", func() {
	var builder ResourceBuilder

	BeforeEach(func() {
		builder = NewResourceBuilder(
			map[string]string{"common": "label"},
			map[string]string{"operator": "label"},
		)
	})

	It("should treat empty rules as nil for Role", func() {
		role := builder.CreateRole("test-role", []rbacv1.PolicyRule{})
		Expect(role.Rules).To(BeNil())
	})

	It("should treat empty rules as nil for ClusterRole", func() {
		labels := map[string]string{"test": "label"}
		clusterRole := CreateClusterRole("test-clusterrole", []rbacv1.PolicyRule{}, labels)
		Expect(clusterRole.Rules).To(BeNil())
	})
})
