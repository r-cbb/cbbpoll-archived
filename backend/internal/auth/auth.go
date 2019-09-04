package auth

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

type AuthClient interface {
	Verifier() func(http.Handler) http.Handler
	Authenticator(http.HandlerFunc) http.HandlerFunc
	CreateJWT(u models.User) (string, error)
	UserTokenFromCtx(ctx context.Context) models.UserToken
}

type JwtClient struct {
	auth *jwtauth.JWTAuth
}

func InitJwtAuth(secretReader, publicReader io.Reader) (*JwtClient, error) {
	var op errors.Op = "auth.InitJwtAuth"
	keytext, err := ioutil.ReadAll(secretReader)
	if err != nil {
		return nil, errors.E(op, errors.KindJWTError, err, "error reading secret key")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keytext)
	if err != nil {
		return nil, errors.E(op, errors.KindJWTError, err, "error parsing private key")
	}

	pubtext, err := ioutil.ReadAll(publicReader)
	if err != nil {
		return nil, errors.E(op, errors.KindJWTError, err, "error reading public key")
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubtext)
	if err != nil {
		return nil, errors.E(op, errors.KindJWTError, err, "error parsing public key")
	}

	return &JwtClient{auth: jwtauth.New("RS256", privateKey, pubKey)}, nil
}

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

// should be able to test this
func (j JwtClient) Authenticator(next http.HandlerFunc) http.HandlerFunc {
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
