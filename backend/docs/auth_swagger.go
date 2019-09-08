package docs

// swagger:route POST /v1/sessions auth new-session
// Request a new session from the server.
//
// Bearer token should be a reddit OAuth access token obtained by completing the reddit oauth flow for installed apps.
// See https://github.com/reddit-archive/reddit/wiki/oauth2 for more information.
// responses:
//   200: sessionResponse
//   400: badRequestError
//   401: unauthorizedError
//   500: unexpectedError
//   503: serviceUnavailableError

// swagger:response sessionResponse
type sessionResponse struct {
	// in: body
	Body struct {
		// example: Concision
		Nickname string `json:"nickname"`
		// example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
		Token    string `json:"token"`
	}
}

// swagger:parameters new-session
type newSessionParameters struct {
	// in: header
	Authorization string
}
