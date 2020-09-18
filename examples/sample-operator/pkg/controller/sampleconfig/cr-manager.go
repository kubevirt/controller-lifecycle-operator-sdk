package sampleconfig

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kubevirt.io/controller-lifecycle-operator-sdk/examples/sample-operator/pkg/apis/sample/v1alpha1"
	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/api"
	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/resources"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// HTTPServerDeploymentName defines the name of the HTTP server deployment
	HTTPServerDeploymentName = "http-server"
	// HTTPServerName defines the name of the HTTP server service
	HTTPServerName     = "http-server"
	deploymentMatchKey = "sample-operator.kubevirt.io"
	httpServerPort     = 8081
)

const (
	commonLabel   = "test.common"
	operatorLabel = "test.operator"
)

var (
	commonLabels    = map[string]string{commonLabel: ""}
	operatorLabels  = map[string]string{operatorLabel: ""}
	resourceBuilder = resources.NewResourceBuilder(commonLabels, operatorLabels)
)

// CrManager provides test CR management functionality
type CrManager struct {
	operatorArgs *OperatorArgs
}

// IsCreating checks whether creation of the managed resources will be executed
func (m *CrManager) IsCreating(cr controllerutil.Object) (bool, error) {
	config := cr.(*v1alpha1.SampleConfig)
	return config.Status.Conditions == nil || len(config.Status.Conditions) == 0, nil
}

// Create creates empty CR
func (m *CrManager) Create() controllerutil.Object {
	return new(v1alpha1.SampleConfig)
}

// Status extracts status from the cr
func (m *CrManager) Status(cr runtime.Object) *sdkapi.Status {
	return &cr.(*v1alpha1.SampleConfig).Status.Status
}

// GetAllResources provides all resources managed by the cr
func (m *CrManager) GetAllResources(_ runtime.Object) ([]runtime.Object, error) {
	namespace := m.operatorArgs.Namespace

	serviceAccount := resourceBuilder.CreateServiceAccount(HTTPServerName)
	serviceAccount.Namespace = namespace

	role := resourceBuilder.CreateRole(HTTPServerName, []rbacv1.PolicyRule{})
	role.Namespace = namespace

	roleBinding := resourceBuilder.CreateRoleBinding(HTTPServerName, HTTPServerName, HTTPServerName, namespace)
	roleBinding.Namespace = namespace

	httpDeployment := m.createHTTPServerDeployment()
	httpService := m.createHTTPServerService()

	return []runtime.Object{
		serviceAccount,
		role,
		roleBinding,
		httpDeployment,
		httpService,
	}, nil
}

func (m *CrManager) createHTTPServerService() *v1.Service {
	httpService := resourceBuilder.CreateService(HTTPServerName, deploymentMatchKey, HTTPServerName, nil)
	httpService.Namespace = m.operatorArgs.Namespace
	httpService.Spec.Type = v1.ServiceTypeNodePort
	httpService.Spec.Ports = []v1.ServicePort{
		{Port: httpServerPort, Name: "http", Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: httpServerPort}, NodePort: 30080},
	}
	return httpService
}

func (m *CrManager) createHTTPServerDeployment() *appsv1.Deployment {
	ports := []v1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: httpServerPort,
			Protocol:      v1.ProtocolTCP,
		},
	}
	container := resourceBuilder.CreatePortsContainer("sample-http-server-container", m.operatorArgs.ServerImage, string(v1.PullAlways), ports)
	podSpec := v1.PodSpec{
		Containers: []v1.Container{
			*container,
		},
	}
	return resourceBuilder.CreateOperatorDeployment(HTTPServerDeploymentName, m.operatorArgs.Namespace, deploymentMatchKey, HTTPServerName, HTTPServerName, 1, podSpec)
}

// GetDependantResourcesListObjects returns resource list objects of dependant resources
func (m *CrManager) GetDependantResourcesListObjects() []runtime.Object {
	return []runtime.Object{
		&appsv1.DeploymentList{},
		&v1.ServiceList{},
		&v1.ServiceAccountList{},
		&rbacv1.RoleBindingList{},
		&rbacv1.RoleList{},
	}
}
