// Counts application-layer HTTP traffic on both ends with a monotonic timer.
// The route group is fixed; paths, queries, headers and payloads never enter the event.
package httpapi

import (
	"github.com/cjtrowbridge/apparat/internal/adapters/telemetry"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"io"
	"net/http"
	"time"
)

type WorkerSurface struct {
	Workspace OwnerWorkspace
	Metrics   *telemetry.Ring
}
type countedBody struct {
	io.ReadCloser
	bytes int
}

func (c *countedBody) Read(p []byte) (int, error) {
	n, e := c.ReadCloser.Read(p)
	c.bytes += n
	return n, e
}

type countedResponse struct {
	http.ResponseWriter
	bytes, status int
}

func (c *countedResponse) Write(p []byte) (int, error) {
	n, e := c.ResponseWriter.Write(p)
	c.bytes += n
	return n, e
}
func (c *countedResponse) WriteHeader(code int) {
	c.status = code
	c.ResponseWriter.WriteHeader(code)
}
func traffic(ring *telemetry.Ring, start time.Time, route, id string, in, out int, failed bool) {
	if ring == nil {
		return
	}
	op := "http-other"
	switch route {
	case "/v1/workspace":
		op = "http-workspace"
	case "/v1/sync":
		op = "http-sync"
	case "/v1/auth", "/v1/auth/challenge":
		op = "http-auth"
	case "/v1/admin/git/projects":
		op = "http-catalog"
	}
	if !domain.ValidID(id) {
		id = ""
	}
	ring.Observe(telemetry.Event{Operation: op, Elapsed: time.Since(start).Microseconds(), Received: in, Sent: out, Correlation: id, Failed: failed})
}
