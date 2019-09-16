package models

import "time"

type VersionInfo struct {
	// example: v1.0.0
	Version string `json:"version"`
}

type Team struct {
	// example: 1
	ID int64 `json:"id"`
	// example: University of Arizona
	FullName string `json:"full_name"`
	// example: Arizona
	ShortName string `json:"short_name"`
	// example: Wildcats
	Nickname string `json:"nickname"`
	// example: Pac-12
	Conference string `json:"conference"`
}

type User struct {
	// example: Concision
	// required: true
	Nickname string `json:"nickname"`
	// example: false
	IsAdmin bool `json:"is_admin"`
	// example: true
	IsVoter     bool         `json:"is_voter"`
	VoterEvents []VoterEvent `json:"voter_events,omitempty"`
}

type VoterEvent struct {
	IsVoter       bool
	EffectiveTime time.Time
}

type Poll struct {
	// example: 2020
	Season int `json:"season"`
	// example: 3
	Week int `json:"week"`
	// description: used to "pretty up" polls like Preseason, Postseason, "Way-too-early", etc.  Empty otherwise.
	WeekName  string    `json:"week_name,omitempty"`
	OpenTime  time.Time `json:"open_time"`
	CloseTime time.Time `json:"close_time"`
}

type Ballot struct {
	ID          int64     `json:"id"`
	UpdatedTime time.Time `json:"updated_time"`
	User        string    `json:"user"`
	Votes       []Vote    `json:"votes"`
	IsOfficial  bool      `json:"is_official"`
}

type Vote struct {
	// example: 1
	TeamID uint64 `json:"team_id"`
	// example: 1
	Rank int `json:"rank"`
	// example: Great away performances so far led by a strong senior class.
	Reason string `json:"reason,omitempty"`
}

/* Information stored in the jwt credentials for a user, allowing
various properties/permissions to be determined without grabbing
their User model from the database.

The semantics here are that a handler function can always grab a UserToken
from a Context (but it may be the "zero" UserToken) and no UserToken methods
access the database.  The zero UserToken represents a request without credentials (anonymous user).
*/
type UserToken struct {
	// example: Concision
	Nickname string `json:"nickname"`
	// example: true
	IsAdmin bool `json:"is_admin"`
}

func (u UserToken) LoggedIn() bool {
	return u.Nickname != ""
}

func (u UserToken) CanManagePolls() bool {
	return u.IsAdmin
}
