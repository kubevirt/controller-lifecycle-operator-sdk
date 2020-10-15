package ovirt_test

import (
	"context"
	"time"

	"kubevirt.io/controller-lifecycle-operator-sdk/examples/sample-operator/pkg/controller/sampleconfig"

	"k8s.io/apimachinery/pkg/api/errors"

	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	fwk "kubevirt.io/controller-lifecycle-operator-sdk/examples/sample-operator/tests/framework"
)

const sampleConfigName = "example-sampleconfig"

var _ = Describe("Operator lifecycle test ", func() {
	var (
		f           = fwk.NewFrameworkOrDie("operator-lifecycle")
		deployments = f.K8sClient.AppsV1().Deployments(f.OperatorInstallNamespace)

		originalOperatorImage   string
		originalOperatorVersion string
		originalServerImage     string
	)

	BeforeEach(func() {
		operatorDeployment, err := deployments.Get(context.TODO(), "sample-operator", v1meta.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		originalOperatorImage = operatorDeployment.Spec.Template.Spec.Containers[0].Image
		env := operatorDeployment.Spec.Template.Spec.Containers[0].Env
		for i := range env {
			switch env[i].Name {
			case "OPERATOR_VERSION":
				originalOperatorVersion = env[i].Value
			case "SERVER_IMAGE":
				originalServerImage = env[i].Value
			}
		}

		_, err = f.EnsureSampleConfig(sampleConfigName)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		operatorDeployment, err := deployments.Get(context.TODO(), "sample-operator", v1meta.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		operatorDeployment.Spec.Template.Spec.Containers[0].Image = originalOperatorImage
		env := operatorDeployment.Spec.Template.Spec.Containers[0].Env
		for i := range env {
			switch env[i].Name {
			case "OPERATOR_VERSION":
				env[i].Value = originalOperatorVersion
			case "SERVER_IMAGE":
				env[i].Value = originalServerImage
			}
		}
		_, err = deployments.Update(context.TODO(), operatorDeployment, v1meta.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())

		_, err = f.EnsureSampleConfig(sampleConfigName)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() (sdkapi.Status, error) {
			sampleConfig, err := f.GetSampleConfig(sampleConfigName)
			if err != nil {
				return sdkapi.Status{}, err
			}
			return sampleConfig.Status.Status, err
		}, 2*time.Minute, time.Second).Should(And(
			WithTransform(func(status sdkapi.Status) sdkapi.Phase {
				return status.Phase
			}, Equal(sdkapi.PhaseDeployed)),
			WithTransform(func(status sdkapi.Status) string {
				return status.OperatorVersion
			}, Equal(originalOperatorVersion)),
		))

	})

	It("should upgrade the http server", func() {
		By("confirming it is actually deployed")
		assertSampleConfigDeployedAndReportingVersion(f, "latest")

		deployments := f.K8sClient.AppsV1().Deployments(f.OperatorInstallNamespace)
		operatorDeployment, err := deployments.Get(context.TODO(), "sample-operator", v1meta.GetOptions{})
		Expect(err).ToNot(HaveOccurred())

		Expect(operatorDeployment.Spec.Template.Spec.Containers).To(HaveLen(1))

		By("checking pre-existing version")
		expectedUpgradedVersion := "v0.0.4"
		Expect(operatorDeployment.Spec.Template.Spec.Containers[0].Image).ToNot(ContainSubstring(expectedUpgradedVersion))
		operatorDeployment.Spec.Template.Spec.Containers[0].Image = "quay.io/jdzon/sample-operator:v0.0.4"
		env := operatorDeployment.Spec.Template.Spec.Containers[0].Env
		for i := range env {
			switch env[i].Name {
			case "OPERATOR_VERSION":
				Expect(env[i].Value).ToNot(Equal(expectedUpgradedVersion))
				env[i].Value = expectedUpgradedVersion
			case "SERVER_IMAGE":
				Expect(env[i].Value).ToNot(ContainSubstring(expectedUpgradedVersion))
				env[i].Value = "quay.io/jdzon/sample-http-server:v0.0.4"
			}
		}
		By("updating versions on the deployment")
		operatorDeployment.Spec.Template.Spec.Containers[0].Env = env
		_, err = deployments.Update(context.TODO(), operatorDeployment, v1meta.UpdateOptions{})
		Expect(err).ToNot(HaveOccurred())

		By("sample config going into Deployed phase and reporting expected versions ")
		assertSampleConfigDeployedAndReportingVersion(f, expectedUpgradedVersion)

		By("having the HTTP server deployment one available replica")
		assertHTTPServerAvailable(f)
	})

	It("should remove the http server", func() {
		By("confirming it is actually deployed")
		assertSampleConfigDeployedAndReportingVersion(f, "latest")
		assertHTTPServerAvailable(f)

		By("removing the SampleConfig")
		err := f.SampleConfigCLient.SampleV1alpha1().SampleConfigs().Delete(context.TODO(), sampleConfigName, v1meta.DeleteOptions{})
		Expect(err).ToNot(HaveOccurred())

		By("Confirming that the HTTP server deployment is gone")
		Eventually(func() error {
			_, err := deployments.Get(context.TODO(), sampleconfig.HTTPServerDeploymentName, v1meta.GetOptions{})
			return err
		}, 2*time.Minute, time.Second).Should(
			And(
				HaveOccurred(),
				WithTransform(func(err error) bool {
					return errors.IsNotFound(err)
				}, BeTrue()),
			),
		)
	})
})

func assertHTTPServerAvailable(f *fwk.Framework) bool {
	return Eventually(func() (int32, error) {
		httpDeployment, err := f.K8sClient.AppsV1().Deployments(f.OperatorInstallNamespace).Get(context.TODO(), sampleconfig.HTTPServerDeploymentName, v1meta.GetOptions{})
		if err != nil {
			return 0, err
		}
		return httpDeployment.Status.AvailableReplicas, nil
	}, 2*time.Minute, time.Second).Should(BeEquivalentTo(1))
}

func assertSampleConfigDeployedAndReportingVersion(f *fwk.Framework, expectedUpgradedVersion string) bool {
	return Eventually(func() (sdkapi.Status, error) {
		sampleConfig, err := f.GetSampleConfig(sampleConfigName)
		if err != nil {
			return sdkapi.Status{}, err
		}
		return sampleConfig.Status.Status, err
	}, 2*time.Minute, time.Second).Should(And(
		WithTransform(func(status sdkapi.Status) sdkapi.Phase {
			return status.Phase
		}, Equal(sdkapi.PhaseDeployed)),
		WithTransform(func(status sdkapi.Status) string {
			return status.OperatorVersion
		}, Equal(expectedUpgradedVersion)),
		WithTransform(func(status sdkapi.Status) string {
			return status.ObservedVersion
		}, Equal(expectedUpgradedVersion)),
		WithTransform(func(status sdkapi.Status) string {
			return status.TargetVersion
		}, Equal(expectedUpgradedVersion)),
	))
}
