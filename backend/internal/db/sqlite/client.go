package sqlite

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

type Client struct {
	db *sqlx.DB
}

func NewClient(filename string) (*Client, error) {
	const op errors.Op = "sqlite.NewClient"

	db, err := sqlx.Open("sqlite3", "file:cbbpoll.db")
	if err != nil {
		return nil, errors.E("could not open sqlite db", err, op, errors.KindDatabaseError)
	}

	return &Client{db: db}, nil
}

func (c *Client) AddTeam(newTeam models.Team) (team models.Team, err error) {
	const op errors.Op = "sqlite.AddTeam"
	var t Team
	t.fromContract(newTeam)

	res, err := c.db.Exec("INSERT INTO team (full_name, short_name, nickname, conference) VALUES ($1, $2, $3, $4)",
		t.FullName, t.ShortName, t.Nickname, t.Conference)

	if err != nil {
		return models.Team{}, errors.E(op, err, "error adding team to db", errors.KindDatabaseError)
	}

	t.ID, err = res.LastInsertId()
	if err != nil {
		return models.Team{}, errors.E(op, err, "error getting id for new team", errors.KindDatabaseError)
	}

	return t.toContract(), nil
}

func (c *Client) GetTeam(id int64) (team models.Team, err error) {
	const op errors.Op = "sqlite.GetTeam"
	var t Team
	err = c.db.Get(&t, "SELECT * FROM team WHERE id = ?", id)

	if err != nil {
		return models.Team{}, errors.E(op, err, "error retrieving team from db", errors.KindDatabaseError)
	}

	return t.toContract(), nil
}

func (c *Client) GetTeams() (teams []models.Team, err error) {
	const op errors.Op = "sqlite.GetTeams"
	var ts []Team
	err = c.db.Select(&ts, "SELECT * from team;")

	if err != nil {
		return nil, errors.E(op, err, "error retrieving teams from db", errors.KindDatabaseError)
	}

	cs := make([]models.Team, len(ts))
	for i := range ts {
		cs[i] = ts[i].toContract()
	}

	return cs, nil
}

func (c *Client) AddUser(newUser models.User) (models.User, error) {
	const op errors.Op = "sqlite.AddUser"
	var u User
	u.fromContract(newUser)

	tx, err := c.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return models.User{}, errors.E(op, err, "error creating transaction", errors.KindDatabaseError)
	}

	var tmp User
	err = tx.Get(&tmp, "SELECT * FROM user WHERE nickname = ?", u.Nickname)
	if err != sql.ErrNoRows {
		_ = tx.Rollback()
		return models.User{}, errors.E(op, err, "user already exists", errors.KindConflict)
	}

	_, err = tx.Exec("INSERT INTO user (nickname, is_admin, is_voter) VALUES ($1, $2, $3)", u.Nickname, u.IsAdmin, u.IsVoter)
	if err != nil {
		_ = tx.Rollback()
		return models.User{}, errors.E(op, err, "error adding user to db", errors.KindDatabaseError)
	}

	err = tx.Commit()
	if err != nil {
		return models.User{}, errors.E(op, err, "error committing transaction", errors.KindDatabaseError)
	}

	return u.toContract(), nil
}

func (c *Client) UpdateUser(user models.User) error {
	const op errors.Op = "sqlite.UpdateUser"
	var u User
	u.fromContract(user)

	_, err := c.db.Exec("UPDATE user SET is_admin = $1, is_voter = $2 WHERE nickname = $3", u.IsAdmin, u.IsVoter, u.Nickname)
	if err != nil {
		return errors.E(op, err, "err updating user", errors.KindDatabaseError)
	}

	return nil
}

func (c *Client) GetUser(name string) (models.User, error) {
	const op errors.Op = "sqlite.GetUser"
	var u User

	err := c.db.Get(&u, "SELECT * FROM user WHERE nickname = ?", u.Nickname)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.E(op, err, "user doesn't exist", errors.KindNotFound)
		}
		return models.User{}, errors.E(op, err, "error retrieving user", errors.KindDatabaseError)
	}

	return u.toContract(), nil
}

func (*Client) AddPoll(newPoll models.Poll) (poll models.Poll, err error) {
	panic("implement me")
}

func (*Client) GetPoll(season int, week int) (poll models.Poll, err error) {
	panic("implement me")
}

func (*Client) AddBallot(newBallot models.Ballot) (ballot models.Ballot, err error) {
	panic("implement me")
}

func (*Client) GetBallot(id int64) (ballot models.Ballot, err error) {
	panic("implement me")
}
