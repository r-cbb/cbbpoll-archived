package sqlite

import "github.com/r-cbb/cbbpoll/internal/models"

type Team struct {
	ID         int64
	FullName   string `db:"full_name"`
	ShortName  string `db:"short_name"`
	Nickname   string
	Conference string
}

func (t *Team) fromContract(ct models.Team) {
	t.ID = ct.ID
	t.FullName = ct.FullName
	t.ShortName = ct.ShortName
	t.Nickname = ct.Nickname
	t.Conference = ct.Conference
}

func (t *Team) toContract() models.Team {
	ct := models.Team{
		ID:         t.ID,
		FullName:   t.FullName,
		ShortName:  t.ShortName,
		Nickname:   t.Nickname,
		Conference: t.Conference,
	}

	return ct
}

type User struct {
	Nickname string
	IsAdmin  bool `db:"is_admin"`
	IsVoter  bool `db:"is_voter"`
}

func (u *User) fromContract(cu models.User) {
	u.Nickname = cu.Nickname
	u.IsAdmin = cu.IsAdmin
	u.IsVoter = cu.IsVoter
}

func (u *User) toContract() models.User {
	cu := models.User{
		Nickname: u.Nickname,
		IsAdmin: u.IsAdmin,
		IsVoter: u.IsVoter,
	}

	return cu
}
