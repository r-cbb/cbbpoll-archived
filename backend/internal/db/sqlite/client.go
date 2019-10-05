package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/r-cbb/cbbpoll/internal/db"
	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

type Client struct {
	db *sqlx.DB
}

func NewClient(filename string) (*Client, error) {
	const op errors.Op = "sqlite.NewClient"

	sqliteDb, err := sqlx.Open("sqlite3", fmt.Sprintf("file:%s?_fk=true", filename))
	if err != nil {
		return nil, errors.E("could not open sqlite db", err, op, errors.KindDatabaseError)
	}

	return &Client{db: sqliteDb}, nil
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
	if err == nil {
		_ = tx.Rollback()
		return models.User{}, errors.E(op, err, "user already exists", errors.KindConflict)
	} else if err != sql.ErrNoRows {
		_ = tx.Rollback()
		return models.User{}, errors.E(op, err, "error checking for existing user", errors.KindDatabaseError)
	}

	_, err = tx.Exec("INSERT INTO user (nickname, is_admin, is_voter, primary_team) VALUES ($1, $2, $3, $4)", u.Nickname, u.IsAdmin, u.IsVoter, u.PrimaryTeam)
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

	_, err := c.db.Exec("UPDATE user SET is_admin = $1, is_voter = $2, primary_team = $3 WHERE nickname = ?", u.IsAdmin, u.IsVoter, u.PrimaryTeam, u.Nickname)
	if err != nil {
		return errors.E(op, err, "err updating user", errors.KindDatabaseError)
	}

	return nil
}

func (c *Client) GetUser(name string) (models.User, error) {
	const op errors.Op = "sqlite.GetUser"
	var u User

	err := c.db.Get(&u, "SELECT * FROM user WHERE nickname = ?", name)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.E(op, err, "user doesn't exist", errors.KindNotFound)
		}
		return models.User{}, errors.E(op, err, "error retrieving user", errors.KindDatabaseError)
	}

	return u.toContract(), nil
}

func (c *Client) AddPoll(newPoll models.Poll) (models.Poll, error) {
	const op errors.Op = "sqlite.AddPoll"
	var p Poll
	p.fromContract(newPoll)

	tx, err := c.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return models.Poll{}, errors.E(op, err, "error creating transaction", errors.KindDatabaseError)
	}

	var tmp Poll
	err = tx.Get(&tmp, "SELECT * FROM poll WHERE season = ? AND week = ?", p.Season, p.Week)
	if err == nil {
		_ = tx.Rollback()
		return models.Poll{}, errors.E(op, err, "poll already exists for week", errors.KindConflict)
	} else if err != sql.ErrNoRows {
		_ = tx.Rollback()
		return models.Poll{}, errors.E(op, err, "error checking for existing poll", errors.KindDatabaseError)
	}

	_, err = tx.Exec("INSERT INTO poll (season, week, week_name, open_time, close_time, last_modified, reddit_url) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		p.Season, p.Week, p.WeekName, p.OpenTime, p.CloseTime, p.LastModified, p.RedditURL)
	if err != nil {
		_ = tx.Rollback()
		return models.Poll{}, errors.E(op, err, "error adding poll to db", errors.KindDatabaseError)
	}

	err = tx.Commit()
	if err != nil {
		return models.Poll{}, errors.E(op, err, "error committing transaction", errors.KindDatabaseError)
	}

	return p.toContract(), nil
}

func (c *Client) GetPoll(season int, week int) (models.Poll, error) {
	const op errors.Op = "sqlite.GetPoll"
	var p Poll

	err := c.db.Get(&p, "SELECT * FROM poll WHERE season = ? and week = ?", season, week)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Poll{}, errors.E(op, err, "no poll found for week", errors.KindNotFound)
		}
		return models.Poll{}, errors.E(op, err, "error retrieving poll", errors.KindDatabaseError)
	}

	return p.toContract(), nil
}

