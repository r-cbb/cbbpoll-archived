package db

import "github.com/r-cbb/cbbpoll/internal/models"

type DBClient interface {
	AddTeam(newTeam models.Team) (team models.Team, err error)
	GetTeam(id int64) (team models.Team, err error)
	GetTeams() (teams []models.Team, err error)

	AddUser(newUser models.User) (user models.User, err error)
	UpdateUser(user models.User) (updated models.User, err error)
	GetUser(name string) (user models.User, err error)

	AddPoll(newPoll models.Poll) (poll models.Poll, err error)
	GetPoll(season int, week int) (poll models.Poll, err error)

	AddBallot(newBallot models.Ballot) (ballot models.Ballot, err error)
	GetBallot()
}
