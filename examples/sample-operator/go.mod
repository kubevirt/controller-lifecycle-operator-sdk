module github.com/kubevirt/controller-lifecycle-operator-sdk/examples/sample-operator

go 1.13

require (
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kubevirt/controller-lifecycle-operator-sdk v0.0.4
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/tools v0.0.0-20200616195046-dc31b401abb5
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.18.6
	kubevirt.io/client-go v0.32.0
	sigs.k8s.io/controller-runtime v0.6.2
)

replace k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by O

replace vbom.ml/util => github.com/fvbommel/util v0.0.0-20180919145318-efcd4e0f9787

replace github.com/kubevirt/controller-lifecycle-operator-sdk => ../../
