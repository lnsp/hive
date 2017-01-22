package randomcalc

import (
	"github.com/lnsp/hive/examples/calculator/addition"
	"github.com/lnsp/hive/examples/calculator/subtraction"
	"github.com/lnsp/hive/lib/discovery"
	"github.com/lnsp/hive/lib/service"
)

type CalculateRequest struct {
	A, B int
}

type CalculateResponse struct {
	Result int
}

var Calculate service.Method
var Service service.Service
var Discovery discovery.Discovery

func calculateHandler(request interface{}) (interface{}, error) {
	req := request.(*CalculateRequest)
	addReq := addition.AddRequest{A: req.A, B: req.B}

	addResponse, err := Discovery.Send("addition", "add", addReq)
	if err != nil {
		return nil, err
	}
	resp := addResponse.(*addition.AddResponse)

	return CalculateResponse{Result: resp.Result}, nil
}

func init() {
	Service = service.New("randomcalc", "0.1.0")
	Calculate = service.NewMethod("calculate", CalculateRequest{}, CalculateResponse{}, calculateHandler)

	Service.Register(Calculate)
	Discovery.Register(addition.Service)
	Discovery.Register(subtraction.Service)
}
