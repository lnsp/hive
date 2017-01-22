package randomcalc

import (
	"math/rand"

	"github.com/lnsp/hive/examples/addition"
	"github.com/lnsp/hive/examples/subtraction"
	"github.com/lnsp/hive/lib/discovery"
	"github.com/lnsp/hive/lib/service"
)

type CalculateRequest struct {
	A, B int
}

type CalculateResponse struct {
	Result int
}

var Discovery discovery.Discovery
var Calculate service.Method
var Service service.Service

func calculateHandler(request interface{}) (interface{}, error) {
	req := request.(*CalculateRequest)

	response := CalculateResponse{}
	state := rand.Intn(2)
	if state == 0 {
		addResponse, err := Discovery.Send("addition", "add", addition.AddRequest{A: req.A, B: req.B})
		if err != nil {
			return nil, err
		}
		response.Result = addResponse.(*addition.AddResponse).Result
	} else {
		subResponse, err := Discovery.Send("subtraction", "subtract", subtraction.SubtractRequest{A: req.A, B: req.B})
		if err != nil {
			return nil, err
		}
		response.Result = subResponse.(*subtraction.SubtractResponse).Result
	}

	return &response, nil
}

func init() {
	Discovery = discovery.New()
	Service = service.New("randomcalc", "0.1.0")
	Service.Socket = ":8081"
	Calculate = service.NewMethod("calculate", CalculateRequest{}, CalculateResponse{}, calculateHandler)

	Service.Register(Calculate)
	Discovery.Register(addition.Service)
	Discovery.Register(subtraction.Service)
}