func addBallotAndVotes(tx *sqlx.Tx, b Ballot, vs []Vote) (Ballot, error) {
	const op errors.Op = "sqlite.addBallotAndVotes"
	var tmp Ballot
	err := tx.Get(&tmp, "SELECT * FROM ballot WHERE user = ? AND poll_season = ? AND poll_week = ?", b.User, b.PollSeason, b.PollWeek)
	if err == nil {
		return Ballot{}, errors.E(op, err, "ballot already exists for user", errors.KindConflict)
	} else if err != sql.ErrNoRows {
		return Ballot{}, errors.E(op, err, "error checking for existing ballot", errors.KindDatabaseError)
	}

	// If ID is provided, accept it, otherwise let DB generate it.
	var query string
	var args []interface{}
	if b.ID != 0 {
		query = "INSERT INTO ballot (id, poll_season, poll_week, updated_time, user, is_official) VALUES ($1, $2, $3, $4, $5, $6)"
		args = []interface{}{b.ID, b.PollSeason, b.PollWeek, b.UpdatedTime, b.User, b.IsOfficial}
	} else {
		query = "INSERT INTO ballot (poll_season, poll_week, updated_time, user, is_official) VALUES ($1, $2, $3, $4, $5)"
		args = []interface{}{b.PollSeason, b.PollWeek, b.UpdatedTime, b.User, b.IsOfficial}
	}

	res, err := tx.Exec(query, args...)
	if err != nil {
		return Ballot{}, errors.E(op, err, "error adding ballot to db", errors.KindDatabaseError)
	}

	b.ID, err = res.LastInsertId()
	if err != nil {
		return Ballot{}, errors.E(op, err, "error getting id for created ballot", errors.KindDatabaseError)
	}

	for _, v := range vs {
		_, err = tx.Exec("INSERT INTO vote (ballot_id, team_id, rank, reason) VALUES ($1, $2, $3, $4)", b.ID, v.TeamID, v.Rank, v.Reason)
		if err != nil {
			return Ballot{}, errors.E(op, err, "error adding votes to db", errors.KindDatabaseError)
		}
	}

	return b, nil
}

func (c *Client) AddBallot(newBallot models.Ballot) (models.Ballot, error) {
	const op errors.Op = "sqlite.AddBallot"
	var b Ballot
	vs := b.fromContract(newBallot)

	tx, err := c.db.Beginx()
	if err != nil {
		return models.Ballot{}, errors.E(op, err, "error creating transaction", errors.KindDatabaseError)
	}

	b, err = addBallotAndVotes(tx, b, vs)
	if err != nil {
		_ = tx.Rollback()
		return models.Ballot{}, errors.E(op, err, "error during ballot creation", errors.KindDatabaseError)
	}

	err = invalidateResults(tx, b.PollSeason, b.PollWeek)
	if err != nil {
		_ = tx.Rollback()
		return models.Ballot{}, errors.E(op, err, "error invalidating poll results")
	}

	err = tx.Commit()
	if err != nil {
		return models.Ballot{}, errors.E(op, err, "error committing transaction", errors.KindDatabaseError)
	}

	cb, err := b.toContract(c)
	if err != nil {
		log.Println("error converting ballot to contract")
	}

	return cb, nil
}

func (c *Client) GetBallot(id int64) (models.Ballot, error) {
	const op errors.Op = "sqlite.GetBallot"
	var b Ballot

	err := c.db.Get(&b, "SELECT * FROM ballot WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Ballot{}, errors.E(op, err, "ballot not found", errors.KindNotFound)
		}
		return models.Ballot{}, errors.E(op, err, "error retrieving ballot", errors.KindDatabaseError)
	}

	cb, err := b.toContract(c)
	if err != nil {
		return models.Ballot{}, errors.E(op, err, "error converting ballot to contract", errors.KindDatabaseError)
	}

	return cb, nil
}

func (c *Client) getVotes(ballotID int64) ([]Vote, error) {
	const op errors.Op = "sqlite.getVotes"
	var vs []Vote

	err := c.db.Select(&vs, "SELECT * FROM vote WHERE ballot_id = ?", ballotID)
	if err != nil {
		return nil, errors.E(op, err, "error retrieving votes for ballot", errors.KindDatabaseError)
	}

	return vs, nil
}

func (c *Client) GetTeamsByID(ids []int64) ([]models.Team, error) {
	const op errors.Op = "sqlite.GetTeamsByID"
	ts := make([]Team, len(ids))
	cts := make([]models.Team, len(ids))

	for i := range ids {
		err := c.db.Get(&ts[i], "SELECT * FROM team WHERE id = ?", ids[i])
		if err != nil {
			return nil, errors.E(op, err, "error retrieving team from db", errors.KindDatabaseError)
		}
		cts[i] = ts[i].toContract()
	}

	return cts, nil
}

