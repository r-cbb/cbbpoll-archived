package server

import (
	"fmt"
	"log"
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
	s.router.HandleFunc(fmt.Sprintf("%s/users", v1), s.handleAddUser()).Methods(http.MethodPost)
	s.router.HandleFunc(fmt.Sprintf("%s/users/me", v1), s.handleUsersMe()).Methods(http.MethodGet)
	s.router.HandleFunc(fmt.Sprintf("%s/users/{name}", v1), s.handleGetUser()).Methods(http.MethodGet).Name("user")
	s.router.HandleFunc(fmt.Sprintf("%s/users/{name}", v1), s.handleUpdateUser()).Methods(http.MethodPut)

	// Polls
	s.router.HandleFunc(fmt.Sprintf("%s/polls", v1), s.handleAddPoll()).Methods(http.MethodPost)
	s.router.HandleFunc(fmt.Sprintf("%s/polls", v1), s.handleListPolls()).Methods(http.MethodGet)
	s.router.HandleFunc(fmt.Sprintf("%s/polls/{season:[0-9]+}/{week:[0-9]+}", v1), s.handleGetPoll()).Methods(http.MethodGet).Name("poll")

	// Ballots
	s.router.HandleFunc(fmt.Sprintf("%s/ballots", v1), s.handleAddBallot()).Methods(http.MethodPost)
	s.router.HandleFunc(fmt.Sprintf("%s/ballots", v1), s.handleListBallots()).Methods(http.MethodGet)
	s.router.HandleFunc(fmt.Sprintf("%s/ballots/{id:[0-9]+}", v1), s.handleEditBallot()).Methods(http.MethodPut)
	s.router.HandleFunc(fmt.Sprintf("%s/ballots/{id:[0-9]+}", v1), s.handleGetBallot()).Methods(http.MethodGet).Name("ballot")
	s.router.HandleFunc(fmt.Sprintf("%s/ballots/{id:[0-9]+}", v1), s.handleDeleteBallot()).Methods(http.MethodDelete)
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
		return
	}
}

func (s *Server) handleAddTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.AuthClient.UserTokenFromCtx(r.Context())

		var newTeam models.Team
		err := s.decode(w, r, &newTeam)
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}

		createdTeam, err := s.App.AddTeam(token, newTeam)

		if err != nil {
			switch errors.Kind(err) {
			case errors.KindUnauthenticated:
				s.respond(w, r, nil, http.StatusUnauthorized)
				return
			case errors.KindUnauthorized:
				s.respond(w, r, nil, http.StatusForbidden)
				return
			default:
				s.respond(w, r, nil, http.StatusInternalServerError)
				return
			}
		}

		teamURL, err := s.router.Get("team").URLPath("id", fmt.Sprintf("%d", createdTeam.ID))
		if err != nil {
			log.Println("Unable to get url for created team")
		} else {
			w.Header().Set("Location", fmt.Sprintf("%s%s", s.host, teamURL.String()))
		}

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

		team, err := s.App.GetTeam(intId)

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
		teams, err := s.App.AllTeams()

		if err != nil {
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		s.respond(w, r, teams, http.StatusOK)
		return
	}
}

func (s *Server) handleAddUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.AuthClient.UserTokenFromCtx(r.Context())

		var user models.User
		err := s.decode(w, r, &user)
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}
		createdUser, err := s.App.AddUser(token, user)

		if err != nil {
			switch errors.Kind(err) {
			case errors.KindUnauthenticated:
				s.respond(w, r, nil, http.StatusUnauthorized)
				return
			case errors.KindUnauthorized:
				s.respond(w, r, nil, http.StatusForbidden)
				return
			case errors.KindConflict:
				s.respond(w, r, nil, http.StatusConflict)
				return
			default:
				s.respond(w, r, nil, http.StatusInternalServerError)
				return
			}
		}

		url, err := s.router.Get("user").URLPath("name", createdUser.Nickname)
		if err != nil {
			log.Println("Error retrieving url for created user")
		} else {
			w.Header().Set("Location", fmt.Sprintf("%s%s", s.host, url))
		}

		s.respond(w, r, createdUser, http.StatusCreated)
		return
	}
}

func (s *Server) handleUsersMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.AuthClient.UserTokenFromCtx(r.Context())

		if !token.LoggedIn() {
			s.respond(w, r, nil, http.StatusUnauthorized)
			return
		}

		user, err := s.App.GetUser(token.Nickname)
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

		user, err := s.App.GetUser(name)
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

