package docs

import "github.com/r-cbb/cbbpoll/internal/models"

// swagger:route GET /v1/teams teams list-teams
// Retrieve all Teams.
// responses:
//   200: teamsResponse
//   500: unexpectedError

// Unexpected error.
// swagger:response unexpectedError
type unexpectedError struct {}

// Bad request.
// swagger:response badRequestError
type badRequestError struct {}

// User unauthorized.
// swagger:response unauthorizedError
type unauthorizedError struct {}

// Service unavailable.
// swagger:response serviceUnavailableError
type serviceUnavailableError struct {}

// List of Teams.
// swagger:response teamsResponse
type teamsResponseWrapper struct {
	// in: body
	Body []models.Team
}

// swagger:route POST /v1/teams teams add-team
// Create a new Team.
// responses:
//   201: teamCreated
//   400: badRequestError
//   500: unexpectedError

// This response includes a JSON representation of the created Team in the response body and the URI for the team in the Location header
// swagger:response teamCreated
type teamCreatedResponse struct {
	// in: body
	Body models.Team
	// in: header
	Location string
}

// swagger:parameters add-team
type addTeamsParameters struct {
	// Team to be added.  id field will be ignored, if present.
	// in: body
	Body models.Team
}

// swagger:operation GET /v1/teams/{teamId} teams get-team
// Retrieve a Team by ID.
// ---
// parameters:
// - name: teamId
//   in: path
//   type: integer
//   format: int64
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/teamResponse"
//     description: A successful request.
//   "400":
//     description: Unable to parse teamId parameter.
//   "404":
//     description: Team not found.
//   "5xx":
//     description: Unexpected error.

// Team object.
// swagger:response teamResponse
type teamResponseWrapper struct {
	// in: body
	Body models.Team
}
