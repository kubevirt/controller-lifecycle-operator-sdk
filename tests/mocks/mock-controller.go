package mocks

import (
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type WatchCall struct {
	Src          source.Source
	Eventhandler handler.EventHandler
	Predicates   []predicate.Predicate
}

type MockController struct {
	WatchCalls []WatchCall
}

func (m *MockController) Reconcile(reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
func (m *MockController) Watch(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
	m.WatchCalls = append(m.WatchCalls, WatchCall{src, eventhandler, predicates})
	return nil
}
func (m *MockController) Start(stop <-chan struct{}) error {
	return nil
}
