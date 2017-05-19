package service

import (
	"github.com/lnsp/hive/lib/service"
)

type HelloRequest struct {
	Name string `json:"name"`
}

type HelloResponse struct {
	Text string `json:"text"`
}

var SayHello service.Method
var Service service.Service

func helloHandler(request interface{}) (interface{}, *service.Error) {
	helloRequest := request.(*HelloRequest)
	helloResponse := HelloResponse{
		Text: "Hello, " + helloRequest.Name + "!",
	}
	return &helloResponse, nil
}

func init() {
	Service = service.New("helloworld", "0.1.0")
	SayHello = service.NewMethod("hello", HelloRequest{}, HelloResponse{}, helloHandler)

	Service.Register(SayHello)
}
