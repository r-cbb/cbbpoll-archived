package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"

	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

type DatastoreClient struct {
	client *datastore.Client
}

// idStruct is a type used to load arbitrary entities out of the Datastore,
// as long as they have an ID field.  The application-level ID is a concession
// to backwards compatibility with the old implementation where objects
// used mysql auto-incrementing primary keys as IDs.  These IDs are used in
// several URLs, so we need to carry them forward.
type idStruct struct {
	ID int64
}

func (i *idStruct) Load(property []datastore.Property) error {
	var ok, foundId bool
	for _, v := range property {
		if v.Name == "ID" {
			i.ID, ok = v.Value.(int64)
			if !ok {
				return fmt.Errorf("error loading ID property")
			}
			foundId = true
		}
	}
	if !foundId {
		return fmt.Errorf("no ID property on load")
	}
	return nil
}

func (i idStruct) Save() ([]datastore.Property, error) {
	return nil, fmt.Errorf("Should never save an idStruct to storage")
}

func NewDatastoreClient(projectId string) (*DatastoreClient, error) {
	const op errors.Op = "datastore.NewDatastoreClient"
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, projectId)
	if err != nil {
		return nil, errors.E("could not connect to datastore", err, op, errors.KindDatabaseError)
	}

	// Verify that we can communicate and authenticate with the datastore service.
	t, err := client.NewTransaction(ctx)
	if err != nil {
		return nil, errors.E("problem opening test transaction", err, op, errors.KindDatabaseError)
	}
	if err := t.Rollback(); err != nil {
		return nil, errors.E("problem rolling back test transaction", err, op, errors.KindDatabaseError)
	}

	return &DatastoreClient{client: client}, nil
}

func (db *DatastoreClient) nextID(kind string) (id int64, err error) {
	ctx := context.Background()
	q := datastore.NewQuery(kind).Order("-ID")
	var ids []idStruct

	_, err = db.client.GetAll(ctx, q, &ids)
	if err != nil {
		return
	}

	if len(ids) == 0 {
		id = 1
	} else {
		id = ids[0].ID + 1
	}

	return
}

func (db *DatastoreClient) AddTeam(team models.Team) (models.Team, error) {
	const op errors.Op = "datastore.GetTeam"
	ctx := context.Background()

	newId, err := db.nextID("Team")
	if err != nil {
		return models.Team{}, errors.E(op, "error finding next available ID", err)
	}
	team.ID = newId
	k := datastore.IDKey("Team", newId, nil)

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return models.Team{}, errors.E(op, "could not create transaction", errors.KindDatabaseError, err)
	}

	var tmp models.Team

	// Perform a Get or Put to ensure atomicity
	err = tx.Get(k, &tmp)
	if err == nil || err != datastore.ErrNoSuchEntity {
		if err == nil {
			err = fmt.Errorf("Datastore 'Get or Put' failed")
		}
		_ = tx.Rollback()
		return models.Team{}, errors.E(op, "concurrency error adding Team", errors.KindConcurrencyProblem, err)
	}

	pk, err := tx.Put(k, &team)
	if err != nil {
		_ = tx.Rollback()
		return models.Team{}, errors.E(op, "error on Put operation for Team", errors.KindDatabaseError, err)
	}

	c, err := tx.Commit()
	if err != nil {
		return models.Team{}, errors.E(op, "error committing transaction", errors.KindDatabaseError, err)
	}

	k = c.Key(pk)
	if k.ID != newId {
		panic("keys don't match")
	}

	return team, nil
}

func (db *DatastoreClient) GetTeam(id int64) (team models.Team, err error) {
	const op errors.Op = "datastore.GetTeam"
	ctx := context.Background()

	k := datastore.IDKey("Team", id, nil)
	err = db.client.Get(ctx, k, &team)

	if err == datastore.ErrNoSuchEntity {
		err = errors.E(errors.KindNotFound, op, err)
	} else if err != nil {
		err = errors.E(op, err)
	}

	return
}

