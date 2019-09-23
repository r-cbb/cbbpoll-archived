package db

import "github.com/r-cbb/cbbpoll/internal/models"

type DBClient interface {
	AddTeam(newTeam models.Team) (team models.Team, err error)
	GetTeam(id int64) (team models.Team, err error)
	GetTeams() (teams []models.Team, err error)

	AddUser(newUser models.User) (user models.User, err error)
	UpdateUser(user models.User) (err error)
	GetUser(name string) (user models.User, err error)
	GetUsers(filter []Filter, sort Sort) ([]models.User, error)

	AddPoll(newPoll models.Poll) (poll models.Poll, err error)
	UpdatePoll(poll models.Poll, results *[]models.Result) error
	GetPoll(id int64) (poll models.Poll, err error)
	GetPollByWeek(season int, week int) (poll models.Poll, err error)

	AddBallot(newBallot models.Ballot) (ballot models.Ballot, err error)
	GetBallot(id int64) (ballot models.Ballot, err error)
	GetBallotsByID(ids []int64) (ballots []models.Ballot, err error)
	GetBallotsByPoll(poll models.Poll) (ballots []models.Ballot, err error)
	DeleteBallot(id int64) (err error)
	UpdateBallot(ballot models.Ballot) error
}

type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

type Sort struct {
	field string
	asc   bool
}