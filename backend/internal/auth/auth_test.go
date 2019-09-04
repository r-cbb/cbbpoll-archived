package auth

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	jwt2 "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"

	"github.com/r-cbb/cbbpoll/internal/errors"
	"github.com/r-cbb/cbbpoll/internal/models"
)

const publicKeyText = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnzyis1ZjfNB0bBgKFMSv
vkTtwlvBsaJq7S5wA+kzeVOVpVWwkWdVha4s38XM/pa/yr47av7+z3VTmvDRyAHc
aT92whREFpLv9cj5lTeJSibyr/Mrm/YtjCZVWgaOYIhwrXwKLqPr/11inWsAkfIy
tvHWTxZYEcXLgAXFuUuaS3uF9gEiNQwzGTU1v0FqkqTBr4B8nW3HCN47XUu0t8Y0
e+lf4s4OxQawWD79J9/5d3Ry0vbV3Am1FtGJiJvOwRsIfVChDpYStTcHTCMqtvWb
V6L11BWkpzGXSW4Hv43qa+GSYOD2QU68Mb59oSk2OB+BtOLpJofmbGEGgvmwyCI9
MwIDAQAB
-----END PUBLIC KEY-----`

const privateKeyText = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAnzyis1ZjfNB0bBgKFMSvvkTtwlvBsaJq7S5wA+kzeVOVpVWw
kWdVha4s38XM/pa/yr47av7+z3VTmvDRyAHcaT92whREFpLv9cj5lTeJSibyr/Mr
m/YtjCZVWgaOYIhwrXwKLqPr/11inWsAkfIytvHWTxZYEcXLgAXFuUuaS3uF9gEi
NQwzGTU1v0FqkqTBr4B8nW3HCN47XUu0t8Y0e+lf4s4OxQawWD79J9/5d3Ry0vbV
3Am1FtGJiJvOwRsIfVChDpYStTcHTCMqtvWbV6L11BWkpzGXSW4Hv43qa+GSYOD2
QU68Mb59oSk2OB+BtOLpJofmbGEGgvmwyCI9MwIDAQABAoIBACiARq2wkltjtcjs
kFvZ7w1JAORHbEufEO1Eu27zOIlqbgyAcAl7q+/1bip4Z/x1IVES84/yTaM8p0go
amMhvgry/mS8vNi1BN2SAZEnb/7xSxbflb70bX9RHLJqKnp5GZe2jexw+wyXlwaM
+bclUCrh9e1ltH7IvUrRrQnFJfh+is1fRon9Co9Li0GwoN0x0byrrngU8Ak3Y6D9
D8GjQA4Elm94ST3izJv8iCOLSDBmzsPsXfcCUZfmTfZ5DbUDMbMxRnSo3nQeoKGC
0Lj9FkWcfmLcpGlSXTO+Ww1L7EGq+PT3NtRae1FZPwjddQ1/4V905kyQFLamAA5Y
lSpE2wkCgYEAy1OPLQcZt4NQnQzPz2SBJqQN2P5u3vXl+zNVKP8w4eBv0vWuJJF+
hkGNnSxXQrTkvDOIUddSKOzHHgSg4nY6K02ecyT0PPm/UZvtRpWrnBjcEVtHEJNp
bU9pLD5iZ0J9sbzPU/LxPmuAP2Bs8JmTn6aFRspFrP7W0s1Nmk2jsm0CgYEAyH0X
+jpoqxj4efZfkUrg5GbSEhf+dZglf0tTOA5bVg8IYwtmNk/pniLG/zI7c+GlTc9B
BwfMr59EzBq/eFMI7+LgXaVUsM/sS4Ry+yeK6SJx/otIMWtDfqxsLD8CPMCRvecC
2Pip4uSgrl0MOebl9XKp57GoaUWRWRHqwV4Y6h8CgYAZhI4mh4qZtnhKjY4TKDjx
QYufXSdLAi9v3FxmvchDwOgn4L+PRVdMwDNms2bsL0m5uPn104EzM6w1vzz1zwKz
5pTpPI0OjgWN13Tq8+PKvm/4Ga2MjgOgPWQkslulO/oMcXbPwWC3hcRdr9tcQtn9
Imf9n2spL/6EDFId+Hp/7QKBgAqlWdiXsWckdE1Fn91/NGHsc8syKvjjk1onDcw0
NvVi5vcba9oGdElJX3e9mxqUKMrw7msJJv1MX8LWyMQC5L6YNYHDfbPF1q5L4i8j
8mRex97UVokJQRRA452V2vCO6S5ETgpnad36de3MUxHgCOX3qL382Qx9/THVmbma
3YfRAoGAUxL/Eu5yvMK8SAt/dJK6FedngcM3JEFNplmtLYVLWhkIlNRGDwkg3I5K
y18Ae9n7dHVueyslrb6weq7dTkYDi3iOYRW8HRkIQh06wEdbxt0shTzAJvvCQfrB
jg/3747WSsf/zBTcHihTRBdAv6OmdhV4/dD5YBfLAkLrd+mX7iE=
-----END RSA PRIVATE KEY-----`

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.E("Some IO error")
}