func (db *DatastoreClient) GetTeams() (teams []models.Team, err error) {
	const op errors.Op = "datastore.GetTeams"
	ctx := context.Background()

	q := datastore.NewQuery("Team").Order("ID")

	_, err = db.client.GetAll(ctx, q, &teams)

	if err != nil {
		return nil, errors.E(op, err, errors.KindDatabaseError, "error getting Teams")
	}

	return
}

func (db *DatastoreClient) GetTeamsByID(ids []int64) ([]models.Team, error) {
	const op errors.Op = "datastore.GetTeamsByID"
	ctx := context.Background()

	ks := make([]*datastore.Key, len(ids))
	for i := range ids {
		ks[i] = datastore.IDKey("Team", ids[i], nil)
	}

	teams := make([]models.Team, len(ks))
	err := db.client.GetMulti(ctx, ks, teams)
	if err != nil {
		return nil, errors.E(op, err, errors.KindDatabaseError, "error getting Teams from IDs")
	}

	return teams, nil
}

func (db *DatastoreClient) GetUser(name string) (user models.User, err error) {
	const op errors.Op = "datastore.GetUser"
	ctx := context.Background()

	k := datastore.NameKey("User", name, nil)
	err = db.client.Get(ctx, k, &user)

	if err == datastore.ErrNoSuchEntity {
		err = errors.E(errors.KindNotFound, op, err)
	} else if err != nil {
		err = errors.E(op, err)
	}

	return
}

func (db *DatastoreClient) GetUsers(filters []Filter, sort Sort) ([]models.User, error) {
	const op errors.Op = "datastore.GetUsers"
	ctx := context.Background()

	q := datastore.NewQuery("User")
	q = filterAndSort(q, filters, sort)

	var users []models.User
	_, err := db.client.GetAll(ctx, q, &users)
	if err != nil {
		return nil, errors.E(op, err, errors.KindDatabaseError, "error getting Users")
	}

	// If there are no results, return an empty list instead of nil
	if users == nil {
		users = []models.User{}
	}

	return users, nil
}

func (db *DatastoreClient) AddUser(user models.User) (models.User, error) {
	const op errors.Op = "datastore.AddUser"
	ctx := context.Background()

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return models.User{}, errors.E(op, "could not create transaction", errors.KindDatabaseError, err)
	}

	var tmp models.User

	k := datastore.NameKey("User", user.Nickname, nil)

	// Perform a Get or Put to ensure atomicity
	err = db.client.Get(ctx, k, &tmp)
	if err == nil {
		_ = tx.Rollback()
		return models.User{}, errors.E(op, "user already exists", errors.KindConflict, err)
	} else if err != datastore.ErrNoSuchEntity {
		_ = tx.Rollback()
		return models.User{}, errors.E(op, "concurrency error adding User", errors.KindConcurrencyProblem, err)
	}

	pk, err := tx.Put(k, &user)
	if err != nil {
		_ = tx.Rollback()
		return models.User{}, errors.E(op, "error on Put operation for User", errors.KindDatabaseError, err)
	}

	c, err := tx.Commit()
	if err != nil {
		return models.User{}, errors.E(op, "error committing transaction", errors.KindDatabaseError, err)
	}

	k = c.Key(pk)
	if k.Name != user.Nickname {
		panic("keys don't match")
	}

	return user, nil
}

