package framework

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/api"
	"kubevirt.io/controller-lifecycle-operator-sdk/examples/sample-operator/pkg/apis/sample/v1alpha1"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetSampleConfig retrieves SampleConfig instance
func (f *Framework) GetSampleConfig(configName string) (*v1alpha1.SampleConfig, error) {
	return f.SampleConfigCLient.SampleV1alpha1().SampleConfigs().Get(context.TODO(), configName, v1meta.GetOptions{})
}

// GetSampleConfigPhase returns phase of the SampleConfig
func (f *Framework) GetSampleConfigPhase(configName string) (sdkapi.Phase, error) {
	config, err := f.GetSampleConfig(configName)
	if err != nil {
		return "", err
	}
	return config.Status.Phase, err
}

// EnsureSampleConfig creates Sample Config if it does not exist
func (f *Framework) EnsureSampleConfig(configName string) (*v1alpha1.SampleConfig, error) {
	sampleConfig, err := f.GetSampleConfig(configName)
	if err != nil {
		if errors.IsNotFound(err) {
			sampleConfig = &v1alpha1.SampleConfig{}
			sampleConfig.SetName(configName)
			sampleConfig, err = f.SampleConfigCLient.SampleV1alpha1().SampleConfigs().Create(context.TODO(), sampleConfig, v1meta.CreateOptions{})
			return sampleConfig, err
		}
	}
	return sampleConfig, err
}
