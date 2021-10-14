package service

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/rpc/api"
	"sync/atomic"
)

type Service struct {
	ctx      context.Context
	cancel   context.CancelFunc
	started  int32
	shutdown int32
}

func (s *Service) Start(ctx context.Context) error {
	if atomic.AddInt32(&s.started, 1) != 1 {
		return fmt.Errorf("Service is already in the process of started")
	}
	s.ctx, s.cancel = context.WithCancel(ctx)

	return nil
}

func (s *Service) Stop() error {
	if atomic.AddInt32(&s.shutdown, 1) != 1 {
		return fmt.Errorf("Service is already in the process of shutting down")
	}
	defer func() {
		s.cancel()
	}()
	return nil
}

func (s *Service) IsStarted() bool {
	return atomic.LoadInt32(&s.started) != 0
}

func (s *Service) IsShutdown() bool {
	return atomic.LoadInt32(&s.shutdown) != 0
}

func (s *Service) APIs() []api.API {
	return nil
}

func (s *Service) Status() error {
	return nil
}