func TestInitJwtAuth(t *testing.T) {
	secretReader := bytes.NewBufferString(privateKeyText)
	publicReader := bytes.NewBufferString(publicKeyText)

	_, err := InitJwtAuth(secretReader, publicReader)
	if err != nil {
		t.Errorf("Unexpected error creating JwtClient: %s", err.Error())
	}

	publicReader = bytes.NewBufferString(publicKeyText)

	_, err = InitJwtAuth(new(errorReader), publicReader)
	if err == nil {
		t.Errorf("Expected an IO error when reading private key")
	}
	if errors.Kind(err) != errors.KindJWTError {
		t.Errorf("Error returning from InitJwtAuth should be KindJWTError (%d), found: %d", errors.KindJWTError, errors.Kind(err))
	}

	secretReader = bytes.NewBufferString(privateKeyText)

	_, err = InitJwtAuth(secretReader, new(errorReader))
	if err == nil {
		t.Errorf("Expected an IO error when reading public key")
	}
	if errors.Kind(err) != errors.KindJWTError {
		t.Errorf("Error returning from InitJwtAuth should be KindJWTError (%d), found: %d", errors.KindJWTError, errors.Kind(err))
	}

	secretReader.Reset()
	_, err = secretReader.WriteString("Some garbage")
	if err != nil {
		t.Errorf("Unexpected error appending to buffer")
	}

	publicReader = bytes.NewBufferString(publicKeyText)

	_, err = InitJwtAuth(secretReader, publicReader)
	if err == nil {
		t.Errorf("Expected an error when passing a bad private key")
	}
	if errors.Kind(err) != errors.KindJWTError {
		t.Errorf("Error returning from InitJwtAuth should be KindJWTError (%d), found: %d", errors.KindJWTError, errors.Kind(err))
	}

	secretReader = bytes.NewBufferString(privateKeyText)
	publicReader.Reset()
	publicReader.WriteString("Some garbage")

	_, err = InitJwtAuth(secretReader, publicReader)
	if err == nil {
		t.Errorf("Expected an error when passing a bad private key")
	}
	if errors.Kind(err) != errors.KindJWTError {
		t.Errorf("Error returning from InitJwtAuth should be KindJWTError (%d), found: %d", errors.KindJWTError, errors.Kind(err))
	}
}