func (db *DatastoreClient) UpdateUser(user models.User) error {
	const op errors.Op = "datastore.UpdateUser"
	ctx := context.Background()

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return errors.E(op, "could not create transaction", errors.KindDatabaseError, err)
	}

	var oldUser models.User
	k := datastore.NameKey("User", user.Nickname, nil)
	err = tx.Get(k, &oldUser)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, "user not found to update", errors.KindNotFound, err)
	}

	user.VoterEvents = oldUser.VoterEvents
	if user.IsVoter != oldUser.IsVoter {
		user.VoterEvents = append([]models.VoterEvent{{IsVoter: user.IsVoter, EffectiveTime: time.Now()}}, oldUser.VoterEvents...)

		// Adjust future polls
		polls, err := db.futurePolls()
		if err != nil {
			_ = tx.Rollback()
			return errors.E(err, op, errors.KindDatabaseError, "error retrieving polls to adjust for voter change")
		}

		for _, poll := range polls {
			if user.IsVoter {
				poll.MissingVoters = addIfNotExists(poll.MissingVoters, user.Nickname)
			} else {
				poll.MissingVoters = removeIfExists(poll.MissingVoters, user.Nickname)
			}
			pollKey := poll.ToKey()
			_, err = tx.Put(pollKey, &poll)
			if err != nil {
				_ = tx.Rollback()
				return errors.E(err, op, errors.KindDatabaseError, "error adding or removing user as voter for future poll")
			}
		}
	}

	_, err = tx.Put(k, &user)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, "error updating user", errors.KindDatabaseError, err)
	}

	_, err = tx.Commit()
	if err != nil {
		return errors.E(op, "error committing transaction", errors.KindConcurrencyProblem, err)
	}

	return nil
}

func (p *Poll) FromContract(c models.Poll) {
	p.Season = c.Season
	p.Week = c.Week
	p.WeekName = c.WeekName
	p.OpenTime = c.OpenTime
	p.CloseTime = c.CloseTime
	p.LastModified = c.LastModified
}

func (p *Poll) ToContract() (cp models.Poll) {
	cp.Season = p.Season
	cp.Week = p.Week
	cp.WeekName = p.WeekName
	cp.OpenTime = p.OpenTime
	cp.CloseTime = p.CloseTime
	cp.LastModified = p.LastModified

	return cp
}

func (p *Poll) ToKey() (key *datastore.Key) {
	return pollKeyFromWeek(p.Season, p.Week)
}

func pollKeyFromWeek(season, week int) *datastore.Key {
	return datastore.NameKey("Poll", fmt.Sprintf("%v-%v", season, week), nil)
}

type ballotPair struct {
	BallotID     int64
	UserNickname string
}

type Poll struct {
	Season            int
	Week              int
	WeekName          string
	OpenTime          time.Time
	CloseTime         time.Time
	LastModified      time.Time
	RedditURL         string
	Results           []Result
	UnofficialResults []Result
	Ballots           []ballotPair
	MissingVoters     []string
}

func (r *Result) ToContract() models.Result {
	contract := models.Result{
		TeamID:          r.TeamID,
		TeamName:        r.TeamName,
		TeamSlug:        r.TeamSlug,
		Rank:            r.Rank,
		FirstPlaceVotes: r.FirstPlaceVotes,
		Points:          r.Points,
	}

	return contract
}

func (r *Result) FromContract(cr models.Result) {
	r.TeamID = cr.TeamID
	r.TeamName = cr.TeamName
	r.TeamSlug = cr.TeamSlug
	r.Rank = cr.Rank
	r.FirstPlaceVotes = cr.FirstPlaceVotes
	r.Points = cr.Points
}

type Result struct {
	TeamID          int64
	TeamName        string
	TeamSlug        string
	Rank            int
	FirstPlaceVotes int
	Points          int
}

