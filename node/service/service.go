package service

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/rpc/api"
	"reflect"
	"sync/atomic"
)

type Service struct {
	ctx      context.Context
	cancel   context.CancelFunc
	started  int32
	shutdown int32
	services *ServiceRegistry
	kind     reflect.Type
}

func (s *Service) Start() error {
	if atomic.AddInt32(&s.started, 1) != 1 {
		return fmt.Errorf("Service is already in the process of started")
	}
	if s.kind != nil {
		log.Debug(fmt.Sprintf("(%v) service start", s.kind))
	}
	s.InitContext()
	if s.services != nil {
		return s.services.StartAll()
	}
	return nil
}

func (s *Service) Stop() error {
	if atomic.AddInt32(&s.shutdown, 1) != 1 {
		return fmt.Errorf("Service is already in the process of shutting down")
	}
	if s.kind != nil {
		log.Debug(fmt.Sprintf("(%v) service stop", s.kind))
	}

	s.cancel()
	if s.services != nil {
		return s.services.StopAll()
	}
	return nil
}

func (s *Service) IsStarted() bool {
	return atomic.LoadInt32(&s.started) != 0
}

func (s *Service) IsShutdown() bool {
	return atomic.LoadInt32(&s.shutdown) != 0
}

func (s *Service) APIs() []api.API {
	if s.services != nil {
		return s.services.APIs()
	}
	return nil
}

func (s *Service) Status() error {
	return nil
}

func (s *Service) Context() context.Context {
	return s.ctx
}

func (s *Service) InitContext() {
	if s.ctx == nil {
		s.ctx, s.cancel = context.WithCancel(context.Background())
	}
}

func (s *Service) Services() *ServiceRegistry {
	return s.services
}

func (s *Service) InitServices() {
	s.services = NewServiceRegistry()
}

func (s *Service) SetKind(kind reflect.Type) {
	s.kind = kind
}
