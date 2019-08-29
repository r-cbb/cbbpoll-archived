package auth

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	"github.com/gorilla/mux"

	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

type AuthClient interface {
	Verifier() func(http.Handler) http.Handler
	Authenticator(ignoredRoutes []*mux.Route) func(http.Handler) http.Handler
	CreateJWT(u models.User) (string, error)
	UserTokenFromCtx(ctx context.Context) models.UserToken
}

type JwtClient struct {
	auth *jwtauth.JWTAuth
}

// TODO: change to take Readers instead of paths, ensure a JwtClient is created when
// some hardcoded keys are passed in.
func InitJwtAuth(secretPath, publicPath string) (*JwtClient, error) {
	var op errors.Op = "auth.InitJwtAuth"
	keytext, err := ioutil.ReadFile(secretPath)
	if err != nil {
		return nil, errors.E(op, errors.KindJWTError, err, "error reading from secret key file")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keytext)
	if err != nil {
		return nil, errors.E(op, errors.KindJWTError, err, "error parsing private key")
	}

	pubtext, err := ioutil.ReadFile(publicPath)
	if err != nil {
		return nil, errors.E(op, errors.KindJWTError, err, "error reading from public key file")
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubtext)
	if err != nil {
		return nil, errors.E(op, errors.KindJWTError, err, "error parsing public key")
	}

	return &JwtClient{auth: jwtauth.New("RS256", privateKey, pubKey)}, nil
}

// TODO: To test, write some code that places a jwt-go token into a Context under key jwtauth.TokenCtxKey
// Pass context into here and ensure UserToken comes out
func (j JwtClient) UserTokenFromCtx(ctx context.Context) (token models.UserToken) {
	const op errors.Op = "auth.userFromContext"
	_, claims, err := jwtauth.FromContext(ctx)

	if err != nil {
		return
	}

	nickname, ok := claims["name"].(string)
	if !ok {
		// Always expect to have a 'name' claim.  If we don't then something is very wrong.
		// We'll treat it like no credential at all.
		return
	}

	isAdmin := claims["admin"].(bool)

	return models.UserToken{
		Nickname: nickname,
		IsAdmin: isAdmin,
	}
}

func (j JwtClient) Verifier() func (http.Handler) http.Handler {
	return jwtauth.Verifier(j.auth)
}


// TODO: To test, no way around it, going to have to make sure j.auth is a valid JWTAuth object (can reuse
// code from test InitJwtAuth likely).  Pass in a user and then inspect the token that comes out.  (Can probably
// just pass the token into the test for UserTokenFromCtx.)
func (j JwtClient) CreateJWT(u models.User) (string, error) {
	var op errors.Op = "auth.createJWT"
	var claims jwtauth.Claims = make(map[string]interface{})
	claims["name"] = u.Nickname
	claims["admin"] = u.IsAdmin

	_, tokenString, err := j.auth.Encode(claims)
	if err != nil {
		return "", errors.E(op, errors.KindJWTError, err, "error creating jwt for user")
	}

	return tokenString, nil
}

func (j JwtClient) Authenticator(excludes []*mux.Route) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			curr := mux.CurrentRoute(r)
			exclude := false
			for _, e := range excludes {
				if e == curr {
					exclude = true
					break
				}
			}

			if exclude {
				next.ServeHTTP(w, r)
				return
			} else {
				j.authenticatorHelper(next).ServeHTTP(w, r)
				return
			}
		})
	}
}

func (j JwtClient) authenticatorHelper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil && err != jwtauth.ErrNoTokenFound{
			http.Error(w, http.StatusText(401), 401)
			return
		}

		if token == nil {
			// No token provided--which is fine.  Pass it through and let application logic handle it.
			next.ServeHTTP(w, r)
			return
		}

		if !token.Valid {
			http.Error(w, http.StatusText(401), 401)
			return
		}

		next.ServeHTTP(w, r)
	})
}