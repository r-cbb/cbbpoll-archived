package app

import (
	"fmt"

	"github.com/r-cbb/cbbpoll/internal/db"
	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

type PollService struct {
	Db db.DBClient
}

func NewPollService(Db db.DBClient) *PollService {
	ps := PollService{Db: Db}
	return &ps
}

func (ps PollService) AddTeam(user models.UserToken, newTeam models.Team) (createdTeam models.Team, err error) {
	const op errors.Op = "app.AddTeam"
	if !user.LoggedIn() {
		return models.Team{}, errors.E(op, "user is not logged in", errors.KindUnauthenticated)
	}

	if !user.IsAdmin {
		return models.Team{}, errors.E(op, "only admins can add teams", errors.KindUnauthorized)
	}

	createdTeam, err = ps.Db.AddTeam(newTeam)
	if errors.Kind(err) == errors.KindConcurrencyProblem {
		// Retry once
		createdTeam, err = ps.Db.AddTeam(newTeam)
	}

	if err != nil {
		return models.Team{}, errors.E(op, err, "error adding team to db")
	}

	return createdTeam, nil
}

func (ps PollService) GetTeam(id int64) (team models.Team, err error) {
	const op errors.Op = "app.GetTeam"
	team, err = ps.Db.GetTeam(id)
	if err != nil {
		return models.Team{}, errors.E(err, op, "error retrieving team from db")
	}

	return team, nil
}

func (ps PollService) AllTeams() (teams []models.Team, err error) {
	const op errors.Op = "app.AllTeams"
	teams, err = ps.Db.GetTeams()
	if err != nil {
		return nil, errors.E(err, op, "error retrieving teams from db")
	}

	return teams, nil
}

// NewUser is only to be used when a user logs in who does not have a user record
// in the database.  This will create the user with base permissions.
func (ps PollService) NewUser(newUser models.User) (models.User, error) {
	const op errors.Op = "app.NewUser"
	createdUser, err := ps.Db.AddUser(newUser)
	if err != nil {
		return models.User{}, errors.E(op, err, "error adding user to db")
	}

	return createdUser, nil
}

func (ps PollService) AddUser(user models.UserToken, newUser models.User) (createdUser models.User, err error) {
	const op errors.Op = "app.AddUser"
	if !user.LoggedIn() {
		return models.User{}, errors.E(op, errors.KindUnauthenticated)
	}

	if !user.IsAdmin {
		return models.User{}, errors.E(op, errors.KindUnauthorized)
	}

	createdUser, err = ps.Db.AddUser(newUser)
	if err != nil {
		return models.User{}, errors.E(op, err, "error adding user to db")
	}

	return createdUser, nil
}

func (ps PollService) GetUser(name string) (models.User, error) {
	const op errors.Op = "app.GetUser"
	user, err := ps.Db.GetUser(name)
	if err != nil {
		return models.User{}, errors.E(op, err, "error retrieving user from db")
	}

	return user, nil
}

func (ps PollService) UpdateUser(user models.UserToken, name string, updatedUser models.User) (models.User, error) {
	const op errors.Op = "app.UpdateUser"
	if !user.LoggedIn() {
		return models.User{}, errors.E(op, errors.KindUnauthenticated)
	}

	if user.Nickname != name && !user.IsAdmin {
		return models.User{}, errors.E(op, errors.KindUnauthorized)
	}

	if updatedUser.Nickname != name {
		return models.User{}, errors.E(op, errors.KindBadRequest, "can't change a user's nickname")
	}

	existingUser, err := ps.Db.GetUser(name)
	if err != nil {
		return models.User{}, errors.E(op, "error retrieving user to update from db")
	}

	if existingUser.IsVoter != updatedUser.IsVoter && !user.IsAdmin {
		return models.User{}, errors.E(op, errors.KindUnauthorized, "only admins can alter voter status")
	}

	if existingUser.IsAdmin != updatedUser.IsAdmin && !user.IsAdmin {
		return models.User{}, errors.E(op, errors.KindUnauthorized, "only admins can change a user's admin status")
	}

	err = ps.Db.UpdateUser(updatedUser)
	if err != nil {
		return models.User{}, errors.E(op, "error updating user in db")
	}

	return updatedUser, nil
}

func (ps PollService) AddPoll(user models.UserToken, poll models.Poll) (models.Poll, error) {
	const op errors.Op = "app.AddPoll"
	if !user.LoggedIn() {
		return models.Poll{}, errors.E(op, errors.KindUnauthenticated)
	}

	if !user.CanManagePolls() {
		return models.Poll{}, errors.E(op, errors.KindUnauthorized, "user doesn't have sufficient permissions to add a poll")
	}

	_, err := ps.Db.GetPoll(poll.Season, poll.Week)
	if errors.Kind(err) != errors.KindNotFound {
		return models.Poll{}, errors.E(op, errors.KindConflict, fmt.Sprintf("poll already exists for season %v week %v", poll.Season, poll.Week))
	}

	newPoll, err := ps.Db.AddPoll(poll)
	if err != nil {
		return models.Poll{}, errors.E(op, "error adding poll to db", err)
	}

	return newPoll, nil
}

func (ps PollService) GetPoll(season int, week int) (models.Poll, error) {
	const op errors.Op = "app.GetPoll"
	poll, err := ps.Db.GetPoll(season, week)
	if err != nil {
		return models.Poll{}, errors.E(op, err, "error retrieving poll from db")
	}

	return poll, nil
}
