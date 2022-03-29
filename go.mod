module kubevirt.io/controller-lifecycle-operator-sdk

go 1.14

require (
	github.com/appscode/jsonpatch v1.0.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/evanphx/json-patch v4.12.0+incompatible
	github.com/go-logr/logr v1.2.3
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	github.com/openshift/custom-resource-status v1.1.2
	golang.org/x/tools v0.1.10
	k8s.io/api v0.23.5
	k8s.io/apiextensions-apiserver v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v0.23.5
	kubevirt.io/controller-lifecycle-operator-sdk/api v0.0.0-00010101000000-000000000000
	sigs.k8s.io/controller-runtime v0.11.1
	sigs.k8s.io/controller-tools v0.8.0
)

replace sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.0

replace kubevirt.io/controller-lifecycle-operator-sdk/api => ./api
