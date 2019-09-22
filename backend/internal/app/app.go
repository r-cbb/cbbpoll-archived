package app

import (
	"fmt"
	"log"
	"sort"

	"github.com/r-cbb/cbbpoll/internal/db"
	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

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
	var users []models.User

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

	_, err := ps.Db.GetPollByWeek(poll.Season, poll.Week)
	if errors.Kind(err) != errors.KindNotFound {
		return models.Poll{}, errors.E(op, errors.KindConflict, fmt.Sprintf("poll already exists for season %v week %v", poll.Season, poll.Week))
	}

	newPoll, err := ps.Db.AddPoll(poll)
	if err != nil {
		return models.Poll{}, errors.E(op, "error adding poll to db", err)
	}

	return newPoll, nil
}

func (ps PollService) GetPollByWeek(season int, week int) (models.Poll, error) {
	const op errors.Op = "app.GetPollByWeek"
	poll, err := ps.Db.GetPollByWeek(season, week)
	if err != nil {
		return models.Poll{}, errors.E(op, err, "error retrieving poll from db")
	}

	if poll.Results != nil {
		return poll, nil
	}

	poll, err = ps.calcPollResults(poll)
	if err != nil {
		return models.Poll{}, errors.E(op, err, "error calculating poll results")
	}

	return poll, nil
}

type resultsSlice []models.Result

func (rs resultsSlice) Len() int {
	return len(rs)
}

func (rs resultsSlice) Less(i, j int) bool {
	return rs[i].Points > rs[j].Points
}

func (rs resultsSlice) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (ps PollService) GetPoll(id int64) (models.Poll, error) {
	const op errors.Op = "app.GetPoll"
	poll, err := ps.Db.GetPoll(id)
	if err != nil {
		return models.Poll{}, errors.E(op, err, "error retrieving poll from db")
	}

	if poll.Results != nil {
		return poll, nil
	}

	poll, err = ps.calcPollResults(poll)
	if err != nil {
		return models.Poll{}, errors.E(op, err, "error calculating poll results")
	}

	return poll, nil
}

func (ps PollService) calcPollResults(poll models.Poll) (models.Poll, error) {
	const op errors.Op = "app.calcPollResults"
	log.Println("Calculating results")
	// calculate and store results
	resMap := make(map[int64]models.Result)

	var ballotIds []int64
	for _, ballot := range poll.Ballots {
		ballotIds = append(ballotIds, ballot.ID)
	}

	ballots, err := ps.Db.GetBallotsByID(ballotIds)
	if err != nil {
		return models.Poll{}, errors.E(op, err, "error retrieving ballots associated with poll")
	}

	for _, ballot := range ballots {
		for _, vote := range ballot.Votes {
			res := resMap[vote.TeamID]
			if vote.Rank == 1 {
				res.FirstPlaceVotes = res.FirstPlaceVotes + 1
			}
			res.Points = res.Points + 26 - vote.Rank
			resMap[vote.TeamID] = res
		}
	}

	results := make(resultsSlice, 0, 25)
	for k, v := range resMap {
		v.TeamID = k
		results = append(results, v)
	}

	sort.Sort(results)
	poll.Results = results

	err = ps.Db.UpdatePoll(poll)
	if err != nil {
		return models.Poll{}, errors.E(op, err, "error updating poll after calculating results")
	}

	return poll, nil
}

func (ps PollService) AddBallot(user models.UserToken, ballot models.Ballot) (models.Ballot, error) {
	const op errors.Op = "app.AddBallot"
	if !user.LoggedIn() {
		return models.Ballot{}, errors.E(op, errors.KindUnauthenticated)
	}

	// Todo handle voter status

	// todo Validate ballot

	newBallot, err := ps.Db.AddBallot(ballot)
	if err != nil {
		return models.Ballot{}, errors.E(op, err, "error adding ballot to DB")
	}

	return newBallot, nil
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

	err = ps.Db.DeleteBallot(id)
	if err != nil {
		return errors.E(op, err, "error deleting ballot")
	}

	return nil
}

func (ps PollService) GetBallotById(user models.UserToken, id int64) (models.Ballot, error) {
	const op errors.Op = "app.GetBallotById"

	// todo handle permission to view this ballot

	ballot, err := ps.Db.GetBallot(id)
	if err != nil {
		return models.Ballot{}, errors.E(op, err, "error retrieving ballot from DB")
	}

	return ballot, nil
}
