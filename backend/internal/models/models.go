package models

import (
	"context"
	"github.com/go-chi/jwtauth"
	"github.com/r-cbb/cbbpoll/internal/errors"
)

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

func UserTokenFromContext(ctx context.Context) UserToken {
	const op errors.Op = "models.userFromContext"
	_, claims, err := jwtauth.FromContext(ctx)

	if err != nil {
		return UserToken{}
	}

	nickname, ok := claims["name"].(string)
	if !ok {
		// Always expect to have a 'name' claim.  If we don't then something is very wrong.
		// We'll treat it like no credential at all.
		return UserToken{}
	}

	isAdmin := claims["admin"].(bool)

	return UserToken{
		Nickname: nickname,
		IsAdmin: isAdmin,
	}
}

func (u UserToken) LoggedIn() bool {
	return u.Nickname != ""
}