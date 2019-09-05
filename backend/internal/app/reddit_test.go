package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/r-cbb/cbbpoll/internal/errors"
)

var token = "This is the token"

func TestRedditClient_UsernameFromToken(t *testing.T) {
	goodResponse := struct{Name string `json:"name"`}{Name: "Concision"}
	badResponse := struct {Foo string}{Foo: "Bar"}

	tests := []struct {
		description string
		expectedCode int
		errorExpected bool
		expectedKind errors.Code
		response interface{}
	} {
		{
			description: "Positive test",
			expectedCode: http.StatusOK,
			errorExpected: false,
			response: goodResponse,
		},
		{
			description: "Reddit down",
			expectedCode: http.StatusServiceUnavailable,
			errorExpected: true,
			expectedKind: errors.KindServiceUnavailable,
		},
		{
			description: "Bad token",
			expectedCode: http.StatusUnauthorized,
			errorExpected: true,
			expectedKind: errors.KindAuthError,
		},
		{
			description: "Bad data from Reddit",
			expectedCode: http.StatusOK,
			errorExpected: true,
			response: badResponse,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ts := httptest.NewServer(FakeRedditHandler(t, test.expectedCode, test.response))
			defer ts.Close()

			client := NewRedditClient(ts.URL)
			name, err := client.UsernameFromToken(token)
			if test.errorExpected && err == nil {
				t.Errorf("Expected error and didn't get one")
			}

			if test.errorExpected && errors.Kind(err) != test.expectedKind {
				t.Errorf("Wrong kind of error")
			}

			if !test.errorExpected && err != nil {
				t.Errorf("Unexpected error: %v", err.Error())
			}

			if !test.errorExpected && name != "Concision" {
				t.Errorf("Received wrong name")
			}
		})
	}
}

func FakeRedditHandler(t *testing.T, status int, response interface{}) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer " + token {
			t.Errorf("UsernameFromToken didn't properly send along Authorization header")
		}

		w.WriteHeader(status)
		if response != nil {
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				t.Errorf("Error encoding json: %v", err.Error())
			}
		}
	}

	return http.HandlerFunc(fn)
}
