package app

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/jwtauth"
	"github.com/gorilla/mux"

	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

func (s *Server) Routes() {
	s.router = mux.NewRouter()

	// API Health & Version
	s.router.HandleFunc("/", s.handlePing()).Methods(http.MethodGet)
	s.router.HandleFunc("/ping", s.handlePing()).Methods(http.MethodGet)

	// Teams
	s.router.HandleFunc("/teams", s.handleAddTeam()).Methods(http.MethodPost)
	s.router.HandleFunc("/teams", s.handleListTeams()).Methods(http.MethodGet)
	s.router.HandleFunc("/teams/{id:[0-9]+}", s.handleGetTeam()).Methods(http.MethodGet)

	// Users
	s.router.HandleFunc("/users/me", s.handleUsersMe()).Methods(http.MethodGet)
	s.router.HandleFunc("/users/{name}", s.handleGetUser()).Methods(http.MethodGet)
}

func (s *Server) AuthRoutes() {
	newSession := s.router.HandleFunc("/sessions", s.handleNewSession()).Methods(http.MethodPost)

	s.router.Use(jwtauth.Verifier(s.TokenAuth))
	s.router.Use(Authenticator([]*mux.Route{newSession}))
}

func (s *Server) handlePing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, struct{ Version string }{Version: s.version()}, http.StatusOK)
	}
}

func (s *Server) handleAddTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newTeam models.Team
		err := s.decode(w, r, &newTeam)
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}

		createdTeam, err := s.Db.AddTeam(newTeam)

		if errors.Kind(err) == errors.KindConcurrencyProblem {
			// Retry once
			createdTeam, err = s.Db.AddTeam(newTeam)
		}

		if err != nil {
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		s.respond(w, r, createdTeam, http.StatusOK)
		return
	}
}

func (s *Server) handleGetTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		intId, err := strconv.Atoi(id)
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}

		team, err := s.Db.GetTeam(int64(intId))
		if err != nil {
			if errors.Kind(err) == errors.KindNotFound {
				s.respond(w, r, nil, http.StatusNotFound)
				return
			}

			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		s.respond(w, r, team, http.StatusOK)
		return
	}
}

func (s *Server) handleListTeams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teams, err := s.Db.GetTeams()
		if err != nil {
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		s.respond(w, r, teams, http.StatusOK)
	}
}

func (s *Server) handleUsersMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := models.UserTokenFromContext(r.Context())

		if !token.LoggedIn() {
			s.respond(w, r, nil, http.StatusUnauthorized)
			return
		}

		user, err := s.Db.GetUser(token.Nickname)
		if err != nil {
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		s.respond(w, r, user, http.StatusOK)
		return
	}
}

func (s *Server) handleGetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]

		team, err := s.Db.GetUser(name)
		if err != nil {
			if errors.Kind(err) == errors.KindNotFound {
				s.respond(w, r, nil, http.StatusNotFound)
				return
			}

			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		s.respond(w, r, team, http.StatusOK)
		return
	}
}

func (s *Server) handleNewSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		splitHeader := strings.Split(authHeader, "Bearer")
		if len(splitHeader) != 2 { // Bearer token not in proper format
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}

		accessToken := strings.TrimSpace(splitHeader[1])

		name, err := usernameFromRedditToken(accessToken)
		if err != nil {
			if errors.Kind(err) == errors.KindAuthError {
				s.respond(w, r, nil, http.StatusUnauthorized) // received a 401 from reddit
				return
			} else if errors.Kind(err) == errors.KindServiceUnavailable {
				s.respond(w, r, nil, http.StatusServiceUnavailable) // Possible reddit api is down
				return
			} else {
				s.respond(w, r, nil, http.StatusInternalServerError)
				return
			}
		}

		// Get user
		var newUser bool
		user, err := s.Db.GetUser(name)
		if errors.Kind(err) == errors.KindNotFound {
			// TODO: fill in IsAdmin by comparing username to list stored locally, maybe in a file.
			user = models.User{
				Nickname: name,
				IsAdmin: false,
			}
			_, err := s.Db.AddUser(user)
			if err != nil {
				s.respond(w, r, nil, http.StatusInternalServerError)
			}
			newUser = true
		}

		out, err := createJWT(user, s.TokenAuth)
		if err != nil {
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		payload := struct {
			Nickname string
			Token string
		}{
			Nickname: name,
			Token: out,
		}

		var status = http.StatusOK
		if newUser {
			status = http.StatusCreated
		}
		s.respond(w, r, payload, status)
	}
}