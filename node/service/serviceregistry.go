package service

import (
	"context"
	"fmt"
	"github.com/Qitmeer/qitmeer/rpc/api"
	"reflect"
)

type IService interface {
	// APIs retrieves the list of RPC descriptors the service provides
	APIs() []api.API

	// Start is called after all services have been constructed and the networking
	// layer was also initialized to spawn any goroutines required by the service.
	Start() error

	// Stop terminates all goroutines belonging to the service, blocking until they
	// are all terminated.
	Stop() error

	Status() error

	IsStarted() bool

	IsShutdown() bool

	Context() context.Context
}

type ServiceRegistry struct {
	services     map[reflect.Type]IService // map of types to services.
	serviceTypes []reflect.Type            // keep an ordered slice of registered service types.
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[reflect.Type]IService),
	}
}

func (s *ServiceRegistry) StartAll() error {
	log.Debug(fmt.Sprintf("Starting %d services: %v", len(s.serviceTypes), s.serviceTypes))
	for _, kind := range s.serviceTypes {
		log.Debug(fmt.Sprintf("Starting service type %v", kind))
		err := s.services[kind].Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ServiceRegistry) StopAll() error {
	result := ""
	for i := len(s.serviceTypes) - 1; i >= 0; i-- {
		kind := s.serviceTypes[i]
		service := s.services[kind]
		if err := service.Stop(); err != nil {
			log.Error(fmt.Sprintf("Could not stop the following service: %v, %v", kind, err))
			result += fmt.Sprintf("(%v)", kind)
		}
	}
	if len(result) > 0 {
		return fmt.Errorf("%s", result)
	}
	return nil
}

func (s *ServiceRegistry) Statuses() map[reflect.Type]error {
	m := make(map[reflect.Type]error)
	for _, kind := range s.serviceTypes {
		m[kind] = s.services[kind].Status()
	}
	return m
}

func (s *ServiceRegistry) RegisterService(service IService) error {
	kind := reflect.TypeOf(service)
	if _, exists := s.services[kind]; exists {
		return fmt.Errorf("service already exists: %v", kind)
	}
	s.services[kind] = service
	s.serviceTypes = append(s.serviceTypes, kind)
	return nil
}

func (s *ServiceRegistry) FetchService(service interface{}) error {
	if reflect.TypeOf(service).Kind() != reflect.Ptr {
		return fmt.Errorf("input must be of pointer type, received value type instead: %T", service)
	}
	element := reflect.ValueOf(service).Elem()
	if running, ok := s.services[element.Type()]; ok {
		element.Set(reflect.ValueOf(running))
		return nil
	}
	return fmt.Errorf("unknown service: %T", service)
}

func (s *ServiceRegistry) GetServices() map[reflect.Type]IService {
	return s.services
}

func (s *ServiceRegistry) APIs() []api.API {
	apis := []api.API{}
	for _, kind := range s.serviceTypes {
		a := s.services[kind].APIs()
		if a == nil {
			continue
		}
		apis = append(apis, a...)
	}
	return apis
}
