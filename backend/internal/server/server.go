package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/r-cbb/cbbpoll/internal/app"
	"github.com/r-cbb/cbbpoll/internal/auth"
	"github.com/r-cbb/cbbpoll/internal/errors"
)

/*
Server is a type that holds state for the server, along with routers and handlers.
*/
type Server struct {
	App          *app.PollService
	AuthClient   auth.Client
	RedditClient RedditClient
	router       *mux.Router
	host         string
}

func NewServer() *Server {
	srv := Server{}
	srv.Routes()

	return &srv
}

func (s *Server) SetHost(host string) {
	s.host = host
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			fmt.Printf("json encode: %s", err)
		}
	}
}

const maxRequestSize = 1 << 20 // 1 MB

func (s *Server) decode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	const op errors.Op = "server.decode"
	return json.NewDecoder(io.LimitReader(r.Body, maxRequestSize)).Decode(&v)
}

func (s Server) version() string {
	return "v0.1.0"
}
