package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	// ErrGeneric is a generic, internal error.
	ErrGeneric = "hive.internal.generic"
	// ErrNetwork is a internal network error.
	ErrNetwork = "hive.internal.network"
	// ErrRequest is a internal request error.
	ErrRequest  = "hive.internal.request"
	jsonMIME    = "application/json"
	queryFormat = "%s://%s:%s/%s"
)

var log = logrus.New()

func init() {
	log.Out = os.Stderr
	log.Level = logrus.DebugLevel
}

// Error stores service error IDs, texts and status codes.
type Error struct {
	ID     string
	Text   string
	Status int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s [%d]: %s", e.ID, e.Status, e.Text)
}

// Instance generates a new instance encapsuling the reference error.
func (e *Error) Instance(err error) *Error {
	return &Error{ID: e.ID, Text: err.Error(), Status: e.Status}
}

// Method is a abstract interface representing a service method.
type Method interface {
	GetRequestType() reflect.Type
	GetResponseType() reflect.Type
	GetName() string
	HandleRequest(*Service, interface{}) (interface{}, *Error)
}

type basicMethod struct {
	RequestType  reflect.Type `json:"request"`
	ResponseType reflect.Type `json:"response"`
	Name         string       `json:"name"`
	handle       func(interface{}) (interface{}, *Error)
}

func (b basicMethod) GetRequestType() reflect.Type {
	return b.RequestType
}

func (b basicMethod) GetResponseType() reflect.Type {
	return b.ResponseType
}

func (b basicMethod) GetName() string {
	return b.Name
}

func (b basicMethod) HandleRequest(service *Service, req interface{}) (interface{}, *Error) {
	return b.handle(req)
}

type contextualMethod struct {
	basicMethod
	contextualHandle func(*Service, interface{}) (interface{}, *Error)
}

func (c contextualMethod) HandleRequest(service *Service, req interface{}) (interface{}, *Error) {
	return c.contextualHandle(service, req)
}

// Service is a microservice infrastructure abstraction.
type Service struct {
	Name         string                 `json:"name"`
	DNSName      string                 `json:"dnsname"`
	Version      string                 `json:"version"`
	Methods      map[string]Method      `json:"methods"`
	Protocol     string                 `json:"protocol"`
	Socket       string                 `json:"socket"`
	Timeout      time.Duration          `json:"timeout"`
	Context      map[string]interface{} `json:"context"`
	ForwardLocal bool                   `json:"forwardLocal"`
	KnownErrors  map[string]Error       `json:"errors"`
}

// New creates a new service.
func New(name, version string) Service {
	return Service{
		Name:         name,
		DNSName:      name,
		Version:      version,
		Methods:      make(map[string]Method),
		Protocol:     "http",
		Socket:       ":80",
		Timeout:      time.Second * 10,
		Context:      make(map[string]interface{}),
		ForwardLocal: false,
		KnownErrors:  makeDefaultErrorMap(),
	}
}

func makeDefaultErrorMap() map[string]Error {
	return map[string]Error{
		"hive.internal.generic": Error{ID: "hive.internal.generic", Status: http.StatusInternalServerError},
		"hive.internal.request": Error{ID: "hive.internal.request", Status: http.StatusBadRequest},
		"hive.internal.network": Error{ID: "hive.internal.network", Status: http.StatusInternalServerError},
	}
}

// RegisterError registers a new error code.
func (service Service) RegisterError(e Error) {
	if service.KnownErrors == nil {
		service.KnownErrors = makeDefaultErrorMap()
	}
	service.KnownErrors[e.ID] = e
}

// Throw generates a new error instance.
func (service Service) Throw(id string, err error) *Error {
	e, ok := service.KnownErrors[id]
	if !ok {
		e = service.KnownErrors[ErrGeneric]
	}
	return e.Instance(err)
}

// SThrow generates a new error code with the specific text.
func (service Service) SThrow(id string, text string) *Error {
	return service.Throw(id, errors.New(text))
}

func sendError(w http.ResponseWriter, e *Error) {
	json, err := json.Marshal(e)
	w.WriteHeader(e.Status)
	_, err = w.Write(json)
	if err != nil {
		http.Error(w, "failed to write error: "+err.Error(), http.StatusInternalServerError)
	}
}

