package app

import (
	"fmt"
	"time"

	"github.com/r-cbb/cbbpoll/internal/db"
	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

var numRanks = 5

type PollService struct {
	Db     db.DBClient
	Admins []string
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
func (ps PollService) NewUser(nickname string) (models.User, error) {
	const op errors.Op = "app.NewUser"
	newUser := models.User{
		Nickname: nickname,
	}

	for _, admin := range ps.Admins {
		if nickname == admin {
			newUser.IsAdmin = true
		}
	}

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

func (ps PollService) GetUsers(user models.UserToken, opts Options) ([]models.User, error) {
	const op errors.Op = "app.GetUsers"

	users, err := ps.Db.GetUsers(opts.unpack())
	if err != nil {
		return nil, errors.E(err, op, "error retrieving users from db")
	}

	return users, nil
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
		return models.User{}, errors.E(op, "error updating user in db", err)
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

func (ps PollService) GetPolls(user models.UserToken, opts Options) ([]models.Poll, error) {
	const op errors.Op = "app.GetPolls"

	if !user.CanManagePolls() {
		opts = opts.HasOpened()
	}

	polls, err := ps.Db.GetPolls(opts.unpack())
	if err != nil {
		return nil, errors.E(op, err, "error retrieving polls from db")
	}

	return polls, nil
}

func (ps PollService) GetResults(user models.UserToken, season int, week int) ([]models.Result, error) {
	const op errors.Op = "app.GetResults"

	poll, err := ps.Db.GetPoll(season, week)
	if err != nil {
		return nil, errors.E(op, err, "error retrieving poll from db")
	}

	if poll.CloseTime.After(time.Now()) && !user.CanManagePolls() {
		return nil, errors.E(op, err, "can't view poll results until after poll close", errors.KindUnauthorized)
	}

	results, err := ps.Db.GetResults(poll, false)
	if err != nil {
		return nil, errors.E(op, err, "error retrieving results for poll")
	}

	if len(results) == 0 {
		results, err = ps.calcPollResults(poll)
		if err != nil {
			return nil, errors.E(op, err, "error calculating poll results")
		}
	}

	return results, nil
}

func (ps PollService) AddBallot(user models.UserToken, ballot models.Ballot) (models.Ballot, error) {
	const op errors.Op = "app.AddBallot"
	if !user.LoggedIn() {
		return models.Ballot{}, errors.E(op, errors.KindUnauthenticated)
	}

	u, err := ps.Db.GetUser(ballot.User)
	if err != nil {
		return models.Ballot{}, errors.E(op, err, errors.KindBadRequest, "user doesn't exist")
	}

	if u.Nickname != user.Nickname && !user.IsAdmin {
		return models.Ballot{}, errors.E(op, errors.KindUnauthorized, "can't submit ballot for another user")
	}

	ballot.IsOfficial = u.IsVoter
	ballot.UpdatedTime = time.Now()

	err = ps.validateBallot(ballot)
	if err != nil {
		return models.Ballot{}, errors.E(op, err, "ballot failed validation", errors.KindBadRequest)
	}

	newBallot, err := ps.Db.AddBallot(ballot)
	if err != nil {
		return models.Ballot{}, errors.E(op, err, "error adding ballot to DB")
	}

	return newBallot, nil
}

func (ps PollService) validateBallot(b models.Ballot) error {
	vs := b.Votes

	if len(vs) != numRanks {
		return errors.E(fmt.Errorf("ballots must contain exactly %v votes", numRanks))
	}

	if containsDuplicates(vs) {
		return errors.E(fmt.Errorf("ballot contains duplicate votes"))
	}

	if err := checkVotes(vs, ps.Db); err != nil {
		return err
	}

	return nil
}

func containsDuplicates(vs []models.Vote) bool {
	seen := make(map[int64]struct{})
	for _, v := range vs {
		_, ok := seen[v.TeamID]
		if ok {
			return true
		}
		seen[v.TeamID] = struct{}{}
	}
	return false
}

func checkVotes(vs []models.Vote, db db.DBClient) error {
	teamIDs := make([]int64, len(vs))
	for i, v := range vs {
		teamIDs[i] = v.TeamID
		if v.TeamID == 0 || v.Rank == 0 {
			return fmt.Errorf("no votes can have a team_id or rank of 0")
		}

		if len(v.Reason) > 140 {
			return fmt.Errorf("reasons can't be longer than 140 characters")
		}
	}

	_, err := db.GetTeamsByID(teamIDs)
	if err != nil {
		return fmt.Errorf("unable to retrieve ballot's teams from db")
	}

	return nil
}

func (ps PollService) DeleteBallot(user models.UserToken, id int64) error {
	const op errors.Op = "app.DeleteBallot"
	if !user.LoggedIn() {
		return errors.E(op, errors.KindUnauthenticated)
	}

	ballot, err := ps.Db.GetBallot(id)
	if err != nil {
		return errors.E(op, "error getting ballot", err)
	}

	if ballot.User != user.Nickname && !user.IsAdmin {
		return errors.E(op, errors.KindUnauthorized, "can't delete someone else's ballot")
	}

	poll, err := ps.Db.GetPoll(ballot.PollSeason, ballot.PollWeek)
	if err != nil {
		return errors.E(op, "error getting poll for ballot")
	}

	if poll.CloseTime.Before(time.Now()) && !user.IsAdmin {
		return errors.E(op, errors.KindBadRequest, "can't delete a ballot for a closed poll")
	}

	err = ps.Db.DeleteBallot(id)
	if err != nil {
		return errors.E(op, err, "error deleting ballot")
	}

	return nil
}

func (ps PollService) GetBallotById(user models.UserToken, id int64) (models.Ballot, error) {
	const op errors.Op = "app.GetBallotById"

	ballot, err := ps.Db.GetBallot(id)
	if err != nil {
		return models.Ballot{}, errors.E(op, err, "error retrieving ballot from DB")
	}

	poll, err := ps.Db.GetPoll(ballot.PollSeason, ballot.PollWeek)
	if err != nil {
		return models.Ballot{}, errors.E(op, err, "error retrieving poll for ballot")
	}

	if poll.CloseTime.After(time.Now()) && !user.IsAdmin && ballot.User != user.Nickname {
		return models.Ballot{}, errors.E(op, err, "users can't see other's ballots until the poll closes", errors.KindUnauthorized)
	}

	return ballot, nil
}