// todo check for existing poll inside transaction
func (db *DatastoreClient) AddPoll(newPoll models.Poll) (models.Poll, error) {
	const op errors.Op = "datastore.AddPoll"
	ctx := context.Background()
	var poll Poll
	poll.FromContract(newPoll)

	k := poll.ToKey()
	poll.LastModified = time.Now()
	currVoters, err := db.allVoters()
	if err != nil {
		return models.Poll{}, errors.E(op, err, "error retrieving voter list")
	}
	poll.MissingVoters = currVoters

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return models.Poll{}, errors.E(op, errors.KindDatabaseError, "unable to create transaction", err)
	}

	var tmp Poll
	err = tx.Get(k, &tmp)
	if err == nil {
		_ = tx.Rollback()
		return models.Poll{}, errors.E(op, errors.KindConflict, "poll for season/week pair already exists", err)
	} else if err != datastore.ErrNoSuchEntity {
		_ = tx.Rollback()
		return models.Poll{}, errors.E(op, errors.KindDatabaseError, "error checking for existing poll", err)
	}

	_, err = tx.Put(k, &poll)
	if err != nil {
		_ = tx.Rollback()
		return models.Poll{}, errors.E(op, "error on Put operation for Poll", errors.KindDatabaseError, err)
	}

	_, err = tx.Commit()
	if err != nil {
		return models.Poll{}, errors.E(op, errors.KindDatabaseError, "error committing transaction", err)
	}

	contract := poll.ToContract()

	return contract, nil
}

func (db *DatastoreClient) GetPoll(season int, week int) (models.Poll, error) {
	const op errors.Op = "datastore.GetPoll"
	ctx := context.Background()

	k := pollKeyFromWeek(season, week)

	var poll Poll
	err := db.client.Get(ctx, k, &poll)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return models.Poll{}, errors.E(op, "poll not found", errors.KindNotFound, err)
		}
		return models.Poll{}, errors.E(op, "error on Get operation for Poll", errors.KindDatabaseError, err)
	}

	contract := poll.ToContract()

	return contract, nil
}

func (db *DatastoreClient) UpdatePoll(poll models.Poll) error {
	const op errors.Op = "datastore.UpdatePoll"
	ctx := context.Background()

	var updatedPoll Poll
	updatedPoll.FromContract(poll)

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return errors.E(op, errors.KindDatabaseError, "unable to create transaction", err)
	}

	k := updatedPoll.ToKey()
	var storedPoll Poll
	err = tx.Get(k, &storedPoll)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, errors.KindDatabaseError, "unable to retrieve poll", err)
	}

	if storedPoll.LastModified != updatedPoll.LastModified {
		// Poll has changed since client made their modifications. Return an error
		// to keep results consistent.
		err = errors.E(op, errors.KindConcurrencyProblem, "poll out of date when attempting to update")
		_ = tx.Rollback()
		return err
	}

	// Update poll fields from param value
	storedPoll.FromContract(poll)

	_, err = tx.Put(k, &storedPoll)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, errors.KindDatabaseError, "error writing poll to db", err)
	}

	_, err = tx.Commit()
	if err != nil {
		return errors.E(op, errors.KindDatabaseError, "error committing transaction", err)
	}

	return nil
}

func (db *DatastoreClient) SetResults(poll models.Poll, official []models.Result, allBallots []models.Result) error {
	const op errors.Op = "datastore.SetResults"
	ctx := context.Background()

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return errors.E(op, errors.KindDatabaseError, "unable to create transaction", err)
	}

	var dbPoll Poll
	dbPoll.FromContract(poll)

	k := dbPoll.ToKey()

	err = tx.Get(k, &dbPoll)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, errors.KindDatabaseError, "unable to retrieve poll", err)
	}

	if dbPoll.LastModified != poll.LastModified {
		// Poll has changed since client calculated results. Return an error
		// to keep results consistent.
		err = errors.E(op, errors.KindConcurrencyProblem, "poll out of date when attempting to set results")
		_ = tx.Rollback()
		return err
	}

	var dbResults []Result

	for _, result := range official {
		var dbResult Result
		dbResult.FromContract(result)
		dbResults = append(dbResults, dbResult)
	}
	dbPoll.Results = dbResults

	dbResults = nil
	for _, result := range allBallots {
		var dbResult Result
		dbResult.FromContract(result)
		dbResults = append(dbResults, dbResult)
	}
	dbPoll.UnofficialResults = dbResults

	_, err = tx.Put(k, &dbPoll)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, errors.KindDatabaseError, "error writing poll to db", err)
	}

	_, err = tx.Commit()
	if err != nil {
		return errors.E(op, errors.KindConcurrencyProblem, "error committing transaction", err)
	}

	return nil
}