// LogInfo logs an message of log level Info.
func (service Service) LogInfo(args ...interface{}) {
	log.Info(args...)
}

// LogDebug logs an message of log level Debug.
func (service Service) LogDebug(args ...interface{}) {
	log.Debug(args...)
}

// LogError logs an message of log level Error.
func (service Service) LogError(args ...interface{}) {
	log.Error(args...)
}

// Send a request to the service.
func (service Service) Send(name string, request interface{}) (interface{}, *Error) {
	if service.ForwardLocal {
		return service.Methods[name].HandleRequest(&service, request)
	}
	method, found := service.Methods[name]
	if !found {
		return nil, service.SThrow(ErrGeneric, "Method "+name+" not found")
	}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, service.Throw(ErrGeneric, err)
	}
	url := fmt.Sprintf(queryFormat, service.Protocol, service.DNSName, service.Socket, method.GetName())
	service.LogDebug("Querying service under ", url)
	resp, err := http.Post(url, jsonMIME, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, service.Throw(ErrNetwork, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, service.Throw(ErrNetwork, err)
	}
	if resp.StatusCode != http.StatusOK {
		errBox := Error{}
		if err := json.Unmarshal(body, &errBox); err != nil {
			return nil, service.Throw(ErrGeneric, err)
		}
		return nil, &errBox
	}
	response := reflect.New(method.GetResponseType()).Interface()
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, service.Throw(ErrGeneric, err)
	}
	return response, nil
}

// Register a service method.
func (service Service) Register(method Method) {
	if service.Methods == nil {
		service.Methods = make(map[string]Method)
	}
	service.Methods[method.GetName()] = method
}

// Run the service.
func (service Service) Run() {
	mux := http.NewServeMux()
	service.Register(NewMethod("", new(struct{}), new(Service), func(interface{}) (interface{}, *Error) {
		return &service, nil
	}))
	// add all methods to server mux
	for name, method := range service.Methods {
		mux.HandleFunc("/"+name, newMethodHandler(&service, method))
	}
	service.LogInfo("All methods successfully activated")
	// init server
	server := &http.Server{
		Addr:           service.Socket,
		Handler:        mux,
		ReadTimeout:    service.Timeout,
		WriteTimeout:   service.Timeout,
		MaxHeaderBytes: 1 << 20,
	}
	// startup
	log.Fatal(server.ListenAndServe())
}

// Create a new method handler.j
func newMethodHandler(service *Service, method Method) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		service.LogDebug("Got request for /" + method.GetName())
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sendError(w, service.Throw(ErrRequest, err))
			return
		}
		request := reflect.New(method.GetRequestType()).Interface()
		err = json.Unmarshal(body, request)
		if err != nil {
			sendError(w, service.Throw(ErrRequest, err))
			return
		}
		response, respErr := method.HandleRequest(service, request)
		if respErr != nil {
			sendError(w, service.SThrow(respErr.ID, respErr.Text))
			return
		}
		json, err := json.Marshal(response)
		if err != nil {
			sendError(w, service.Throw(ErrGeneric, err))
			return
		}
		_, err = w.Write(json)
		if err != nil {
			sendError(w, service.Throw(ErrNetwork, err))
			return
		}
	}
}

// NewMethod creates a new default method handler.
func NewMethod(name string, requestType interface{}, responseType interface{}, handler func(interface{}) (interface{}, *Error)) Method {
	return basicMethod{
		Name:         name,
		RequestType:  reflect.TypeOf(requestType),
		ResponseType: reflect.TypeOf(responseType),
		handle:       handler,
	}
}

// NewContextualMethod creates a new context-aware method handler.
func NewContextualMethod(name string, requestType interface{}, responseType interface{}, handler func(*Service, interface{}) (interface{}, *Error)) Method {
	return contextualMethod{
		basicMethod: basicMethod{
			Name:         name,
			RequestType:  reflect.TypeOf(requestType),
			ResponseType: reflect.TypeOf(responseType),
		},
		contextualHandle: handler,
	}
}