func (c *Client) GetUsers(filter []db.Filter, sort db.Sort) ([]models.User, error) {
	const op errors.Op = "sqlite.GetUsers"
	var us []User

	query := "SELECT * FROM user"
	var args []interface{}

	for i, f := range filter {
		if i == 0 {
			query += fmt.Sprintf(" WHERE %s %s ?", f.Field, f.Operator)
		} else {
			query += fmt.Sprintf(" AND %s %s ?", f.Field, f.Operator)
		}
		args = append(args, f.Value)
	}

	err := c.db.Select(&us, query, args...)
	if err != nil {
		return nil, errors.E(op, err, "error retrieving users", errors.KindDatabaseError)
	}

	cus := make([]models.User, len(us))
	for i := range us {
		cus[i] = us[i].toContract()
	}

	return cus, nil
}

func (c *Client) UpdatePoll(poll models.Poll) error {
	const op errors.Op = "sqlite.UpdatePoll"

	var p Poll
	p.fromContract(poll)
	p.LastModified = time.Now()

	res, err := c.db.Exec("UPDATE poll SET week_name = $1, open_time = $2, close_time = $3, last_modified = $4, reddit_url = $5 WHERE season = $6 AND week = $7",
		p.WeekName, p.OpenTime, p.CloseTime, p.LastModified, p.RedditURL, p.Season, p.Week)

	if err != nil {
		return errors.E(op, err, "error updating poll", errors.KindDatabaseError)
	}

	if aff, err := res.RowsAffected(); aff != 0 || err != nil {
		return errors.E(op, err, "poll not found to update", errors.KindNotFound)
	}

	return nil
}

func (c *Client) GetPolls(filter []db.Filter, sort db.Sort) ([]models.Poll, error) {
	const op errors.Op = "sqlite.GetPolls"
	var ps []Poll

	query := "SELECT * FROM poll"
	var args []interface{}

	for i, f := range filter {
		if i == 0 {
			query += fmt.Sprintf(" WHERE %s %s ?", f.Field, f.Operator)
		} else {
			query += fmt.Sprintf(" AND %s %s ?", f.Field, f.Operator)
		}
		args = append(args, f.Value)
	}

	err := c.db.Select(&ps, query, args...)
	if err != nil {
		return nil, errors.E(op, err, "error retrieving polls", errors.KindDatabaseError)
	}

	cps := make([]models.Poll, len(ps))
	for i := range ps {
		cps[i] = ps[i].toContract()
	}

	return cps, nil
}

func invalidateResults(tx *sqlx.Tx, season int, week int) error {
	const op errors.Op = "sqlite.invalidateResults"

	_, err := tx.Exec("DELETE FROM result WHERE poll_season = ? and poll_week = ?", season, week)
	if err != nil {
		return errors.E(op, err, "error deleting result rows", errors.KindDatabaseError)
	}

	return nil
}

func (c *Client) SetResults(poll models.Poll, official []models.Result, allBallots []models.Result) error {
	const op errors.Op = "sqlite.SetResults"

	tx, err := c.db.Beginx()
	if err != nil {
		return errors.E(op, err, "error creating transaction", errors.KindDatabaseError)
	}

	var p Poll
	err = tx.Get(&p, "SELECT * FROM poll WHERE season = ? AND week = ?", poll.Season, poll.Week)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, err, "poll not found", errors.KindNotFound)
	}

	if p.LastModified != poll.LastModified {
		_ = tx.Rollback()
		return errors.E(op, err, "results set out of date with poll in db", errors.KindConcurrencyProblem)
	}

	for _, res := range official {
		var r Result
		r.fromContract(res, poll, true)
		_, err := tx.Exec("INSERT INTO result (poll_season, poll_week, team_id, team_name, team_slug, rank, first_place_votes, points, official) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
			r.Season, r.Week, r.TeamID, r.TeamName, r.TeamSlug, r.Rank, r.FirstPlaceVotes, r.Points, r.Official)
		if err != nil {
			_ = tx.Rollback()
			return errors.E(op, err, "error inserting result into db", errors.KindDatabaseError)
		}
	}

	for _, res := range allBallots {
		var r Result
		r.fromContract(res, poll, false)
		_, err := tx.Exec("INSERT INTO result (poll_season, poll_week, team_id, team_name, team_slug, rank, first_place_votes, points, official) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
			r.Season, r.Week, r.TeamID, r.TeamName, r.TeamSlug, r.Rank, r.FirstPlaceVotes, r.Points, r.Official)
		if err != nil {
			_ = tx.Rollback()
			return errors.E(op, err, "error inserting result into db", errors.KindDatabaseError)
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.E(op, err, "error committing transaction", errors.KindDatabaseError)
	}

	return nil
}

