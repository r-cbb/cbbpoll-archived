package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

/*
Server is a type that holds state for the app, along with routers and handlers.
 */
type Server struct {
	db     *sql.DB
	router *mux.Router
}

func NewServer() *Server {
	srv := Server{
		db: nil,
	}
	srv.Routes()
	return &srv
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.WriteHeader(status)
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			fmt.Printf("json encode: %s", err)
		} // TODO: logger
	}
}
