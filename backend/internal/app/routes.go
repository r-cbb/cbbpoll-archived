package app

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) Routes() {
	s.router = mux.NewRouter()
	s.router.HandleFunc("/", s.handleIndex())
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, struct{ Foo string}{Foo: "hello world"}, 200)
	}
}