package docs

import "github.com/r-cbb/cbbpoll/internal/models"

// swagger:route GET /v1/ping meta ping
// Server health check and version information.
// responses:
//   200: pingResponse

// Server version.
// swagger:response pingResponse
type pingResponseWrapper struct {
	// in:body
	Body models.VersionInfo
}
