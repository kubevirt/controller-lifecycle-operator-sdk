package mocks

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type WatchCall struct {
	Src source.Source
}

type MockController struct {
	WatchCalls []WatchCall
}

func (m *MockController) Reconcile(context.Context, reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
func (m *MockController) Watch(src source.Source) error {
	m.WatchCalls = append(m.WatchCalls, WatchCall{src})
	return nil
}
func (m *MockController) Start(context.Context) error {
	return nil
}

func (m *MockController) GetLogger() logr.Logger {
	return log.Log
}