func TestJwtClient_EncodeDecode(t *testing.T) {
	// Create a JWT
	secretReader := bytes.NewBufferString(privateKeyText)
	publicReader := bytes.NewBufferString(publicKeyText)

	client, err := InitJwtAuth(secretReader, publicReader)
	if err != nil {
		t.Errorf("Unexpected error creating JwtClient: %s", err.Error())
	}

	u := models.User{
		Nickname: "Concision",
		IsAdmin:  true,
	}

	jwtStr, err := client.CreateJWT(u)
	if err != nil {
		t.Errorf("Unexpected error from CreateJWT: %s", err.Error())
	}

	var claims = make(jwtauth.Claims)
	claims["name"] = u.Nickname
	claims["admin"] = u.IsAdmin

	jwt, _, err := client.auth.Encode(claims)
	if err != nil {
		t.Errorf("Unexpected error from jwtauth.Encode: %s", err.Error())
	}

	if jwt.Raw != jwtStr {
		t.Errorf("JWT returned from JwtClient.CreateJWT() doesn't match jwtauth.Encode()")
	}

	// Decode the JWT into a UserToken
	ctx := context.Background()
	tokenCtx := context.WithValue(ctx, jwtauth.TokenCtxKey, jwt)

	token := client.UserTokenFromCtx(tokenCtx)

	if token.LoggedIn() == false {
		t.Errorf("Token should be logged in")
	}

	if token.Nickname != u.Nickname {
		t.Errorf("Wrong Nickname from UserToken: %s, but expected %s", token.Nickname, u.Nickname)
	}

	if token.IsAdmin != u.IsAdmin {
		t.Errorf("Mismatched IsAdmin states.  From Token: %v, From User: %v", token.IsAdmin, u.IsAdmin)
	}
}

func TestJwtClient_BadTokenIsLoggedOut(t *testing.T) {
	secretReader := bytes.NewBufferString(privateKeyText)
	publicReader := bytes.NewBufferString(publicKeyText)

	client, err := InitJwtAuth(secretReader, publicReader)
	if err != nil {
		t.Errorf("Unexpected error creating JwtClient: %s", err.Error())
	}

	token := client.UserTokenFromCtx(context.Background())

	if token.LoggedIn() {
		t.Errorf("Bad token led to LoggedIn user")
	}
}

func TestJwtClient_Authenticator(t *testing.T) {
	token := jwt2.Token{Claims: jwt2.MapClaims{"foo": "bar"}, Valid: true}
	badToken := token
	badToken.Valid = false
	jwt := JwtClient{}
	tests := []struct {
		description  string
		ctx          context.Context
		expectedCode int
	}{
		{
			description:  "No token",
			ctx:          context.Background(),
			expectedCode: http.StatusOK,
		},
		{
			description: "Good token",
			ctx: context.WithValue(
				context.WithValue(
					context.Background(),
					jwtauth.TokenCtxKey,
					&token),
				jwtauth.ErrorCtxKey,
				nil),
			expectedCode: http.StatusOK,
		},
		{
			description: "Error retrieving token",
			ctx: context.WithValue(
				context.WithValue(
					context.Background(),
					jwtauth.TokenCtxKey,
					&token),
				jwtauth.ErrorCtxKey,
				fmt.Errorf("Some error")),
			expectedCode: http.StatusUnauthorized,
		},
		{
			description:  "Invalid token",
			ctx:          context.WithValue(context.Background(), jwtauth.TokenCtxKey, &badToken),
			expectedCode: http.StatusUnauthorized,
		},
	}

	handler := jwt.Authenticator(GetTestHandler())

	for _, test := range tests {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "http://cbbpoll.com", nil)
		r = r.WithContext(test.ctx)

		handler.ServeHTTP(w, r)
		if w.Result().StatusCode != test.expectedCode {
			t.Errorf("Expected status code: %v, received: %v", test.expectedCode, w.Result().StatusCode)
		}
	}
}

func GetTestHandler() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	return http.HandlerFunc(fn)
}

func TestJwtClient_UserTokenFromCtxError(t *testing.T) {
	client := JwtClient{}
	token := client.UserTokenFromCtx(context.WithValue(context.Background(), jwtauth.ErrorCtxKey, fmt.Errorf("some error")))
	if token.LoggedIn() || token.IsAdmin {
		t.Error("UserTokenFromCtx() with a token error in the context should not return a logged in user")
	}

}
