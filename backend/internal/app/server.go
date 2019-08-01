package app

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/r-cbb/cbbpoll/backend/internal/db"
	"net/http"
)

/*
Server is a type that holds state for the app, along with routers and handlers.
*/
type Server struct {
	db     *db.DBClient
	router *mux.Router
}

func NewServer() (*Server, error) {
	srv := Server{}
	var err error
	srv.db, err = db.NewDBClient("cbbpoll")
	if err != nil {
		fmt.Printf("error: %v", err.Error())
		return nil, err
	}
	srv.Routes()
	return &srv, nil
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

func (s *Server) decode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}