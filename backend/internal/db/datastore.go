package db

import (
	"context"
	"fmt"
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

func (db *DatastoreClient) GetUsers(filters Filter) ([]models.User, error) {
	const op errors.Op = "datastore.GetUsers"
	ctx := context.Background()

	q := datastore.NewQuery("User")

	for k, v := range filters {
		q = q.Filter(k+" =", v)
	}

	q = q.Order("Nickname")

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
		return errors.E(op, "user not found to update", errors.KindNotFound, err)
	}

	user.VoterEvents = oldUser.VoterEvents
	if user.IsVoter != oldUser.IsVoter {
		user.VoterEvents = append([]models.VoterEvent{{user.IsVoter, time.Now()}}, oldUser.VoterEvents...)
	}

	_, err = tx.Put(k, &user)
	if err != nil {
		return errors.E(op, "error updating user", errors.KindDatabaseError, err)
	}

	_, err = tx.Commit()
	if err != nil {
		return errors.E(op, "error committing transaction", errors.KindConcurrencyProblem, err)
	}

	return nil
}

func (db *DatastoreClient) AddPoll(newPoll models.Poll) (poll models.Poll, err error) {
	const op errors.Op = "datastore.AddPoll"
	ctx := context.Background()

	k := datastore.IncompleteKey("Poll", nil)
	newPoll.LastModified = time.Now()

	k, err = db.client.Put(ctx, k, &newPoll)
	if err != nil {
		return models.Poll{}, errors.E(op, "error on Put operation for Poll", errors.KindDatabaseError, err)
	}

	newPoll.ID = k.ID

	return newPoll, nil
}

func (db *DatastoreClient) GetPoll(id int64) (models.Poll, error) {
	const op errors.Op = "datastore.GetPoll"
	ctx := context.Background()

	k := datastore.IDKey("Poll", id, nil)
	var poll models.Poll
	err := db.client.Get(ctx, k, &poll)
	if err != nil {
		return models.Poll{}, errors.E(op, errors.KindDatabaseError, "error on Get operation for poll", err)
	}
	poll.ID = id

	return poll, nil
}

func (db *DatastoreClient) GetPollByWeek(season int, week int) (models.Poll, error) {
	const op errors.Op = "datastore.GetPollByWeek"
	ctx := context.Background()

	q := datastore.NewQuery("Poll").Filter("Season =", season).Filter("Week =", week)

	var polls []models.Poll
	ks, err := db.client.GetAll(ctx, q, &polls)
	if err != nil {
		return models.Poll{}, errors.E(op, "error on Get operation for Poll", errors.KindDatabaseError, err)
	}

	if len(polls) > 1 {
		return models.Poll{}, errors.E(op, fmt.Sprintf("more than one poll found for season %v, week %v\n", season, week), errors.KindConflict)
	}

	if len(polls) == 0 {
		return models.Poll{}, errors.E(op, "poll not found", errors.KindNotFound)
	}

	poll := polls[0]
	poll.ID = ks[0].ID

	return poll, nil
}

func (db *DatastoreClient) UpdatePoll(poll models.Poll) error {
	const op errors.Op = "datastore.UpdatePoll"
	ctx := context.Background()

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return errors.E(op, errors.KindDatabaseError, "unable to create transaction", err)
	}

	k := datastore.IDKey("Poll", poll.ID, nil)
	var tmp models.Poll
	err = tx.Get(k, &tmp)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, errors.KindDatabaseError, "unable to retrieve poll", err)
	}

	if tmp.LastModified != poll.LastModified {
		// Poll has changed since client made their modifications. Return an error
		// to keep results consistent.
		err = errors.E(op, errors.KindConcurrencyProblem, "poll out of date when attempting to update")
		_ = tx.Rollback()
		return err
	}

	_, err = tx.Put(k, &poll)
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

type Ballot struct {
	ID          int64
	Poll        *datastore.Key
	UpdatedTime time.Time
	User        *datastore.Key
	Votes       []models.Vote
	IsOfficial  bool
}

func ballotToContract(b Ballot) models.Ballot {
	contract := models.Ballot{
		ID:          b.ID,
		Poll:        b.Poll.ID,
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
		Poll:        datastore.IDKey("Poll", c.Poll, nil),
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
	newId, err := db.nextID("Ballot")
	if err != nil {
		return models.Ballot{}, errors.E(op, "error finding next available ID", err)
	}
	newBallot.ID = newId

	ballot := ballotFromContract(newBallot)
	k := datastore.IDKey("Ballot", newId, nil)

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return models.Ballot{}, errors.E(op, "could not create transaction", errors.KindDatabaseError, err)
	}

	var tmp models.Ballot

	// Perform a Get or Put to ensure atomicity
	err = tx.Get(k, &tmp)
	if err == nil || err != datastore.ErrNoSuchEntity {
		if err == nil {
			err = fmt.Errorf("Datastore 'Get or Put' failed")
		}
		_ = tx.Rollback()
		return models.Ballot{}, errors.E(op, "concurrency error adding Ballot", errors.KindConcurrencyProblem, err)
	}

	pk, err := tx.Put(k, &ballot)
	if err != nil {
		_ = tx.Rollback()
		return models.Ballot{}, errors.E(op, "error on Put operation for Ballot", errors.KindDatabaseError, err)
	}

	var poll models.Poll
	err = tx.Get(ballot.Poll, &poll)
	if err != nil {
		_ = tx.Rollback()
		return models.Ballot{}, errors.E(op, "error retrieving poll for Ballot", errors.KindDatabaseError)
	}

	poll.Ballots = append(poll.Ballots, models.BallotRef{ID: newId, User: ballot.User.Name})
	poll.Results = nil
	poll.LastModified = time.Now()
	_, err = tx.Put(ballot.Poll, &poll)
	if err != nil {
		_ = tx.Rollback()
		return models.Ballot{}, errors.E(op, "error adding Ballot to Poll entity", errors.KindDatabaseError)
	}

	c, err := tx.Commit()
	if err != nil {
		return models.Ballot{}, errors.E(op, "error committing transaction", errors.KindDatabaseError, err)
	}

	k = c.Key(pk)
	if k.ID != newId {
		panic("keys don't match")
	}

	return ballotToContract(ballot), nil
}

func (db *DatastoreClient) EditBallot(ballot models.Ballot) error {
	const op errors.Op = "datastore.EditBallot"
	ctx := context.Background()

	dbBallot := ballotFromContract(ballot)
	k := datastore.IDKey("Ballot", dbBallot.ID, nil)
	pollKey := dbBallot.Poll

	tx, err := db.client.NewTransaction(ctx)
	if err != nil {
		return errors.E(op, "could not create transaction", errors.KindDatabaseError, err)
	}

	var poll models.Poll

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

	var poll models.Poll
	var ballot Ballot

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

	newBallots := make([]models.BallotRef, 0, len(poll.Ballots))
	for _, ref := range poll.Ballots {
		if ref.ID != id {
			newBallots = append(newBallots, ref)
		}
	}
	poll.Ballots = newBallots
	poll.Results = nil
	poll.LastModified = time.Now()

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

	if err == datastore.ErrNoSuchEntity {
		err = errors.E(errors.KindNotFound, op, err)
	} else if err != nil {
		err = errors.E(errors.KindDatabaseError, op, err)
	}

	return ballotToContract(b), nil
}
