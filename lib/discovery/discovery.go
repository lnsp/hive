package discovery

import "github.com/lnsp/hive/lib/service"

type Discovery struct {
	services map[string]service.Service
}

func (discovery Discovery) Register(service service.Service) {
	discovery.services[service.Name] = service
}

func (discovery Discovery) Retrieve(name string) service.Service {
	return discovery.services[name]
}

func (discovery Discovery) Send(name string, method string, request interface{}) (interface{}, *service.Error) {
	return discovery.services[name].Send(method, request)
}

func New() Discovery {
	return Discovery{
		make(map[string]service.Service, 0),
	}
}
