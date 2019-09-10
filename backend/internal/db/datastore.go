package db

import (
	"context"
	"fmt"

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
		return nil, errors.E(op, err, errors.KindDatabaseError, "error getting all Teams")
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
	if err == nil || err != datastore.ErrNoSuchEntity {
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