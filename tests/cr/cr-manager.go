package v1beta1

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/api"
	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/resources"
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
func (m *ConfigCrManager) IsCreating(cr client.Object) (bool, error) {
	config := cr.(*Config)
	return config.Status.Conditions == nil || len(config.Status.Conditions) == 0, nil
}

// Create creates empty CR
func (m *ConfigCrManager) Create() client.Object {
	return new(Config)
}

// Status extracts status from the cr
func (m *ConfigCrManager) Status(cr client.Object) *sdkapi.Status {
	return &cr.(*Config).Status.Status
}

// GetAllResources provides all resources managed by the cr
func (m *ConfigCrManager) GetAllResources(cr client.Object) ([]client.Object, error) {
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

	return []client.Object{operatorDeployment}, nil
}

// GetDependantResourcesListObjects returns resource list objects of dependant resources
func (m *ConfigCrManager) GetDependantResourcesListObjects() []client.ObjectList {
	return []client.ObjectList{
		&appsv1.DeploymentList{},
		&extv1.CustomResourceDefinitionList{}}
}
