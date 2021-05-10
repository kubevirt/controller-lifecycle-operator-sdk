package reconciler_test

import (
	"context"
	"fmt"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/go-logr/logr"
	v1 "github.com/openshift/custom-resource-status/conditions/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/api"
	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/callbacks"
	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/reconciler"
	testcr "kubevirt.io/controller-lifecycle-operator-sdk/tests/cr"
	"kubevirt.io/controller-lifecycle-operator-sdk/tests/mocks"
)

type modifyResource func(toModify client.Object) (client.Object, client.Object, error)
type isModifySubject func(resource client.Object) bool
type isUpgraded func(postUpgradeObj client.Object, deisredObj client.Object) bool
type createUnusedObject func() (client.Object, error)

type mockCallbackDispatcher struct {
}

type args struct {
	config         *testcr.Config
	client         client.Client
	reconciler     *reconciler.Reconciler
	version        string
	mockController *mocks.MockController
	recorder       *record.FakeRecorder
}

const (
	finalizerName      = "my-finalizer"
	version            = "v1.5.0"
	createVersionLabel = "create-version"

	normalCreateSuccess = "Normal CreateResourceSuccess Successfully created resource"
)

var log = logf.Log.WithName("tests")

var callbackDispatcher = &mockCallbackDispatcher{}
var (
	invokeCallbacks func(interface{}, callbacks.ReconcileState, client.Object, client.Object) error
	addCallback     func(client.Object, callbacks.ReconcileCallback)
)

