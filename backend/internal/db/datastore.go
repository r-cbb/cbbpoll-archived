package db

import (
	"cloud.google.com/go/datastore"
	"context"
	"fmt"
	"github.com/r-cbb/cbbpoll/backend/internal/cbbpoll"
)

// Eventually rename DBClient to DatastoreClient and abstract out an interface type DBClient
type DBClient struct {
	client *datastore.Client
}

func NewDBClient(projectId string) (*DBClient, error) {
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, projectId)
	if err != nil {
		return nil, err
	}

	// Verify that we can communicate and authenticate with the datastore service.
	t, err := client.NewTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("datastoredb: could not connect: %v", err)
	}
	if err := t.Rollback(); err != nil {
		return nil, fmt.Errorf("datastoredb: could not connect: %v", err)
	}

	return &DBClient{client: client}, nil
}

func (db *DBClient) AddTeam(team cbbpoll.Team) (id int64, err error) {
	ctx := context.Background()
	k := datastore.IncompleteKey("Team", nil)

	q := datastore.NewQuery("Team").Order("-ID").Limit(1)

	var teams []cbbpoll.Team

	keys, err := db.client.GetAll(ctx, q, &teams)

	var newId int64
	newId = 0
	if len(keys) > 0 {
		fmt.Printf("greatest ID found: %v", teams[0].ID)
		newId = teams[0].ID + 1
	}

	team.ID = newId

	k, err = db.client.Put(ctx, k, &team)
	if err != nil {
		return 0, fmt.Errorf("datastoredb: could not put Team: %v", err)
	}
	return newId, nil
}

func (db *DBClient) GetTeam(id int64) (team cbbpoll.Team, err error) {
	ctx := context.Background()

	q := datastore.NewQuery("Team").Filter("ID =", id)

	var teams []cbbpoll.Team

	keys, err := db.client.GetAll(ctx, q, &teams)

	if len(keys) != 1 {
		return cbbpoll.Team{}, fmt.Errorf("not found")
	}

	return teams[0], nil
}