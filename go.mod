module kubevirt.io/controller-lifecycle-operator-sdk

go 1.14

require (
	github.com/appscode/jsonpatch v1.0.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/evanphx/json-patch v5.6.0+incompatible
	github.com/go-logr/logr v1.2.4
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.27.10
	github.com/openshift/custom-resource-status v1.1.2
	golang.org/x/tools v0.9.3
	k8s.io/api v0.28.3
	k8s.io/apiextensions-apiserver v0.28.3
	k8s.io/apimachinery v0.28.3
	k8s.io/client-go v0.28.3
	kubevirt.io/controller-lifecycle-operator-sdk/api v0.0.0-00010101000000-000000000000
	sigs.k8s.io/controller-runtime v0.15.3
	sigs.k8s.io/controller-tools v0.8.0
)

replace sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.0

replace kubevirt.io/controller-lifecycle-operator-sdk/api => ./api
