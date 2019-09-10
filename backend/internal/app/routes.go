package app

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

const v1 = "/v1"

func (s *Server) Routes() {
	s.router = mux.NewRouter()

	// API Health & Version
	s.router.HandleFunc(fmt.Sprintf("%s/ping", v1), s.handlePing()).Methods(http.MethodGet)

	// Teams
	s.router.HandleFunc(fmt.Sprintf("%s/teams", v1), s.handleAddTeam()).Methods(http.MethodPost)
	s.router.HandleFunc(fmt.Sprintf("%s/teams", v1), s.handleListTeams()).Methods(http.MethodGet)
	s.router.HandleFunc(fmt.Sprintf("%s/teams/{id:[0-9]+}", v1), s.handleGetTeam()).Methods(http.MethodGet).Name("team")

	// Users
	s.router.HandleFunc(fmt.Sprintf("%s/users/me", v1), s.handleUsersMe()).Methods(http.MethodGet)
	s.router.HandleFunc(fmt.Sprintf("%s/users/{name}", v1), s.handleGetUser()).Methods(http.MethodGet).Name("user")
}

func (s *Server) AuthRoutes() {
	newSession := s.router.HandleFunc(fmt.Sprintf("%s/sessions", v1), s.handleNewSession()).Methods(http.MethodPost)

	s.router.Use(s.AuthClient.Verifier())
	s.router.Use(SelectiveMiddleware(s.AuthClient.Authenticator, []*mux.Route{newSession}))
}

func (s *Server) handlePing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		version := models.VersionInfo{
			Version: s.version(),
		}
		s.respond(w, r, version, http.StatusOK)
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
			fmt.Println("concurrency error, retrying once")
			createdTeam, err = s.Db.AddTeam(newTeam)
		}

		if err != nil {
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		teamURL, err := s.router.Get("team").URLPath("id", fmt.Sprintf("%d", createdTeam.ID))
		if err != nil {
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", fmt.Sprintf("%s%s", s.host, teamURL.String()))
		s.respond(w, r, createdTeam, http.StatusCreated)
		return
	}
}

func (s *Server) handleGetTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}

		team, err := s.Db.GetTeam(intId)
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
		token := s.AuthClient.UserTokenFromCtx(r.Context())

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

		user, err := s.Db.GetUser(name)
		if err != nil {
			if errors.Kind(err) == errors.KindNotFound {
				s.respond(w, r, nil, http.StatusNotFound)
				return
			}

			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		s.respond(w, r, user, http.StatusOK)
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

		name, err := s.RedditClient.UsernameFromToken(accessToken)
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
		var createdUser models.User
		user, err := s.Db.GetUser(name)
		if errors.Kind(err) == errors.KindNotFound {
			// TODO: fill in IsAdmin by comparing username to list stored locally, maybe in a file.
			user = models.User{
				Nickname: name,
				IsAdmin:  false,
			}
			createdUser, err = s.Db.AddUser(user)
			if err != nil {
				s.respond(w, r, nil, http.StatusInternalServerError)
			}
			newUser = true
		}

		token, err := s.AuthClient.CreateJWT(user)
		if err != nil {
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		payload := struct {
			Nickname string `json:"nickname"`
			Token    string `json:"token"`
		}{
			Nickname: name,
			Token:    token,
		}

		var status = http.StatusOK
		if newUser {
			status = http.StatusCreated

			url, err := s.router.Get("user").URLPath("name", createdUser.Nickname)
			if err != nil {
				s.respond(w, r, nil, http.StatusInternalServerError)
				return
			}

			w.Header().Set("Location", fmt.Sprintf("%s%s", s.host, url))
		}
		s.respond(w, r, payload, status)
	}
}