func (c *Client) GetResults(poll models.Poll, includeProvisional bool) ([]models.Result, error) {
	const op errors.Op = "sqlite.GetResults"
	var rs []Result

	err := c.db.Select(&rs, "SELECT * FROM result WHERE poll_season = ? AND poll_week = ?", poll.Season, poll.Week)
	if err != nil {
		return nil, errors.E(op, err, "error retrieving rows from db", errors.KindDatabaseError)
	}

	crs := make([]models.Result, len(rs))
	for i := range rs {
		crs[i] = rs[i].toContract()
	}

	return crs, nil
}

func (c *Client) GetBallotsByPoll(poll models.Poll) ([]models.Ballot, error) {
	const op errors.Op = "sqlite.GetBallotsByPoll"
	var bs []Ballot

	err := c.db.Select(&bs, "SELECT * FROM ballot WHERE poll_season = ? AND poll_week = ?", poll.Season, poll.Week)
	if err != nil {
		return nil, errors.E(op, err, "error retrieving ballots for poll", errors.KindDatabaseError)
	}

	cbs := make([]models.Ballot, len(bs))
	for i := range bs {
		cb, err := bs[i].toContract(c)
		if err != nil {
			return nil, errors.E(op, err, "error converting ballots to contracts", errors.KindDatabaseError)
		}
		cbs[i] = cb
	}

	return cbs, nil
}

func deleteBallotAndVotes(tx *sqlx.Tx, id int64) error {
	const op errors.Op = "sqlite.deleteBallotAndVotes"

	_, err := tx.Exec("DELETE FROM vote WHERE ballot_id = ?", id)
	if err != nil {
		return errors.E(op, err, "error deleting votes", errors.KindDatabaseError)
	}

	res, err := tx.Exec("DELETE FROM ballot WHERE id = ?", id)
	if err != nil {
		return errors.E(op, err, "error deleting ballot", errors.KindDatabaseError)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return errors.E(op, err, "error getting affected rows", errors.KindDatabaseError)
	}

	if aff == 0 {
		return errors.E(op, err, "no ballot found with given id", errors.KindNotFound)
	}

	return nil
}

func (c *Client) DeleteBallot(id int64) error {
	const op errors.Op = "sqlite.DeleteBallot"

	tx, err := c.db.Beginx()
	if err != nil {
		return errors.E(op, err, "error creating transaction", errors.KindDatabaseError)
	}

	// Need to grab the ballot so we know which poll to invalidate results for
	var b Ballot
	err = tx.Get(&b, "SELECT * FROM ballot WHERE id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			_ = tx.Rollback()
			return errors.E(op, err, "ballot not found", errors.KindNotFound)
		}
		_ = tx.Rollback()
		return errors.E(op, err, "error retrieving ballot", errors.KindDatabaseError)
	}

	err = invalidateResults(tx, b.PollSeason, b.PollWeek)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, err, "error invalidating poll results")
	}

	err = deleteBallotAndVotes(tx, id)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, err, "error during ballot deletion")
	}

	err = tx.Commit()
	if err != nil {
		return errors.E(op, err, "error committing transaction", errors.KindDatabaseError)
	}
	return nil
}

func (c *Client) UpdateBallot(ballot models.Ballot) error {
	const op errors.Op = "sqlite.UpdateBallot"
	var b Ballot
	vs := b.fromContract(ballot)
	tx, err := c.db.Beginx()
	if err != nil {
		return errors.E(op, err, "error creating transaction", errors.KindDatabaseError)
	}

	err = deleteBallotAndVotes(tx, ballot.ID)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, err, "error removing old ballot from db")
	}

	b, err = addBallotAndVotes(tx, b, vs)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, err, "error adding updated ballot to db, rolling back")
	}

	err = invalidateResults(tx, b.PollSeason, b.PollWeek)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, err, "error invalidating poll results")
	}

	err = tx.Commit()
	if err != nil {
		return errors.E(op, err, "error committing transaction", errors.KindDatabaseError)
	}

	return nil
}
