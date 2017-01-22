package addition

import (
	"github.com/lnsp/hive/lib/service"
)

type AddRequest struct {
	A, B int
}

type AddResponse struct {
	Result int
}

var Add service.Method
var Service service.Service

func addHandler(request interface{}) (interface{}, error) {
	req := request.(*AddRequest)
	return &AddResponse{
		Result: req.A + req.B,
	}, nil
}

func init() {
	Service = service.New("addition", "0.1.0")
	Add = service.NewMethod("add", AddRequest{}, AddResponse{}, addHandler)

	Service.Register(Add)
}