var _ = Describe("Reconciler", func() {

	BeforeEach(func() {
		invokeCallbacks = func(interface{}, callbacks.ReconcileState, client.Object, client.Object) error {
			return nil
		}
		addCallback = func(client.Object, callbacks.ReconcileCallback) {}
	})

	Describe("exported methods", func() {
		It("should init the CR", func() {
			args := createArgs(version)

			err := args.reconciler.CrInit(args.config, version)

			Expect(err).ToNot(HaveOccurred())

			Expect(args.config.Status.Phase).To(BeEquivalentTo(sdkapi.PhaseDeploying))
			Expect(args.config.Status.TargetVersion).To(BeEquivalentTo(version))
			Expect(args.config.Status.OperatorVersion).To(BeEquivalentTo(version))

			Expect(args.config.GetFinalizers()).To(HaveLen(1))
			Expect(args.config.GetFinalizers()[0]).To(BeEquivalentTo(finalizerName))
		})

		It("should set CR to error state", func() {
			args := createArgs(version)
			err := args.reconciler.CrError(args.config)

			Expect(err).ToNot(HaveOccurred())
			Expect(args.config.Status.Phase).To(BeEquivalentTo(sdkapi.PhaseError))
		})

		It("should set CR version", func() {
			args := createArgs(version)
			newVersion := "v0.1.0"

			err := args.reconciler.CrSetVersion(args.config, newVersion)
			Expect(err).ToNot(HaveOccurred())

			Expect(args.config.Status.Phase).To(Equal(sdkapi.PhaseDeployed))
			Expect(args.config.Status.ObservedVersion).To(Equal(newVersion))
			Expect(args.config.Status.OperatorVersion).To(Equal(newVersion))
			Expect(args.config.Status.TargetVersion).To(Equal(newVersion))
		})

		It("should register CR watching in cantroller", func() {
			args := createArgs(version)

			err := args.reconciler.WatchCR()

			Expect(err).ToNot(HaveOccurred())

			Expect(args.mockController.WatchCalls).To(HaveLen(1))
			src := args.mockController.WatchCalls[0].Src
			kind, ok := src.(*source.Kind)
			Expect(ok).To(BeTrue())
			Expect(kind.Type).To(BeAssignableToTypeOf(&testcr.Config{}))
		})
	})

	Describe("deploying operator", func() {
		Context("Operator lifecycle", func() {
			It("should deploy", func() {
				args := createArgs(version)
				doReconcile(args)
				setDeploymentsReady(args)

				Expect(args.config.Status.OperatorVersion).Should(Equal(version))
				Expect(args.config.Status.TargetVersion).Should(Equal(version))
				Expect(args.config.Status.ObservedVersion).Should(Equal(version))

				Expect(args.config.Status.Conditions).Should(HaveLen(3))
				Expect(v1.IsStatusConditionTrue(args.config.Status.Conditions, v1.ConditionAvailable)).To(BeTrue())
				Expect(v1.IsStatusConditionFalse(args.config.Status.Conditions, v1.ConditionProgressing)).To(BeTrue())
				Expect(v1.IsStatusConditionFalse(args.config.Status.Conditions, v1.ConditionDegraded)).To(BeTrue())

				Expect(args.config.Finalizers).Should(HaveLen(1))
				validateEvents(args.recorder, createReadyEventValidationMap())
			})

			It("should become ready", func() {
				args := createArgs(version)
				doReconcile(args)

				Expect(setDeploymentsReady(args)).To(BeTrue())
				validateEvents(args.recorder, createReadyEventValidationMap())
			})

			It("should create all resources", func() {
				args := createArgs(version)
				doReconcile(args)

				resources := getAllResources(args.config)

				for _, r := range resources {
					_, err := getObject(args.client, r)
					Expect(err).ToNot(HaveOccurred())
				}
				validateEvents(args.recorder, createNotReadyEventValidationMap())
			})

			It("should delete", func() {
				args := createArgs(version)
				doReconcile(args)

				args.config.DeletionTimestamp = &metav1.Time{Time: time.Now()}
				err := args.client.Update(context.TODO(), args.config)
				Expect(err).ToNot(HaveOccurred())

				doReconcile(args)

				Expect(args.config.Finalizers).Should(BeEmpty())
				Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeleted))
			})
		})
	})

	Describe("Upgrading operator", func() {
		DescribeTable("should upgrade", func(prevVersion, newVersion string) {
			args := createArgs(prevVersion)
			doReconcile(args)

			setDeploymentsReady(args)

			Expect(args.config.Status.ObservedVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.OperatorVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.TargetVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))

			setDeploymentsDegraded(args)
			args.version = newVersion
			doReconcile(args)

			//verify upgraded has started
			Expect(args.config.Status.OperatorVersion).Should(Equal(newVersion))
			Expect(args.config.Status.ObservedVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.TargetVersion).Should(Equal(newVersion))
			Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseUpgrading))

			//change deployment to ready
			isReady := setDeploymentsReady(args)
			Expect(isReady).Should(Equal(true))

			//verify versions were updated
			Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))
			Expect(args.config.Status.OperatorVersion).Should(Equal(newVersion))
			Expect(args.config.Status.TargetVersion).Should(Equal(newVersion))
			Expect(args.config.Status.ObservedVersion).Should(Equal(newVersion))
		},
			Entry("increasing semver ", "v1.9.5", "v1.10.0"),
			Entry("invalid semver", "devel", "v1.9.5"),
			Entry("increasing  semver no prefix", "1.9.5", "1.10.0"),
			Entry("invalid  semver no prefix", "devel", "1.9.5"),
			Entry("no previous version", "", "1.9.5"),
		)

		DescribeTable("should not upgrade", func(prevVersion, newVersion string) {
			args := createArgs(prevVersion)
			doReconcile(args)

			setDeploymentsReady(args)

			Expect(args.config.Status.ObservedVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.OperatorVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.TargetVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))

			setDeploymentsDegraded(args)
			args.version = newVersion
			doReconcile(args)

			// verify upgraded hasn't started
			Expect(args.config.Status.OperatorVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.ObservedVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.TargetVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))

			//change deployment to ready
			isReady := setDeploymentsReady(args)
			Expect(isReady).Should(Equal(true))

			//verify versions remained unchaged
			Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))
			Expect(args.config.Status.OperatorVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.TargetVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.ObservedVersion).Should(Equal(prevVersion))
		},
			Entry("identical semver", "v1.10.0", "v1.10.0"),
			Entry("identical  semver no prefix", "1.10.0", "1.10.0"),
			Entry("invalid  semver with prefix", "devel1.9.5", "devel1.9.5"),
		)

		DescribeTable("should fail on downgrade", func(prevVersion, newVersion string) {
			args := createArgs(prevVersion)
			doReconcile(args)

			setDeploymentsReady(args)

			Expect(args.config.Status.ObservedVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.OperatorVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.TargetVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))

			setDeploymentsDegraded(args)
			args.version = newVersion

			doReconcileError(args)
		},
			Entry("decreasing semver", "v1.10.0", "v1.9.5"),
			Entry("decreasing  semver no prefix", "1.10.0", "1.9.5"),
		)

	})

	DescribeTable("Restores objects on upgrade", func(modify modifyResource, tomodify isModifySubject, upgraded isUpgraded) {
		newVersion := "v0.0.2"
		prevVersion := "v0.0.1"

		args := createArgs(newVersion)
		doReconcile(args)
		setDeploymentsReady(args)

		//verify on int version is set
		Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))

		//Modify CRD to be of previousVersion
		_ = args.reconciler.CrSetVersion(args.config, prevVersion)
		err := args.client.Update(context.TODO(), args.config)
		Expect(err).ToNot(HaveOccurred())

		setDeploymentsDegraded(args)

		//find the resource to modify
		oOriginal, oModified, err := getModifiedResource(args, modify, tomodify)
		Expect(err).ToNot(HaveOccurred())

		//update object via client, with curObject
		err = args.client.Update(context.TODO(), oModified)
		Expect(err).ToNot(HaveOccurred())

		//verify object is modified
		storedObj, err := getObject(args.client, oModified)
		Expect(err).ToNot(HaveOccurred())

		Expect(reflect.DeepEqual(storedObj, oModified)).Should(Equal(true))

		doReconcile(args)

		//verify upgraded has started
		Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseUpgrading))

		//change deployment to ready
		isReady := setDeploymentsReady(args)
		Expect(isReady).Should(Equal(true))

		doReconcile(args)
		Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))

		//verify that stored object equals to object in getResources
		storedObj, err = getObject(args.client, oModified)
		Expect(err).ToNot(HaveOccurred())

		Expect(upgraded(storedObj, oOriginal)).Should(Equal(true))
	},
		Entry("verify - deployment updated on upgrade - deployment spec changed - modify container",
			func(toModify client.Object) (client.Object, client.Object, error) { //Modify
				deploymentOrig, ok := toModify.(*appsv1.Deployment)
				if !ok {
					return toModify, toModify, fmt.Errorf("wrong type")
				}
				deployment := deploymentOrig.DeepCopy()

				containers := deployment.Spec.Template.Spec.Containers
				containers[0].Env = []corev1.EnvVar{
					{
						Name:  "FAKE_ENVVAR_1",
						Value: "env var value 1",
					},
					{
						Name:  "FAKE_ENVVAR_2",
						Value: "env var value 2",
					},
				}

				return toModify, deployment, nil
			},
			func(resource client.Object) bool { //find resource for test
				deployment, ok := resource.(*appsv1.Deployment)
				if !ok {
					return false
				}
				if deployment.Name == testcr.OperatorDeploymentName {
					return true
				}
				return false
			},
			func(postUpgradeObj client.Object, deisredObj client.Object) bool { //check resource was upgraded
				postDep, ok := postUpgradeObj.(*appsv1.Deployment)
				if !ok {
					return false
				}

				desiredDep, ok := deisredObj.(*appsv1.Deployment)
				if !ok {
					return false
				}

				for key, envVar := range desiredDep.Spec.Template.Spec.Containers[0].Env {
					if postDep.Spec.Template.Spec.Containers[0].Env[key].Name != envVar.Name {
						return false
					}
				}

				return len(desiredDep.Spec.Template.Spec.Containers[0].Env) == len(postDep.Spec.Template.Spec.Containers[0].Env)
			}),
		Entry("verify - deployment updated on upgrade - deployment spec changed - add new container",
			func(toModify client.Object) (client.Object, client.Object, error) { //Modify
				deploymentOrig, ok := toModify.(*appsv1.Deployment)
				if !ok {
					return toModify, toModify, fmt.Errorf("wrong type")
				}
				deployment := deploymentOrig.DeepCopy()

				containers := deployment.Spec.Template.Spec.Containers
				container := corev1.Container{
					Name:            "FAKE_CONTAINER",
					Image:           fmt.Sprintf("%s/%s:%s", "fake-repo", "fake-image", "fake-tag"),
					ImagePullPolicy: "FakePullPolicy",
					Args:            []string{"-v=10"},
				}
				containers = append(containers, container)
				deployment.Spec.Template.Spec.Containers = containers
				return toModify, deployment, nil
			},
			func(resource client.Object) bool { //find resource for test
				deployment, ok := resource.(*appsv1.Deployment)
				if !ok {
					return false
				}
				return deployment.Name == testcr.OperatorDeploymentName
			},
			func(postUpgradeObj client.Object, deisredObj client.Object) bool { //check resource was upgraded
				postDep, ok := postUpgradeObj.(*appsv1.Deployment)
				if !ok {
					return false
				}

				desiredDep, ok := deisredObj.(*appsv1.Deployment)
				if !ok {
					return false
				}

				for key, container := range desiredDep.Spec.Template.Spec.Containers {
					if postDep.Spec.Template.Spec.Containers[key].Name != container.Name {
						return false
					}
				}

				return len(desiredDep.Spec.Template.Spec.Containers) == len(postDep.Spec.Template.Spec.Containers)
			}),
		Entry("verify - deployment updated on upgrade - deployment spec changed - remove existing container",
			func(toModify client.Object) (client.Object, client.Object, error) { //Modify
				deploymentOrig, ok := toModify.(*appsv1.Deployment)
				if !ok {
					return toModify, toModify, fmt.Errorf("wrong type")
				}
				deployment := deploymentOrig.DeepCopy()

				deployment.Spec.Template.Spec.Containers = nil

				return toModify, deployment, nil
			},
			func(resource client.Object) bool { //find resource for test
				deployment, ok := resource.(*appsv1.Deployment)
				if !ok {
					return false
				}
				if deployment.Name == testcr.OperatorDeploymentName {
					return true
				}
				return false
			},
			func(postUpgradeObj client.Object, deisredObj client.Object) bool { //check resource was upgraded
				postDep, ok := postUpgradeObj.(*appsv1.Deployment)
				if !ok {
					return false
				}

				desiredDep, ok := deisredObj.(*appsv1.Deployment)
				if !ok {
					return false
				}

				return len(postDep.Spec.Template.Spec.Containers) == len(desiredDep.Spec.Template.Spec.Containers)
			}),
	)
	DescribeTable("Removes unused objects on upgrade", func(createObj createUnusedObject) {
		newVersion := "v0.0.2"
		prevVersion := "v0.0.1"

		args := createArgs(newVersion)
		doReconcile(args)

		setDeploymentsReady(args)

		//verify on int version is set
		Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))

		//Modify CRD to be of previousVersion
		_ = args.reconciler.CrSetVersion(args.config, prevVersion)
		err := args.client.Update(context.TODO(), args.config)
		Expect(err).ToNot(HaveOccurred())

		setDeploymentsDegraded(args)
		unusedObj, err := createObj()
		Expect(err).ToNot(HaveOccurred())
		unusedMetaObj := unusedObj.(metav1.Object)
		unusedMetaObj.SetLabels(make(map[string]string))
		unusedMetaObj.GetLabels()[createVersionLabel] = prevVersion
		err = controllerutil.SetControllerReference(args.config, unusedMetaObj, scheme.Scheme)
		Expect(err).ToNot(HaveOccurred())

		//add unused object via client, with curObject
		err = args.client.Create(context.TODO(), unusedObj)
		Expect(err).ToNot(HaveOccurred())

		doReconcile(args)

		//verify upgraded has started
		Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseUpgrading))

		//verify unused exists before upgrade is done
		_, err = getObject(args.client, unusedObj)
		Expect(err).ToNot(HaveOccurred())

		//change deployment to ready
		isReady := setDeploymentsReady(args)
		Expect(isReady).Should(Equal(true))

		doReconcile(args)
		Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))

		//verify that object no longer exists after upgrade
		_, err = getObject(args.client, unusedObj)
		Expect(errors.IsNotFound(err)).Should(Equal(true))
	},
		Entry("verify - unused deployment deleted",
			func() (client.Object, error) {
				deployment := testcr.ResourceBuilder.CreateDeployment("fake-deployment", testcr.Namespace, "match-key", "match-value", "", int32(1), corev1.PodSpec{}, nil)
				return deployment, nil
			}),

		Entry("verify - unused crd deleted",
			func() (client.Object, error) {
				crd := &extv1.CustomResourceDefinition{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apiextensions.k8s.io/v1",
						Kind:       "CustomResourceDefinition",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "fake.configs.test",
						Labels: map[string]string{
							testcr.OperatorLabel: "",
						},
					},
					Spec: extv1.CustomResourceDefinitionSpec{
						Group: "test",
						Scope: "Cluster",

						Versions: []extv1.CustomResourceDefinitionVersion{
							{
								Name:    "v1beta1",
								Served:  true,
								Storage: true,
								AdditionalPrinterColumns: []extv1.CustomResourceColumnDefinition{
									{Name: "Age", Type: "date", JSONPath: ".metadata.creationTimestamp"},
									{Name: "Phase", Type: "string", JSONPath: ".status.phase"},
								},
							},
						},
						Names: extv1.CustomResourceDefinitionNames{
							Kind:     "FakeConfig",
							ListKind: "ConfigList",
							Plural:   "fakeconfigs",
							Singular: "fakeconfig",
							Categories: []string{
								"all",
							},
							ShortNames: []string{"fakeconfig", "fakeconfigs"},
						},
					},
				}
				return crd, nil
			}),
	)

	Describe("Config CR deletion during upgrade", func() {
		It("should delete CR if it is marked for deletion and not begin upgrade flow", func() {
			newVersion := "v0.0.2"
			prevVersion := "v0.0.1"

			args := createArgs(newVersion)
			doReconcile(args)

			//set deployment to ready
			isReady := setDeploymentsReady(args)
			Expect(isReady).Should(Equal(true))

			//verify on int version is set
			Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))

			//Modify CRD to be of previousVersion
			_ = args.reconciler.CrSetVersion(args.config, prevVersion)
			//mark CR for deletion
			args.config.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})
			err := args.client.Update(context.TODO(), args.config)
			Expect(err).ToNot(HaveOccurred())

			doReconcile(args)

			//verify the version cr is deleted and upgrade hasn't started
			Expect(args.config.Status.OperatorVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.ObservedVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.TargetVersion).Should(Equal(prevVersion))
			Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeleted))
		})

		It("should delete CR if it is marked for deletion during upgrade flow", func() {
			newVersion := "v0.0.2"
			prevVersion := "v0.0.1"

			args := createArgs(newVersion)
			doReconcile(args)
			setDeploymentsReady(args)

			//verify on int version is set
			Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeployed))

			//Modify CRD to be of previousVersion
			_ = args.reconciler.CrSetVersion(args.config, prevVersion)
			err := args.client.Update(context.TODO(), args.config)
			Expect(err).ToNot(HaveOccurred())
			setDeploymentsDegraded(args)

			//begin upgrade
			doReconcile(args)

			//mark CR for deletion
			args.config.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})
			err = args.client.Update(context.TODO(), args.config)
			Expect(err).ToNot(HaveOccurred())

			doReconcile(args)

			//set deployment to ready
			isReady := setDeploymentsReady(args)
			Expect(isReady).Should(Equal(false))

			doReconcile(args)
			//verify the version cr is marked as deleted
			Expect(args.config.Status.Phase).Should(Equal(sdkapi.PhaseDeleted))
		})
	})
})

