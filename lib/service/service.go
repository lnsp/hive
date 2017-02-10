package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/Sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.Out = os.Stderr
	log.Level = logrus.DebugLevel
}

// A microservice method.
type Method struct {
	RequestType  reflect.Type `json:"request"`
	ResponseType reflect.Type `json:"response"`
	Name         string       `json:"name"`
	handle       func(interface{}) (interface{}, error)
}

// A microservice.
type Service struct {
	Name     string            `json:"name"`
	DNSName  string            `json:"dnsname"`
	Version  string            `json:"version"`
	Methods  map[string]Method `json:"methods"`
	Protocol string            `json:"protocol"`
	Socket   string            `json:"socket"`
	Timeout  time.Duration     `json:"timeout"`
}

// Create a new service.
func New(name, version string) Service {
	return Service{
		Name:     name,
		DNSName:  name,
		Version:  version,
		Methods:  make(map[string]Method),
		Protocol: "http",
		Socket:   ":80",
		Timeout:  time.Second * 10,
	}
}

// Send a request to the service.
func (service Service) Send(name string, request interface{}) (interface{}, error) {
	log.Info("initiating request to service", service.Name, "->", name)
	method, found := service.Methods[name]
	if !found {
		return nil, errors.New("method not found")
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := service.Protocol + "://" + service.DNSName + service.Socket + "/" + method.Name
	log.Info("request url is: ", url)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, errors.New("failed service request: " + err.Error())
	}
	defer resp.Body.Close()

	log.Debug("received response from service ", service.DNSName)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed service request: " + err.Error())
	}

	response := reflect.New(method.ResponseType).Interface()
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, errors.New("failed to parse response: " + err.Error())
	}
	log.Debug("generated ", method.ResponseType.String(), " from response")

	return response, nil
}

// Register a service method.
func (service Service) Register(method Method) {
	if service.Methods == nil {
		service.Methods = make(map[string]Method)
	}
	service.Methods[method.Name] = method
}

// Run the service.
func (service Service) Run() {
	mux := http.NewServeMux()
	service.Register(NewMethod("", new(struct{}), new(Service), func(interface{}) (interface{}, error) {
		return &service, nil
	}))

	// add all methods to server mux
	for name, method := range service.Methods {
		mux.HandleFunc("/"+name, newMethodHandler(method))
	}
	log.Info("register all methods successful")

	// init server
	server := &http.Server{
		Addr:           service.Socket,
		Handler:        mux,
		ReadTimeout:    service.Timeout,
		WriteTimeout:   service.Timeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Info("starting up service")

	// startup
	log.Fatal(server.ListenAndServe())
}

// Create a new method handler.j
func newMethodHandler(method Method) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("got request for /" + method.Name)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
			return
		}

		request := reflect.New(method.RequestType).Interface()
		err = json.Unmarshal(body, request)
		if err != nil {
			http.Error(w, "invalid json request: "+err.Error(), http.StatusBadRequest)
			return
		}
		log.Debug("created request object of type ", method.RequestType)

		response, err := method.handle(request)
		if err != nil {
			http.Error(w, "response error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		json, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "failed to pack json response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(json)
		if err != nil {
			http.Error(w, "failed to write response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Create a new method.
func NewMethod(name string, requestType interface{}, responseType interface{}, handler func(interface{}) (interface{}, error)) Method {
	return Method{
		Name:         name,
		RequestType:  reflect.TypeOf(requestType),
		ResponseType: reflect.TypeOf(responseType),
		handle:       handler,
	}
}
