package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestSelectiveMiddleware(t *testing.T) {
	inner := GetTestHandler()
	router := mux.NewRouter()

	// Set up routes, both are handled by GetTestHandler
	ignoredRoute := router.HandleFunc("/foo", inner.ServeHTTP)
	router.HandleFunc("/bar", inner.ServeHTTP)

	// Use SelectiveMiddleware to apply BadRequestWare to the /bar endpoint
	router.Use(SelectiveMiddleware(BadRequestWare, []*mux.Route{ignoredRoute}))

	ts := httptest.NewServer(router)
	defer ts.Close()

	tests := []struct {
		description  string
		path         string
		expectedCode int
	}{
		{
			description:  "No Match",
			path:         "/bar",
			expectedCode: http.StatusBadRequest,
		},
		{
			description:  "Match",
			path:         "/foo",
			expectedCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			url := ts.URL + test.path
			resp, err := http.Get(url)

			if err != nil {
				t.Errorf("unexpected error from http.get: %v", err.Error())
			}

			if resp.StatusCode != test.expectedCode {
				t.Errorf("Expected status code: %v, received: %v", test.expectedCode, resp.StatusCode)
			}
		})
	}
}

// Middleware that returns http.StatusBadRequest every time
func BadRequestWare(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
}

// Handler that always returns StatusOK
func GetTestHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	return http.HandlerFunc(fn)
}
