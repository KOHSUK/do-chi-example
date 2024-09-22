package main

import (
	"fmt"
	"net/http"
	"os"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
)

const (
	_ = iota
	GET
	POST
	PUT
	DELETE
)

type Route interface {
	http.Handler

	Path() string
	Method() int
}

type HelloHandler struct{}

func NewHelloHandler(i do.Injector) (*HelloHandler, error) {
	return &HelloHandler{}, nil
}

func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func (h *HelloHandler) Path() string {
	return "/hello"
}

func (h *HelloHandler) Method() int {
	return GET
}

type ByeHandler struct{}

func NewByeHandler(i do.Injector) (*ByeHandler, error) {
	return &ByeHandler{}, nil
}

func (h *ByeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Bye, World!"))
}

func (h *ByeHandler) Path() string {
	return "/bye"
}

func (h *ByeHandler) Method() int {
	return GET
}

type Server struct {
	mux *chi.Mux
}

func RegisterRoutes[H Route](r chi.Router, i do.Injector) {
	handler, err := do.Invoke[H](i)
	if err != nil {
		panic(err)
	}

	var method string
	switch handler.Method() {
	case GET:
		method = http.MethodGet
	case POST:
		method = http.MethodPost
	case PUT:
		method = http.MethodPut
	case DELETE:
		method = http.MethodDelete
	default:
		panic("invalid method")
	}

	r.Method(method, handler.Path(), handler)
}

func NewServer(i do.Injector) (*Server, error) {
	r := chi.NewRouter()

	RegisterRoutes[*HelloHandler](r, i)
	RegisterRoutes[*ByeHandler](r, i)

	return &Server{mux: r}, nil
}

func (s *Server) Start() error {
	fmt.Println("Starting server")

	err := http.ListenAndServe(":8080", s.mux)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	injector := do.New()

	do.Provide(injector, NewHelloHandler)
	do.Provide(injector, NewByeHandler)
	do.Provide(injector, NewServer)

	server, err := do.Invoke[*Server](injector)
	if err != nil {
		panic(err)
	}

	server.Start()

	injector.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
}
