package models

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
	// example: true
	IsAdmin bool `json:"is_admin"`
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
	IsAdmin  bool   `json:"is_admin"`
}

func (u UserToken) LoggedIn() bool {
	return u.Nickname != ""
}