func (db *DatastoreClient) GetResults(poll models.Poll, includeProvisional bool) ([]models.Result, error) {
	const op errors.Op = "datastore.GetResults"
	ctx := context.Background()

	var dbPoll Poll
	dbPoll.FromContract(poll)
	k := dbPoll.ToKey()

	err := db.client.Get(ctx, k, &dbPoll)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, errors.E(op, errors.KindNotFound, "poll not found", err)
		}
		return nil, errors.E(op, errors.KindDatabaseError, "error retrieving Poll from db", err)
	}

	var dbResults []Result
	if includeProvisional {
		dbResults = dbPoll.UnofficialResults
	} else {
		dbResults = dbPoll.Results
	}

	var contracts []models.Result
	for _, dbResult := range dbResults {
		contracts = append(contracts, dbResult.ToContract())
	}

	return contracts, nil
}

type Ballot struct {
	ID          int64
	Poll        *datastore.Key
	UpdatedTime time.Time
	User        *datastore.Key
	Votes       []models.Vote
	IsOfficial  bool
}

func ballotToContract(b Ballot) models.Ballot {
	keyName := b.Poll.Name
	strs := strings.Split(keyName, "-")
	if len(strs) != 2 {
		return models.Ballot{}
	}

	season, err := strconv.Atoi(strs[0])
	if err != nil {
		return models.Ballot{}
	}

	week, err := strconv.Atoi(strs[1])
	if err != nil {
		return models.Ballot{}
	}

	contract := models.Ballot{
		ID:          b.ID,
		PollSeason:  season,
		PollWeek:    week,
		UpdatedTime: b.UpdatedTime,
		User:        b.User.Name,
		Votes:       b.Votes,
		IsOfficial:  b.IsOfficial,
	}

	return contract
}

func ballotFromContract(c models.Ballot) Ballot {
	ballot := Ballot{
		ID:          c.ID,
		Poll:        pollKeyFromWeek(c.PollSeason, c.PollWeek),
		UpdatedTime: c.UpdatedTime,
		User:        datastore.NameKey("User", c.User, nil),
		Votes:       c.Votes,
		IsOfficial:  c.IsOfficial,
	}

	return ballot
}

func (db *DatastoreClient) AddBallot(newBallot models.Ballot) (models.Ballot, error) {
	const op errors.Op = "datastore.AddBallot"
	ctx := context.Background()

	// get next available ID
	newID, err := db.nextID("Ballot")
	if err != nil {
		return models.Ballot{}, errors.E(op, "error finding next available ID", err)
	}
	newBallot.ID = newID

	ballot := ballotFromContract(newBallot)
	k := datastore.IDKey("Ballot", newID, nil)

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return models.Ballot{}, errors.E(op, "could not create transaction", errors.KindDatabaseError, err)
	}

	var tmp models.Ballot

	// Perform a Get or Put to ensure atomicity on selected ID
	err = tx.Get(k, &tmp)
	if err == nil || err != datastore.ErrNoSuchEntity {
		if err == nil {
			err = fmt.Errorf("datastore 'Get or Put' failed")
		}
		_ = tx.Rollback()
		return models.Ballot{}, errors.E(op, "concurrency error adding Ballot", errors.KindConcurrencyProblem, err)
	}

	var poll Poll
	err = tx.Get(ballot.Poll, &poll)
	if err != nil {
		_ = tx.Rollback()
		return models.Ballot{}, errors.E(op, "error retrieving poll for Ballot", errors.KindDatabaseError)
	}

	for _, bp := range poll.Ballots {
		if bp.UserNickname == newBallot.User {
			_ = tx.Rollback()
			return models.Ballot{}, errors.E(op, "user already has a ballot submitted for this poll", errors.KindConflict, err)
		}
	}

	pk, err := tx.Put(k, &ballot)
	if err != nil {
		_ = tx.Rollback()
		return models.Ballot{}, errors.E(op, "error on Put operation for Ballot", errors.KindDatabaseError, err)
	}

	// Add ballot to Poll's list of ballots, clear all results
	poll.Ballots = append(poll.Ballots, ballotPair{BallotID: newID, UserNickname: newBallot.User})
	poll.Results = nil
	poll.UnofficialResults = nil

	// Remove the user from the list of missing voters if they are there
	poll.MissingVoters = removeIfExists(poll.MissingVoters, newBallot.User)
	poll.LastModified = time.Now()

	_, err = tx.Put(ballot.Poll, &poll)
	if err != nil {
		_ = tx.Rollback()
		return models.Ballot{}, errors.E(op, "error adding Ballot to Poll entity", errors.KindDatabaseError)
	}

	c, err := tx.Commit()
	if err != nil {
		return models.Ballot{}, errors.E(op, "error committing transaction", errors.KindConcurrencyProblem, err)
	}

	k = c.Key(pk)
	if k.ID != newID {
		panic("keys don't match")
	}

	return ballotToContract(ballot), nil
}

