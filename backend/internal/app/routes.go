package app

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"

	"github.com/r-cbb/cbbpoll/backend/internal/cbbpoll"
)

func (s *Server) Routes() {
	s.router = mux.NewRouter()
	s.router.HandleFunc("/", s.handleIndex())
	s.router.HandleFunc("/team", s.handleAddTeam()).Methods(http.MethodPost)
	s.router.HandleFunc("/team/{id:[0-9]+}", s.handleGetTeam()).Methods(http.MethodGet)
}

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, struct{ Foo string}{Foo: "hello world"}, 200)
	}
}

func (s *Server) handleAddTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newTeam cbbpoll.Team
		err := s.decode(w, r, &newTeam)
		if err != nil {
			fmt.Printf("bad input")
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}

		id, err := s.db.AddTeam(newTeam)
		if err != nil {
			fmt.Printf("db error: %v", err.Error())
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		s.respond(w, r, id, http.StatusOK)
		return
	}
}

func (s *Server) handleGetTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		intId, err := strconv.Atoi(id)
		team, err := s.db.GetTeam(int64(intId))
		if err != nil {
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}
		s.respond(w, r, team, http.StatusOK)
		return
	}
}