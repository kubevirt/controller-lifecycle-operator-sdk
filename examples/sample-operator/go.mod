module kubevirt.io/controller-lifecycle-operator-sdk/examples/sample-operator

go 1.13

require (
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/operator-framework/operator-sdk v0.18.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/tools v0.0.0-20201224043029-2b0845dc783e
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.20.2
	kubevirt.io/client-go v0.32.0
	kubevirt.io/controller-lifecycle-operator-sdk v0.0.7
	sigs.k8s.io/controller-runtime v0.8.3
)

replace k8s.io/client-go => k8s.io/client-go v0.20.2

replace k8s.io/api => k8s.io/api v0.20.2

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.2

replace k8s.io/kubectl => k8s.io/kubectl v0.20.2

replace vbom.ml/util => github.com/fvbommel/util v0.0.0-20180919145318-efcd4e0f9787

replace kubevirt.io/controller-lifecycle-operator-sdk => ../../
