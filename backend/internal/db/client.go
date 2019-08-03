package db

import "github.com/r-cbb/cbbpoll/backend/pkg"

type DBClient interface {
	AddTeam(team pkg.Team) (id int64, err error)
	GetTeam(id int64) (team pkg.Team, err error)
}
