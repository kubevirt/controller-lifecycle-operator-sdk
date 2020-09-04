[![Build Status](https://travis-ci.com/kubevirt/controller-lifecycle-operator-sdk.svg?branch=master)](https://travis-ci.com/kubevirt/controller-lifecycle-operator-sdk)
# controller-lifecycle-operator-sdk
Library helping in building  HCO (Hyperconverged Cluster Operator - https://github.com/kubevirt/hyperconverged-cluster-operator) compatible Kubernetes operators.

## Building parts
The controller-lifecycle-operator-sdk consists of following parts, that can be used together or separately:
- [API](#API) (`pkg/sdk/api`) package defining common types and constants that are required by the other parts of the library and that are compatible with HCO;
- [SDK](pkg/sdk) package providing several helper functions used in the other parts of the library;
- [Callbacks](#Callbacks) (`pkg/sdk/callbacks`) package providing object-type bound callback registration and dispatching facilities - used to execute additional resource-management login in predefined places in the reconciliation loop;
- [Resources](pkg/sdk/resources) package providing resource definition helpers (deployment, service, etc. builders);
- [OpenAPI](pkg/sdk/resources/openapi) package providing OpenAPI definition of the common `Status` structure;
- [Reconciler](#Reconciler) (`pkg/sdk/reconciler`)  package providing `Reconciler` structure responsible for executing operator lifecycle reconciliation

### API
The `pkg/sdk/api` provides definition of a `Status` structure that has to be used in operator configuration Custom Resource as follows to allow the `Reconciler` to work properly:

```go
// ConfigStatus defines the observed state of Config
type ConfigStatus struct {
	sdkapi.Status `json:",inline"`
}

// Config is the Schema for the config API
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec   `json:"spec,omitempty"`
	Status ConfigStatus `json:"status,omitempty"`
}
```

The `sdkapi.Status` is inlined in a configuration CR-specific `ConfigStatus` structure.

The package defines also a set of phases that the configuration CR can be assigned.

### Callbacks
The `pkg/sdk/callbacks` package defines `CallbackDispatcher` structure that allows its clients to register and notify callbacks assigned to Kubernetes resource object types (anything that fulfills `runtime.Object` interface).
`CallbackDispatcher` implements following interface: 
```go
type CallbackDispatcher interface {
	// AddCallback registers a callback for given object type
	AddCallback(runtime.Object, callbacks.ReconcileCallback)

	// InvokeCallbacks executes callbacks for desired/current object type
	InvokeCallbacks(l logr.Logger, cr interface{}, s callbacks.ReconcileState, desiredObj, currentObj runtime.Object, recorder record.EventRecorder) error
} 
``` 

`AddCallback` method registers `callback` function under the _type_ of `obj` key; there can be multiple callbacks registered for the same object type.
`InvokeCallbacks` method executes all callbacks registered under the type of `desiredObj` and `currentObj`; `s` provides information about the stage of reconciliation when the call is made. `desiredObj` and `currentObj` are resources representing desired state of some object, and the current one (as stored in the cluster). It is the callback's responsibility to move the object to the desired state.


### Reconciler
`Reconciler` structure from `pkg/sdk/reconciler` package is meant to work as a delegate for a `Reconcile` method in a controller that serves as a HCO-deployed operator. That method is responsible for managing both operator's deployment and state of any resources under its purview. 

The `Reconciler` to work properly requires its client to provide an implementation of a `CrManager` interface. The interface defines several methods that are implementor domain-specific, like creation of a configuration Custom Resource, retrieval of `sdkapi.Status` sub-resource from the configuration CustomResource or others that can be found in [reconciler.go](pkg/sdk/reconciler/reconciler.go).

## Reference implementation
[The reference implementation](examples/sample-operator) shows how the SDK can be used to manage other resources.                