package discover

import "github.com/lnsp/hive/lib/service"

var services map[string]service.Service

func Register(service service.Service) {
	services[service.Name] = service
}

func Retrieve(name string) service.Service {
	return services[name]
}

func Send(name string, method string, request interface{}) (interface{}, error) {
	return services[name].Send(method, request)
}
