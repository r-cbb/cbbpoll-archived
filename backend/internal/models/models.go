package models

// todo struct tags for json encoding/decoding

type Team struct {
	ID         int64
	FullName   string
	ShortName  string
	Nickname   string
	Conference string
}

type User struct {
	Nickname string
	IsAdmin  bool
}

/* Information stored in the jwt credentials for a user, allowing
various properties/permissions to be determined without grabbing
their User model from the database.

The semantics here are that a handler function can always grab a UserToken
from a Context (but it may be the "zero" UserToken) and no UserToken methods
access the database.  The zero UserToken represents a request without credentials (anonymous user).
*/
type UserToken struct {
	Nickname string
	IsAdmin  bool
}

func (u UserToken) LoggedIn() bool {
	return u.Nickname != ""
}