func (db *DatastoreClient) UpdateBallot(ballot models.Ballot) error {
	const op errors.Op = "datastore.UpdateBallot"
	ctx := context.Background()

	dbBallot := ballotFromContract(ballot)
	k := datastore.IDKey("Ballot", dbBallot.ID, nil)
	pollKey := dbBallot.Poll

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return errors.E(op, "could not create transaction", errors.KindDatabaseError, err)
	}

	var poll Poll

	err = tx.Get(pollKey, &poll)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, "error retrieving poll for Ballot", errors.KindDatabaseError, err)
	}

	// invalidate poll results
	poll.Results = nil
	poll.LastModified = time.Now()
	_, err = tx.Put(pollKey, &poll)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, "error clearing poll results", errors.KindDatabaseError, err)
	}

	_, err = tx.Put(k, &dbBallot)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, "error updating poll", errors.KindDatabaseError, err)
	}

	_, err = tx.Commit()
	if err != nil {
		return errors.E(op, "error committing transaction", errors.KindDatabaseError, err)
	}

	return nil
}

func (db *DatastoreClient) DeleteBallot(id int64) error {
	const op errors.Op = "datastore.DeleteBallot"
	ctx := context.Background()

	k := datastore.IDKey("Ballot", id, nil)

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return errors.E(op, "could not create transaction", errors.KindDatabaseError, err)
	}

	var poll Poll
	var ballot Ballot
	var user models.User

	err = tx.Get(k, &ballot)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = errors.E(op, "ballot doesn't exist", errors.KindNotFound, err)
		}
		_ = tx.Rollback()
		return errors.E(op, "error retrieving ballot to delete", err)
	}

	pollKey := ballot.Poll
	err = tx.Get(pollKey, &poll)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, "error retrieving poll for ballot", errors.KindDatabaseError, err)
	}

	newBallots := make([]ballotPair, 0, len(poll.Ballots))
	for _, ref := range poll.Ballots {
		if ref.BallotID != id {
			newBallots = append(newBallots, ref)
		}
	}

	poll.Ballots = newBallots
	poll.Results = nil
	poll.LastModified = time.Now()

	userKey := ballot.User
	err = tx.Get(userKey, &user)
	if err != nil && err != datastore.ErrNoSuchEntity {
		// if user doesn't exist we'll best effort delete this ballot--don't want unremovable ballots in
		// the case we need to delete a user from the system
		_ = tx.Rollback()
		return errors.E(op, "error retrieving user for ballot", errors.KindDatabaseError, err)
	}

	// If user is a voter we need to add them to the poll's list of missing voters
	if user.IsVoter {
		poll.MissingVoters = addIfNotExists(poll.MissingVoters, user.Nickname)
	}

	_, err = tx.Put(pollKey, &poll)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, "error updating associated poll for deleted ballot", errors.KindDatabaseError, err)
	}

	err = tx.Delete(k)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, "error deleting ballot", errors.KindDatabaseError, err)
	}

	_, err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, "error committing transaction", errors.KindDatabaseError, err)
	}

	return nil
}

