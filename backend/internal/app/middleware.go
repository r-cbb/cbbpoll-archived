package app

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func SelectiveMiddleware(m Middleware, excludes []*mux.Route) func(http.Handler) http.Handler {
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
				wrapped := next.(http.HandlerFunc)
				wrapped = m(wrapped)
				wrapped.ServeHTTP(w, r)
			}
		})
	}
}

