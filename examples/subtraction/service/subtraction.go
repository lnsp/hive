package service

import (
	"github.com/lnsp/hive/lib/service"
)

type SubtractRequest struct {
	A, B int
}

type SubtractResponse struct {
	Result int
}

var Subtract service.Method
var Service service.Service

func subtractHandler(request interface{}) (interface{}, *service.Error) {
	req := request.(*SubtractRequest)
	return &SubtractResponse{
		Result: req.A - req.B,
	}, nil
}

func init() {
	Service = service.New("subtraction", "0.1.0")
	Subtract = service.NewMethod("subtract", SubtractRequest{}, SubtractResponse{}, subtractHandler)

	Service.Register(Subtract)
}