func (s *Server) handleUpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.AuthClient.UserTokenFromCtx(r.Context())
		vars := mux.Vars(r)
		name := vars["name"]

		var user models.User
		err := s.decode(w, r, &user)
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}

		updatedUser, err := s.App.UpdateUser(token, name, user)
		if err != nil {
			switch errors.Kind(err) {
			case errors.KindUnauthenticated:
				s.respond(w, r, nil, http.StatusUnauthorized)
				return
			case errors.KindUnauthorized:
				s.respond(w, r, nil, http.StatusForbidden)
				return
			case errors.KindNotFound:
				s.respond(w, r, nil, http.StatusNotFound)
				return
			case errors.KindBadRequest:
				s.respond(w, r, nil, http.StatusBadRequest)
				return
			}
		}

		s.respond(w, r, updatedUser, http.StatusOK)
		return
	}
}

func (s *Server) handleAddPoll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.AuthClient.UserTokenFromCtx(r.Context())

		var poll models.Poll
		err := s.decode(w, r, &poll)
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}

		newPoll, err := s.App.AddPoll(token, poll)

		if err != nil {
			log.Println(err.Error())
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		url, err := s.router.Get("poll").URLPath(
			"season", strconv.FormatInt(int64(newPoll.Season), 10),
			"week", strconv.FormatInt(int64(newPoll.Week), 10))
		if err != nil {
			log.Println(fmt.Sprintf("Error retrieving url for created poll: %s", err.Error()))
		} else {
			w.Header().Set("Location", fmt.Sprintf("%s%s", s.host, url))
		}

		s.respond(w, r, newPoll, http.StatusCreated)
		return
	}
}

func (s *Server) handleListPolls() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("handleListPolls not implemented")
		return
	}
}

func (s *Server) handleGetPoll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		season, err := strconv.Atoi(vars["season"])
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
		}
		week, err := strconv.Atoi(vars["week"])
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
		}

		poll, err := s.App.GetPollByWeek(season, week)
		if err != nil {
			if errors.Kind(err) == errors.KindNotFound {
				s.respond(w, r, nil, http.StatusNotFound)
				return
			}

			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		s.respond(w, r, poll, http.StatusOK)
		return
	}
}

func (s *Server) handleAddBallot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.AuthClient.UserTokenFromCtx(r.Context())

		var ballot models.Ballot
		err := s.decode(w, r, &ballot)
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}

		newBallot, err := s.App.AddBallot(token, ballot)

		if err != nil {
			log.Println(err.Error())
			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		url, err := s.router.Get("ballot").URLPath("id", strconv.FormatInt(int64(newBallot.ID), 10))
		if err != nil {
			log.Println(fmt.Sprintf("Error retrieving url for created ballot: %s", err.Error()))
		} else {
			w.Header().Set("Location", fmt.Sprintf("%s%s", s.host, url))
		}

		s.respond(w, r, newBallot, http.StatusCreated)
		return
	}
}

func (s *Server) handleGetBallot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.AuthClient.UserTokenFromCtx(r.Context())
		vars := mux.Vars(r)
		id := vars["id"]
		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
			return
		}

		ballot, err := s.App.GetBallotById(token, intId)

		if err != nil {
			log.Println(err)
			if errors.Kind(err) == errors.KindNotFound {
				s.respond(w, r, nil, http.StatusNotFound)
				return
			}

			s.respond(w, r, nil, http.StatusInternalServerError)
			return
		}

		s.respond(w, r, ballot, http.StatusOK)
		return
	}
}

func (s *Server) handleListBallots() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		return
	}
}

func (s *Server) handleEditBallot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		return
	}
}

func (s *Server) handleDeleteBallot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.AuthClient.UserTokenFromCtx(r.Context())
		vars := mux.Vars(r)
		id := vars["id"]
		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			s.respond(w, r, nil, http.StatusBadRequest)
		}

		err = s.App.DeleteBallot(token, intId)
		if err != nil {
			switch errors.Kind(err) {
			case errors.KindUnauthenticated:
				s.respond(w, r, nil, http.StatusUnauthorized)
				return
			case errors.KindUnauthorized:
				s.respond(w, r, nil, http.StatusForbidden)
				return
			case errors.KindNotFound:
				s.respond(w, r, nil, http.StatusNotFound)
				return
			default:
				s.respond(w, r, nil, http.StatusInternalServerError)
				return
			}
		}

		s.respond(w, r, nil, http.StatusOK)
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
		user, err := s.App.GetUser(name)
		if errors.Kind(err) == errors.KindNotFound {
			user, err = s.App.NewUser(name)
			if err != nil {
				s.respond(w, r, nil, http.StatusInternalServerError)
				return
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

			url, err := s.router.Get("user").URLPath("name", user.Nickname)
			if err != nil {
				s.respond(w, r, nil, http.StatusInternalServerError)
				return
			}

			w.Header().Set("Location", fmt.Sprintf("%s%s", s.host, url))
		}
		s.respond(w, r, payload, status)
		return
	}
}
