package auth

import (
	"net/http"
	"testing"
)

func getTestHandler() http.HandlerFunc {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		panic("test entered test handler, this should not happen")
	}
	return http.HandlerFunc(fn)
}

func TestAuthenticator(t *testing.T) {

}