func getConfig(c client.Client, cr *testcr.Config) (*testcr.Config, error) {
	result, err := getObject(c, cr)
	if err != nil {
		return nil, err
	}
	return result.(*testcr.Config), nil
}

func createClient(scheme *runtime.Scheme, objs ...runtime.Object) client.Client {
	return fakeClient.NewFakeClientWithScheme(scheme, objs...)
}

func createReconciler(client client.Client, s *runtime.Scheme, recorder record.EventRecorder) *reconciler.Reconciler {
	crManager := &testcr.ConfigCrManager{}
	return reconciler.NewReconciler(crManager, log, client, callbackDispatcher, s, createVersionLabel, "update-version", "last-applied-config", 0, finalizerName, recorder)
}

func createConfig(name, uid string) *testcr.Config {
	return &testcr.Config{ObjectMeta: metav1.ObjectMeta{Name: name, UID: types.UID(uid)}, Status: testcr.ConfigStatus{}}
}

func (m *mockCallbackDispatcher) InvokeCallbacks(_ logr.Logger, cr interface{}, s callbacks.ReconcileState, desiredObj, currentObj client.Object, recorder record.EventRecorder) error {
	return invokeCallbacks(cr, s, desiredObj, currentObj)
}

func (m *mockCallbackDispatcher) AddCallback(obj client.Object, cb callbacks.ReconcileCallback) {
	addCallback(obj, cb)
}

