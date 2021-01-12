module kubevirt.io/controller-lifecycle-operator-sdk

go 1.14

require (
	github.com/appscode/jsonpatch v0.0.0-20190108182946-7c0e3b262f30
	github.com/blang/semver v3.5.1+incompatible
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/openshift/custom-resource-status v0.0.0-20200602122900-c002fd1547ca
	github.com/prometheus/client_golang v1.1.0 // indirect
	go.uber.org/multierr v1.3.0 // indirect
	golang.org/x/tools v0.0.0-20200616195046-dc31b401abb5
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.4.0
)

replace sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v1.0.0
