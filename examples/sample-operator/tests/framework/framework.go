package framework

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/kubevirt/controller-lifecycle-operator-sdk/examples/sample-operator/pkg/api-client/clientset/versioned"

	"github.com/onsi/ginkgo"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// run-time flags
var (
	kubectlPath              *string
	kubeConfig               *string
	master                   *string
	kubeVirtInstallNamespace *string
)

// Framework supports common operations used by functional/e2e tests.
// This package is based on https://github.com/kubevirt/containerized-data-importer/blob/master/tests/framework/framework.go
type Framework struct {
	// NsPrefix is a prefix for generated namespace
	NsPrefix string
	//  k8sClient provides our k8s client pointer
	K8sClient *kubernetes.Clientset
	//SampleConfigCLient provides SampleConfig client  pointer
	SampleConfigCLient *versioned.Clientset
	// RestConfig provides a pointer to our REST client config.
	RestConfig *rest.Config

	// KubectlPath is a test run-time flag so we can find kubectl
	KubectlPath string
	// KubeConfig is a test run-time flag to store the location of our test setup kubeconfig
	KubeConfig string
	// Master is a test run-time flag to store the id of our master node
	Master string
	// OperatorInstallNamespace namespace where KubeVirt is installed
	OperatorInstallNamespace string
}

// initialize run-time flags
func init() {
	// Make sure that go test flags are registered when the framework is created
	testing.Init()
	kubectlPath = flag.String("kubectl-path", "kubectl", "The path to the kubectl binary")
	kubeConfig = flag.String("kubeconfig", "/var/run/kubernetes/admin.kubeconfig", "The absolute path to the kubeconfig file")
	master = flag.String("master", "", "master url:port")
	kubeVirtInstallNamespace = flag.String("namespace", "kubevirt", "Set the namespace operator is installed in")
}

// NewFrameworkOrDie calls NewFramework and handles errors by calling Fail. Config is optional, but
// if passed there can only be one.
func NewFrameworkOrDie(prefix string) *Framework {
	f, err := NewFramework(prefix)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s", err)
		ginkgo.Fail(fmt.Sprintf("failed to create test framework: %v", err))
	}
	return f
}

// NewFramework makes a new framework and sets up the global BeforeEach/AfterEach's.
// Test run-time flags are parsed and added to the Framework struct.
func NewFramework(prefix string) (*Framework, error) {
	f := &Framework{
		NsPrefix: prefix,
	}

	// handle run-time flags
	if !flag.Parsed() {
		flag.Parse()
	}

	f.KubectlPath = *kubectlPath
	f.KubeConfig = *kubeConfig
	f.Master = *master
	f.OperatorInstallNamespace = *kubeVirtInstallNamespace

	restConfig, err := f.LoadConfig()
	if err != nil {
		// Can't use Expect here due this being called outside of an It block, and Expect
		// requires any calls to it to be inside an It block.
		return nil, errors.Wrap(err, "ERROR, unable to load RestConfig")
	}
	f.RestConfig = restConfig
	// clients
	kcs, err := f.GetKubeClient()
	if err != nil {
		return nil, errors.Wrap(err, "ERROR, unable to create K8SClient")
	}
	f.K8sClient = kcs
	scc, err := f.GetSampleConfigClient()
	if err != nil {
		return nil, errors.Wrap(err, "ERROR, unable to create SampleConfigClient")
	}
	f.SampleConfigCLient = scc

	ginkgo.BeforeEach(f.BeforeEach)
	ginkgo.AfterEach(f.AfterEach)

	return f, err
}

// BeforeEach provides a set of operations to run before each test
func (f *Framework) BeforeEach() {
}

// AfterEach provides a set of operations to run after each test
func (f *Framework) AfterEach() {
	f.CleanUp()
}

// CleanUp provides a set of operations clean the namespace
func (f *Framework) CleanUp() {

}

// GetKubeClient returns a Kubernetes rest client
func (f *Framework) GetKubeClient() (*kubernetes.Clientset, error) {
	return GetKubeClientFromRESTConfig(f.RestConfig)
}

// GetSampleConfigClient gets an instance of a Sample Config client
func (f *Framework) GetSampleConfigClient() (*versioned.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags(f.Master, f.KubeConfig)
	if err != nil {
		return nil, err
	}

	scClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return scClient, nil
}

// LoadConfig loads our specified kubeconfig
func (f *Framework) LoadConfig() (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags(f.Master, f.KubeConfig)
}

// GetKubeClientFromRESTConfig provides a function to get a K8s client using the REST config
func GetKubeClientFromRESTConfig(config *rest.Config) (*kubernetes.Clientset, error) {
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	return kubernetes.NewForConfig(config)
}

// RunKubectlCommand ...
func (f *Framework) RunKubectlCommand(args ...string) (string, error) {
	var errb bytes.Buffer
	cmd := f.createKubectlCommand(args...)

	cmd.Stderr = &errb
	stdOutBytes, err := cmd.Output()
	if err != nil {
		if len(errb.String()) > 0 {
			return errb.String(), err
		}
	}
	return string(stdOutBytes), nil
}

// createKubectlCommand returns the Cmd to execute kubectl
func (f *Framework) createKubectlCommand(args ...string) *exec.Cmd {
	kubeconfig := f.KubeConfig
	path := f.KubectlPath

	cmd := exec.Command(path, args...)
	kubeconfEnv := fmt.Sprintf("KUBECONFIG=%s", kubeconfig)
	cmd.Env = append(os.Environ(), kubeconfEnv)

	return cmd
}
