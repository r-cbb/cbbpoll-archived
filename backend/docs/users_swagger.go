package docs

import "github.com/r-cbb/cbbpoll/internal/models"

// swagger:route GET /v1/users/me users users-me
// Return user information for authenticated user.
// security:
//   api_key: []
// responses:
//    200: userResponse
//    401: unauthorizedError
//    500: unexpectedError

// The requested User object
// swagger:response userResponse
type userResponse struct {
	// in: body
	Body models.User
}

// swagger:operation GET /v1/users/{userId} users get-user
// Retrieve a User by name.
// ---
// parameters:
// - name: userId
//   in: path
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/userResponse"
//     description: A successful request.
//   "404":
//     description: User not found.
//   "5xx":
//     description: Unexpected error.
