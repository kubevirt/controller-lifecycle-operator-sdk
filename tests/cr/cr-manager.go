package v1beta1

import (
	sdkapi "github.com/kubevirt/controller-lifecycle-operator-sdk/pkg/sdk/api"
	"github.com/kubevirt/controller-lifecycle-operator-sdk/pkg/sdk/resources"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	Namespace              = "operator-namespace"
	OperatorDeploymentName = "operator-deployment"
)

const (
	CommonLabel   = "test.common"
	OperatorLabel = "test.operator"
)

var (
	commonLabels    = map[string]string{CommonLabel: ""}
	operatorLabels  = map[string]string{OperatorLabel: ""}
	ResourceBuilder = resources.NewResourceBuilder(commonLabels, operatorLabels)
)

// ConfigCrManager provides test CR management functionality
type ConfigCrManager struct {
}

// IsCreating checks whether creation of the managed resources will be executed
func (m *ConfigCrManager) IsCreating(cr controllerutil.Object) (bool, error) {
	config := cr.(*Config)
	return config.Status.Conditions == nil || len(config.Status.Conditions) == 0, nil
}

// Create creates empty CR
func (m *ConfigCrManager) Create() controllerutil.Object {
	return new(Config)
}

// Status extracts status from the cr
func (m *ConfigCrManager) Status(cr runtime.Object) *sdkapi.Status {
	return &cr.(*Config).Status.Status
}

// GetAllResources provides all resources managed by the cr
func (m *ConfigCrManager) GetAllResources(cr runtime.Object) ([]runtime.Object, error) {
	container := ResourceBuilder.CreateContainer("a-container", "image", string(v1.PullIfNotPresent))
	container.Env = []v1.EnvVar{
		{Name: "Foo", Value: "BAR"},
	}
	podSpec := v1.PodSpec{
		Containers: []v1.Container{
			*container,
		},
	}
	operatorDeployment := ResourceBuilder.CreateOperatorDeployment(OperatorDeploymentName, Namespace, "key", "value", "svc-account", 1, podSpec)

	return []runtime.Object{operatorDeployment}, nil
}

// GetDependantResourcesListObjects returns resource list objects of dependant resources
func (m *ConfigCrManager) GetDependantResourcesListObjects() []runtime.Object {
	return []runtime.Object{
		&appsv1.DeploymentList{},
		&extv1.CustomResourceDefinitionList{}}
}
