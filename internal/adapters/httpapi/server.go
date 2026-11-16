// Serves bounded loopback-only read queries without persistence access.
// Transport validation and safe DTOs preserve the pre-enrollment security gate.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Queries interface {
	Health() domain.NodeHealth
	Device() domain.Device
	List(context.Context) ([]domain.Service, error)
	Get(context.Context, string) (domain.Service, error)
	Capabilities(context.Context) ([]domain.Capability, error)
}
type Server struct {
	server   *http.Server
	listener net.Listener
	done     chan error
}

func Listen(address string, queries Queries) (*Server, error) {
	parsed, err := netip.ParseAddrPort(address)
	if err != nil || !parsed.Addr().IsLoopback() {
		return nil, errors.New("REST listen address must be a numeric loopback IP:port")
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	server := &http.Server{Handler: Handler(queries), ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second, IdleTimeout: 15 * time.Second, MaxHeaderBytes: 8192}
	instance := &Server{server: server, listener: listener, done: make(chan error, 1)}
	go func() {
		instance.done <- server.Serve(listener)
	}()
	return instance, nil
}
func (server *Server) Address() string {
	return server.listener.Addr().String()
}
func (server *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := server.server.Shutdown(ctx)
	if err != nil {
		_ = server.server.Close()
	}
	serveErr := <-server.done
	if err != nil {
		return err
	}
	if errors.Is(serveErr, http.ErrServerClosed) {
		return nil
	}
	return serveErr
}
func Handler(queries Queries) http.Handler {
	slots := make(chan struct{}, 8)
	var sequence atomic.Uint64
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("X-Request-ID", fmt.Sprint(sequence.Add(1)))
		writer.Header().Set("Cache-Control", "no-store")
		if request.Method != http.MethodGet {
			writer.Header().Set("Allow", "GET")
			respond(writer, 405, map[string]string{"error": "method_not_allowed"})
			return
		}
		if request.ContentLength > 0 || len(request.TransferEncoding) > 0 {
			respond(writer, 400, map[string]string{"error": "body_not_allowed"})
			return
		}
		accept := request.Header.Get("Accept")
		if !acceptsJSON(accept) {
			respond(writer, 406, map[string]string{"error": "not_acceptable"})
			return
		}
		select {
		case slots <- struct{}{}:
			defer func() {
				<-slots
			}()
		default:
			respond(writer, 503, map[string]string{"error": "busy"})
			return
		}
		ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
		defer cancel()
		result, err := route(ctx, request.URL.Path, queries)
		if err != nil {
			code := 500
			message := "query_failed"
			if errors.Is(err, domain.ErrNotFound) {
				code = 404
				message = "not_found"
			}
			if errors.Is(err, domain.ErrUnavailable) {
				code = 503
				message = "not_ready"
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				code = 504
				message = "query_cancelled"
			}
			respond(writer, code, map[string]string{"error": message})
			return
		}
		respond(writer, 200, result)
	})
}
func route(ctx context.Context, path string, queries Queries) (any, error) {
	switch path {
	case "/v1/health":
		return queries.Health(), nil
	case "/v1/device":
		return queries.Device(), nil
	case "/v1/capabilities":
		return queries.Capabilities(ctx)
	case "/v1/services":
		return queries.List(ctx)
	}
	prefix := "/v1/services/"
	if !strings.HasPrefix(path, prefix) {
		return nil, domain.ErrNotFound
	}
	remainder := strings.TrimPrefix(path, prefix)
	parts := strings.Split(remainder, "/")
	if len(parts) > 2 || !domain.ValidID(parts[0]) {
		return nil, domain.ErrNotFound
	}
	service, err := queries.Get(ctx, parts[0])
	if err != nil {
		return nil, err
	}
	if len(parts) == 1 {
		return service, nil
	}
	if parts[1] == "capabilities" {
		return service.Capabilities, nil
	}
	return nil, domain.ErrNotFound
}
func respond(writer http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil || len(data) > 1<<20 {
		status = 500
		data = []byte(`{"error":"response_unavailable"}`)
	}
	writer.WriteHeader(status)
	_, _ = writer.Write(data)
}

func acceptsJSON(header string) bool {
	if header == "" {
		return true
	}
	for _, part := range strings.Split(header, ",") {
		media, parameters, err := mime.ParseMediaType(strings.TrimSpace(part))
		if err != nil {
			continue
		}
		quality := 1.0
		{
			value, ok := parameters["q"]
			if ok {
				quality, err = strconv.ParseFloat(value, 64)
			}
		}
		if err == nil && quality > 0 && quality <= 1 && (media == "application/json" || media == "application/*" || media == "*/*") {
			return true
		}
	}
	return false
}