func reconcileRequest(name string) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{Name: name}}
}

func createArgs(version string) *args {
	config := createConfig("test", "unique-id")
	s := scheme.Scheme
	err := testcr.AddToScheme(s)
	if err != nil {
		Fail(err.Error())
	}
	err = extv1.AddToScheme(s)
	if err != nil {
		Fail(err.Error())
	}
	client := createClient(s, config)
	mockController := mocks.MockController{}
	recorder := record.NewFakeRecorder(250)
	r := createReconciler(client, s, recorder)
	r.WithController(&mockController)

	return &args{
		config:         config,
		client:         client,
		reconciler:     r,
		version:        version,
		mockController: &mockController,
		recorder:       recorder,
	}
}

func doReconcile(args *args) {
	result, err := args.reconciler.Reconcile(reconcileRequest(args.config.Name), args.version, log)
	Expect(err).ToNot(HaveOccurred())
	Expect(result.Requeue).To(BeFalse())

	args.config, err = getConfig(args.client, args.config)
	Expect(err).ToNot(HaveOccurred())
}

func doReconcileError(args *args) {
	result, err := args.reconciler.Reconcile(reconcileRequest(args.config.Name), args.version, log)
	Expect(err).To(HaveOccurred())
	Expect(result.Requeue).To(BeFalse())

	args.config, err = getConfig(args.client, args.config)
	Expect(err).ToNot(HaveOccurred())
}

