package app

import (
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	"github.com/gorilla/mux"

	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

func createJWT(u models.User, tokenAuth *jwtauth.JWTAuth) (string, error) {
	var op errors.Op = "auth.createJWT"
	var claims jwtauth.Claims = make(map[string]interface{})
	claims["name"] = u.Nickname
	claims["admin"] = u.IsAdmin

	_, tokenString, err := tokenAuth.Encode(claims)
	if err != nil {
		return "", errors.E(op, errors.KindJWTError, err, "error creating jwt for user")
	}

	return tokenString, nil
}

func InitJwtAuth(secretPath, publicPath string) (*jwtauth.JWTAuth, error) {
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

	return jwtauth.New("RS256", privateKey, pubKey), nil
}

func Authenticator(excludes []*mux.Route) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			curr := mux.CurrentRoute(r)
			exclude := false
			for _, e := range excludes {
				if e == curr {
					exclude = true
				}
			}

			if exclude {
				next.ServeHTTP(w, r)
				return
			} else {
				authenticatorHelper(next).ServeHTTP(w, r)
				return
			}
		})
	}
}

func authenticatorHelper(next http.Handler) http.Handler {
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