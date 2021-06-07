package callbacks_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/callbacks"
	testcr "kubevirt.io/controller-lifecycle-operator-sdk/tests/cr"
)

const (
	namespace = "namespace"
)

var _ = Describe("Callback dispatcher ", func() {
	recorder := &record.FakeRecorder{}
	log := logf.Log.WithName("tests")
	s := scheme.Scheme
	existingService := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc-1",
			Namespace: "svc-ns",
		},
	}
	client := fakeClient.NewFakeClientWithScheme(s, &existingService)
	cd := callbacks.NewCallbackDispatcher(log, client, client, s, namespace)

	It("should register and invoke callback", func() {
		desiredObj := v1.Pod{}
		currentObj := v1.Pod{}
		cr := testcr.Config{}
		reconcileState := callbacks.ReconcileStatePreCreate

		callbackArguments := new([]*callbacks.ReconcileCallbackArgs)
		callback := func(args *callbacks.ReconcileCallbackArgs) error {
			tmp := append(*callbackArguments, args)
			callbackArguments = &tmp
			return nil
		}

		By("registering callback")
		cd.AddCallback(&desiredObj, callback)

		By("invoking callback")

		err := cd.InvokeCallbacks(log, cr, reconcileState, &desiredObj, &currentObj, recorder)

		Expect(err).ToNot(HaveOccurred())

		Expect(*callbackArguments).To(HaveLen(1))
		args := (*callbackArguments)[0]

		Expect(args.Logger).To(Equal(log))
		Expect(args.Scheme).To(Equal(s))
		Expect(args.Namespace).To(Equal(namespace))
		Expect(args.State).To(Equal(reconcileState))
		Expect(args.Client).To(Equal(client))
		Expect(args.CurrentObject).To(Equal(&currentObj))
		Expect(args.DesiredObject).To(Equal(&desiredObj))
		Expect(args.Resource).To(Equal(cr))
	})

	It("should register and invoke callback for current object only", func() {
		currentObj := v1.Pod{}
		cr := testcr.Config{}
		reconcileState := callbacks.ReconcileStatePreCreate

		callbackArguments := new([]*callbacks.ReconcileCallbackArgs)
		callback := func(args *callbacks.ReconcileCallbackArgs) error {
			tmp := append(*callbackArguments, args)
			callbackArguments = &tmp
			return nil
		}

		By("registering callback")
		cd.AddCallback(&currentObj, callback)

		By("invoking callback")

		err := cd.InvokeCallbacks(log, cr, reconcileState, nil, &currentObj, recorder)

		Expect(err).ToNot(HaveOccurred())

		Expect(*callbackArguments).To(HaveLen(1))
		args := (*callbackArguments)[0]

		Expect(args.Logger).To(Equal(log))
		Expect(args.Scheme).To(Equal(s))
		Expect(args.Namespace).To(Equal(namespace))
		Expect(args.State).To(Equal(reconcileState))
		Expect(args.Client).To(Equal(client))
		Expect(args.CurrentObject).To(Equal(&currentObj))
		Expect(args.DesiredObject).To(BeNil())
		Expect(args.Resource).To(Equal(cr))
	})

	It("should register and invoke callback for nil current object in non-Pre-Create state", func() {
		desiredObj := v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "svc-1",
				Namespace: "svc-ns",
			},
		}
		cr := testcr.Config{}
		reconcileState := callbacks.ReconcileStatePostCreate

		callbackArguments := new([]*callbacks.ReconcileCallbackArgs)
		callback := func(args *callbacks.ReconcileCallbackArgs) error {
			tmp := append(*callbackArguments, args)
			callbackArguments = &tmp
			return nil
		}

		By("registering callback")
		cd.AddCallback(&desiredObj, callback)

		By("invoking callback")

		err := cd.InvokeCallbacks(log, cr, reconcileState, &desiredObj, nil, recorder)

		Expect(err).ToNot(HaveOccurred())

		Expect(*callbackArguments).To(HaveLen(1))
		args := (*callbackArguments)[0]

		Expect(args.Logger).To(Equal(log))
		Expect(args.Scheme).To(Equal(s))
		Expect(args.Namespace).To(Equal(namespace))
		Expect(args.State).To(Equal(reconcileState))
		Expect(args.Client).To(Equal(client))
		Expect(args.CurrentObject).ToNot(BeNil())
		currentSvc := args.CurrentObject.(*v1.Service)
		Expect(currentSvc.Name).To(Equal(existingService.Name))
		Expect(currentSvc.Namespace).To(Equal(existingService.Namespace))
		Expect(args.DesiredObject).To(Equal(&desiredObj))
		Expect(args.Resource).To(Equal(cr))
	})

	It("should register and invoke callback for nil current object in non-Pre-Create state and non-existing cluster instance", func() {
		desiredObj := v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "svc-missing",
				Namespace: "svc-ns",
			},
		}
		cr := testcr.Config{}
		reconcileState := callbacks.ReconcileStatePostCreate

		callbackArguments := new([]*callbacks.ReconcileCallbackArgs)
		callback := func(args *callbacks.ReconcileCallbackArgs) error {
			tmp := append(*callbackArguments, args)
			callbackArguments = &tmp
			return nil
		}

		By("registering callback")
		cd.AddCallback(&desiredObj, callback)

		By("invoking callback")

		err := cd.InvokeCallbacks(log, cr, reconcileState, &desiredObj, nil, recorder)

		Expect(err).ToNot(HaveOccurred())

		Expect(*callbackArguments).To(HaveLen(1))
		args := (*callbackArguments)[0]

		Expect(args.Logger).To(Equal(log))
		Expect(args.Scheme).To(Equal(s))
		Expect(args.Namespace).To(Equal(namespace))
		Expect(args.State).To(Equal(reconcileState))
		Expect(args.Client).To(Equal(client))
		Expect(args.CurrentObject).To(BeNil())
		Expect(args.DesiredObject).To(Equal(&desiredObj))
		Expect(args.Resource).To(Equal(cr))
	})

	It("should propagate callback error", func() {
		desiredObj := v1.Pod{}
		currentObj := v1.Pod{}
		cr := testcr.Config{}
		reconcileState := callbacks.ReconcileStatePostCreate

		callbackError := fmt.Errorf("Boom!")
		callback := func(args *callbacks.ReconcileCallbackArgs) error {
			return callbackError
		}

		By("registering callback")
		cd.AddCallback(&desiredObj, callback)

		By("invoking callback")

		err := cd.InvokeCallbacks(log, cr, reconcileState, &desiredObj, &currentObj, recorder)
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(callbackError))
	})

})