func setDeploymentsReady(args *args) bool {
	crManager := testcr.ConfigCrManager{}
	resources, err := crManager.GetAllResources(args.config)
	Expect(err).ToNot(HaveOccurred())
	running := false

	for _, r := range resources {
		d, ok := r.(*appsv1.Deployment)
		if !ok {
			continue
		}

		Expect(running).To(BeFalse())

		d, err := getDeployment(args.client, d)
		Expect(err).ToNot(HaveOccurred())
		if d.Spec.Replicas != nil {
			d.Status.Replicas = *d.Spec.Replicas
			d.Status.ReadyReplicas = d.Status.Replicas
			err = args.client.Update(context.TODO(), d)
			Expect(err).ToNot(HaveOccurred())
		}

		doReconcile(args)

		if len(args.config.Status.Conditions) == 3 &&
			v1.IsStatusConditionTrue(args.config.Status.Conditions, v1.ConditionAvailable) &&
			v1.IsStatusConditionFalse(args.config.Status.Conditions, v1.ConditionProgressing) &&
			v1.IsStatusConditionFalse(args.config.Status.Conditions, v1.ConditionDegraded) {
			running = true
		}
	}

	return running
}

func setDeploymentsDegraded(args *args) {
	resources := getAllResources(args.config)

	for _, r := range resources {
		d, ok := r.(*appsv1.Deployment)
		if !ok {
			continue
		}

		d, err := getDeployment(args.client, d)
		Expect(err).ToNot(HaveOccurred())
		if d.Spec.Replicas != nil {
			d.Status.Replicas = int32(0)
			d.Status.ReadyReplicas = d.Status.Replicas
			err = args.client.Update(context.TODO(), d)
			Expect(err).ToNot(HaveOccurred())
		}

	}
	doReconcile(args)
}

