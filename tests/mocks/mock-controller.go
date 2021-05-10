package mocks

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

func (m *MockController) Reconcile(context.Context, reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
func (m *MockController) Watch(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
	m.WatchCalls = append(m.WatchCalls, WatchCall{src, eventhandler, predicates})
	return nil
}
func (m *MockController) Start(context.Context) error {
	return nil
}

func (m *MockController) GetLogger() logr.Logger {
	return log.Log
}
