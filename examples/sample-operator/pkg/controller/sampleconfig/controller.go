package sampleconfig

import (
	"context"
	"fmt"

	"kubevirt.io/client-go/util"
	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/callbacks"

	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/reconciler"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/apimachinery/pkg/runtime"
	samplev1alpha1 "kubevirt.io/controller-lifecycle-operator-sdk/examples/sample-operator/pkg/apis/sample/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	createVersionLabel          = "operator.sample.lifecycle.kubevirt.io/createVersion"
	updateVersionLabel          = "operator.sample.lifecycle.kubevirt.io/updateVersion"
	lastAppliedConfigAnnotation = "operator.sample.lifecycle.kubevirt.io/lastAppliedConfiguration"
)

// OperatorArgs contains the required parameters to generate all namespaced resources
type OperatorArgs struct {
	OperatorVersion string `required:"true" split_words:"true"`
	ServerImage     string `required:"true" split_words:"true"`
	Namespace       string
}

var log = logf.Log.WithName("controller_sampleconfig")

// Add creates a new SampleConfig Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return r.add(mgr)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (*ReconcileSampleConfig, error) {
	operatorArgs, err := getOperatorArgs()
	if err != nil {
		return nil, err
	}

	scheme := mgr.GetScheme()

	cachingClient := mgr.GetClient()
	uncachedClient, err := client.New(mgr.GetConfig(), client.Options{
		Scheme: scheme,
		Mapper: mgr.GetRESTMapper(),
	})
	if err != nil {
		return nil, err
	}

	callbackDispatcher := callbacks.NewCallbackDispatcher(log, cachingClient, uncachedClient, scheme, operatorArgs.Namespace)
	eventRecorder := mgr.GetEventRecorderFor("sample-config-operator")
	r := reconciler.NewReconciler(&CrManager{operatorArgs: operatorArgs}, log, cachingClient, callbackDispatcher, scheme, mgr.GetCache, createVersionLabel, updateVersionLabel, lastAppliedConfigAnnotation, 0, "sample-finalizer", true, eventRecorder)

	reconcileConfig := &ReconcileSampleConfig{
		client:       cachingClient,
		scheme:       scheme,
		reconciler:   r,
		operatorArgs: operatorArgs,
	}
	return reconcileConfig, nil
}

func getOperatorArgs() (*OperatorArgs, error) {
	operatorArgs := new(OperatorArgs)
	err := envconfig.Process("", operatorArgs)
	if err != nil {
		return nil, err

	}

	namespace, err := util.GetNamespace()
	if err != nil {
		return nil, err
	}
	operatorArgs.Namespace = namespace

	log.Info("", "VARS", fmt.Sprintf("%+v", operatorArgs))
	return operatorArgs, err
}

// SetController sets the controller dependency
func (r *ReconcileSampleConfig) SetController(controller controller.Controller) {
	r.reconciler.WithController(controller)
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func (r *ReconcileSampleConfig) add(mgr manager.Manager) error {
	// Create a new controller
	c, err := controller.New("sampleconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	r.SetController(c)

	// Watch for changes to primary resource SampleConfig
	err = c.Watch(source.Kind(mgr.GetCache(), &samplev1alpha1.SampleConfig{}), &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSampleConfig implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSampleConfig{}

// ReconcileSampleConfig reconciles a SampleConfig object
type ReconcileSampleConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme

	operatorArgs *OperatorArgs

	reconciler *reconciler.Reconciler
}

// Reconcile reads that state of the cluster for a SampleConfig object and makes changes based on the state read
// and what is in the SampleConfig.Spec
func (r *ReconcileSampleConfig) Reconcile(_ context.Context, request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SampleConfig")
	return r.reconciler.Reconcile(request, r.operatorArgs.OperatorVersion, reqLogger)
}