func getAllResources(cr client.Object) []client.Object {
	crManager := testcr.ConfigCrManager{}
	resources, err := crManager.GetAllResources(cr)
	Expect(err).ToNot(HaveOccurred())
	return resources
}

func getDeployment(client client.Client, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	result, err := getObject(client, deployment)
	if err != nil {
		return nil, err
	}
	return result.(*appsv1.Deployment), nil
}

func getObject(c client.Client, obj client.Object) (client.Object, error) {
	metaObj := obj.(metav1.Object)
	key := client.ObjectKey{Namespace: metaObj.GetNamespace(), Name: metaObj.GetName()}

	typ := reflect.ValueOf(obj).Elem().Type()
	result := reflect.New(typ).Interface().(client.Object)

	if err := c.Get(context.TODO(), key, result); err != nil {
		return nil, err
	}

	return result, nil
}

func getModifiedResource(args *args, modify modifyResource, tomodify isModifySubject) (client.Object, client.Object, error) {
	resources := getAllResources(args.config)

	//find the resource to modify
	var orig client.Object
	for _, resource := range resources {
		r, err := getObject(args.client, resource)
		Expect(err).ToNot(HaveOccurred())
		if tomodify(r) {
			orig = r
			break
		}
	}
	//apply modify function on resource and return modified one
	return modify(orig)
}

func validateEvents(recorder *record.FakeRecorder, match map[string]bool) {
	events := recorder.Events
	// Closing the channel allows me to do non blocking reads of the channel, once the channel runs out of items the loop exits.
	close(events)
	for event := range events {
		val, ok := match[event]
		Expect(ok).To(BeTrue(), "Event [%s] was not expected", event)
		if !val {
			match[event] = true
		}
	}
	for k, v := range match {
		Expect(v).To(BeTrue(), "Event [%s] not observed", k)
	}
}

func createReadyEventValidationMap() map[string]bool {
	match := createNotReadyEventValidationMap()
	match["Normal DeployCompleted Deployment Completed"] = false
	return match
}

func createNotReadyEventValidationMap() map[string]bool {
	// match is map of strings and if we observed the event.
	// We are not interested in the order of the events, just that the events happen at least once.
	match := make(map[string]bool)
	match["Normal DeployStarted Started Deployment"] = false
	match[normalCreateSuccess+" *v1.Deployment "+testcr.OperatorDeploymentName] = false
	return match
}