func (db *DatastoreClient) GetBallot(id int64) (ballot models.Ballot, err error) {
	const op errors.Op = "datastore.GetBallot"
	ctx := context.Background()

	var b Ballot
	k := datastore.IDKey("Ballot", id, nil)
	err = db.client.Get(ctx, k, &b)

	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = errors.E(errors.KindNotFound, op, err)
		} else {
			err = errors.E(errors.KindDatabaseError, op, err)
		}
		return models.Ballot{}, err
	}

	return ballotToContract(b), nil
}

func (db *DatastoreClient) GetBallotsByPoll(poll models.Poll) ([]models.Ballot, error) {
	const op errors.Op = "datastore.GetBallotsByPoll"
	ctx := context.Background()

	var tmp Poll
	tmp.FromContract(poll)

	k := tmp.ToKey()
	err := db.client.Get(ctx, k, &tmp)

	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			err = errors.E(errors.KindNotFound, op, err)
		} else {
			err = errors.E(errors.KindDatabaseError, op, err, "error retrieving poll")
		}
		return nil, err
	}

	var ks []*datastore.Key

	for _, bp := range tmp.Ballots {
		ks = append(ks, datastore.IDKey("Ballot", bp.BallotID, nil))
	}

	bs := make([]*Ballot, len(ks))

	err = db.client.GetMulti(ctx, ks, bs)
	if err != nil {
		err = errors.E(errors.KindDatabaseError, op, err, "error retrieving ballots by ID")
		return nil, err
	}

	var contracts []models.Ballot

	for _, ballot := range bs {
		contracts = append(contracts, ballotToContract(*ballot))
	}

	return contracts, nil
}

// Helpers

func filterAndSort(q *datastore.Query, filters []Filter, sort Sort) *datastore.Query {
	for _, filter := range filters {
		q = q.Filter(fmt.Sprintf("%s %s", filter.Field, filter.Operator), filter.Value)
	}
	if sort.field != "" {
		sortStr := sort.field
		if !sort.asc {
			sortStr = "-" + sortStr
		}

		q = q.Order(sortStr)
	}
	return q
}

func (db *DatastoreClient) allVoters() ([]string, error) {
	const op errors.Op = "datastore.allVoters"
	ctx := context.Background()

	q := datastore.NewQuery("User").Filter("IsVoter =", true)

	var users []models.User
	_, err := db.client.GetAll(ctx, q, &users)
	if err != nil {
		return nil, errors.E(op, err, errors.KindDatabaseError, "error getting Users")
	}

	// If there are no results, return an empty list instead of nil
	if users == nil {
		users = []models.User{}
	}

	var nicknames []string
	for _, user := range users {
		nicknames = append(nicknames, user.Nickname)
	}

	return nicknames, nil
}

func (db *DatastoreClient) futurePolls() ([]Poll, error) {
	const op errors.Op = "datastore.futurePolls"
	ctx := context.Background()

	q := datastore.NewQuery("Poll").Filter("CloseTime >", time.Now())

	var polls []Poll
	_, err := db.client.GetAll(ctx, q, &polls)
	if err != nil {
		return nil, errors.E(op, err, errors.KindDatabaseError, "error getting polls closing in future")
	}

	if polls == nil {
		polls = []Poll{}
	}

	return polls, nil
}

func removeIfExists(haystack []string, needle string) []string {
	var result []string
	for _, str := range haystack {
		if needle != str {
			result = append(result, str)
		}
	}

	return result
}

func addIfNotExists(haystack []string, needle string) []string {
	found := false
	for _, str := range haystack {
		if needle == str {
			found = true
			break
		}
	}

	if found {
		return haystack
	}

	return append(haystack, needle)
}
