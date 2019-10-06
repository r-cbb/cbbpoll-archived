package sqlite

import (
	"time"

	"github.com/r-cbb/cbbpoll/internal/models"
)

type Team struct {
	ID         int64
	FullName   string `db:"full_name"`
	ShortName  string `db:"short_name"`
	Nickname   string
	Conference string
	Slug       string
}

func (t *Team) fromContract(ct models.Team) {
	t.ID = ct.ID
	t.FullName = ct.FullName
	t.ShortName = ct.ShortName
	t.Nickname = ct.Nickname
	t.Conference = ct.Conference
	t.Slug = ct.Slug
}

func (t *Team) toContract() models.Team {
	ct := models.Team{
		ID:         t.ID,
		FullName:   t.FullName,
		ShortName:  t.ShortName,
		Nickname:   t.Nickname,
		Conference: t.Conference,
		Slug:       t.Slug,
	}

	return ct
}

type User struct {
	Nickname    string
	IsAdmin     bool   `db:"is_admin"`
	IsVoter     bool   `db:"is_voter"`
	PrimaryTeam *int64 `db:"primary_team"`
}

func (u *User) fromContract(cu models.User) {
	u.Nickname = cu.Nickname
	u.IsAdmin = cu.IsAdmin
	u.IsVoter = cu.IsVoter
	if cu.PrimaryTeam != 0 {
		u.PrimaryTeam = &cu.PrimaryTeam
	}
}

func (u *User) toContract() models.User {
	cu := models.User{
		Nickname: u.Nickname,
		IsAdmin:  u.IsAdmin,
		IsVoter:  u.IsVoter,
	}

	if u.PrimaryTeam != nil {
		cu.PrimaryTeam = *u.PrimaryTeam
	}

	return cu
}

type Poll struct {
	Season       int
	Week         int
	WeekName     string    `db:"week_name"`
	OpenTime     time.Time `db:"open_time"`
	CloseTime    time.Time `db:"close_time"`
	LastModified time.Time `db:"last_modified"`
	RedditURL    string    `db:"reddit_url"`
}

func (p *Poll) fromContract(cp models.Poll) {
	p.Season = cp.Season
	p.Week = cp.Week
	p.WeekName = cp.WeekName
	p.OpenTime = cp.OpenTime
	p.CloseTime = cp.CloseTime
	p.LastModified = cp.LastModified
	p.RedditURL = cp.RedditURL
}

func (p *Poll) toContract() models.Poll {
	cp := models.Poll{
		Season:       p.Season,
		Week:         p.Week,
		WeekName:     p.WeekName,
		OpenTime:     p.OpenTime,
		CloseTime:    p.CloseTime,
		LastModified: p.LastModified,
		RedditURL:    p.RedditURL,
	}

	return cp
}

type Ballot struct {
	ID          int64
	PollSeason  int       `db:"poll_season"`
	PollWeek    int       `db:"poll_week"`
	UpdatedTime time.Time `db:"updated_time"`
	User        string
	IsOfficial  bool `db:"is_official"`
}

func (b *Ballot) fromContract(cb models.Ballot) []Vote {
	b.ID = cb.ID
	b.PollSeason = cb.PollSeason
	b.PollWeek = cb.PollWeek
	b.UpdatedTime = cb.UpdatedTime
	b.User = cb.User
	b.IsOfficial = cb.IsOfficial

	vs := make([]Vote, len(cb.Votes))

	for i, cv := range cb.Votes {
		vs[i].fromContract(cv, cb)
	}

	return vs
}

type voteGetter interface {
	getVotes(ballotID int64) ([]Vote, error)
}

func (b *Ballot) toContract(vg voteGetter) (models.Ballot, error) {
	cb := models.Ballot{
		ID:          b.ID,
		PollSeason:  b.PollSeason,
		PollWeek:    b.PollWeek,
		UpdatedTime: b.UpdatedTime,
		User:        b.User,
		IsOfficial:  b.IsOfficial,
	}

	vs, err := vg.getVotes(b.ID)
	if err != nil {
		return models.Ballot{}, err
	}

	cvs := make([]models.Vote, len(vs))
	for i := range vs {
		cvs[i] = vs[i].toContract()
	}

	cb.Votes = cvs

	return cb, nil
}

type Vote struct {
	BallotID int64 `db:"ballot_id"`
	TeamID   int64 `db:"team_id"`
	Rank     int
	Reason   string
}

func (v *Vote) fromContract(cv models.Vote, cb models.Ballot) {
	v.BallotID = cb.ID
	v.TeamID = cv.TeamID
	v.Rank = cv.Rank
	v.Reason = cv.Reason
}

func (v *Vote) toContract() models.Vote {
	cv := models.Vote{
		TeamID: v.TeamID,
		Rank:   v.Rank,
		Reason: v.Reason,
	}

	return cv
}

type Result struct {
	Season          int    `db:"poll_season"`
	Week            int    `db:"poll_week"`
	TeamID          int64  `db:"team_id"`
	TeamName        string `db:"team_name"`
	TeamSlug        string `db:"team_slug"`
	Rank            int
	FirstPlaceVotes int `db:"first_place_votes"`
	Points          int
	Official        bool
}

func (r *Result) fromContract(cr models.Result, cp models.Poll, official bool) {
	r.Season = cp.Season
	r.Week = cp.Week
	r.TeamID = cr.TeamID
	r.TeamName = cr.TeamName
	r.TeamSlug = cr.TeamSlug
	r.Rank = cr.Rank
	r.FirstPlaceVotes = cr.FirstPlaceVotes
	r.Points = cr.Points
	r.Official = official
}

func (r *Result) toContract() models.Result {
	cr := models.Result{
		TeamID:          r.TeamID,
		TeamName:        r.TeamName,
		TeamSlug:        r.TeamSlug,
		Rank:            r.Rank,
		FirstPlaceVotes: r.FirstPlaceVotes,
		Points:          r.Points,
	}

	return cr
}
