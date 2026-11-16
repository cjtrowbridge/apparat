// Tests the loopback wire contract, bounded admission and safe failure mapping.
// OpenAPI is parsed from its JSON-compatible YAML representation during tests.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type fixture struct {
	list func(context.Context) ([]domain.Service, error)
}

func (f fixture) Health() domain.NodeHealth {
	return domain.NodeHealth{Version: "0.1.0", State: "ready", Ready: true}
}
func (f fixture) Device() domain.Device {
	return domain.Device{OS: "linux", Architecture: "amd64", IdentityStatus: "absent"}
}
func (f fixture) List(ctx context.Context) ([]domain.Service, error) {
	if f.list != nil {
		return f.list(ctx)
	}
	return []domain.Service{{ID: "00000000000000000000000000", Name: "one", Driver: "mock-text", Enabled: true, Revision: 1, Health: "healthy", Capabilities: []domain.Capability{}}}, nil
}
func (f fixture) Get(ctx context.Context, id string) (domain.Service, error) {
	list, err := f.List(ctx)
	if err != nil {
		return domain.Service{}, err
	}
	for _, s := range list {
		if s.ID == id {
			return s, nil
		}
	}
	return domain.Service{}, domain.ErrNotFound
}
func (f fixture) Capabilities(context.Context) ([]domain.Capability, error) {
	return []domain.Capability{}, nil
}
func TestContractRoutes(t *testing.T) {
	data, err := os.ReadFile("../../../docs/openapi/apparat-v1.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var contract map[string]any
	{
		err = json.Unmarshal(data, &contract)
		if err != nil {
			t.Fatal(err)
		}
	}
	count := 0
	for path, entry := range contract["paths"].(map[string]any) {
		for method, value := range entry.(map[string]any) {
			operation, ok := value.(map[string]any)
			if !ok || operation["x-phase"] != float64(8) {
				continue
			}
			count++
			request := httptest.NewRequest(strings.ToUpper(method), strings.ReplaceAll(path, "{service_id}", "00000000000000000000000000"), nil)
			response := httptest.NewRecorder()
			Handler(fixture{}).ServeHTTP(response, request)
			if response.Code != 200 || response.Header().Get("Content-Type") != "application/json" || response.Header().Get("X-Request-ID") == "" {
				t.Fatal(path, response)
			}
			var body any
			{
				err = json.Unmarshal(response.Body.Bytes(), &body)
				if err != nil {
					t.Fatal(err)
				}
			}
			schema := operation["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
			validateSchema(t, contract, schema, body)
		}
	}
	if count != 6 {
		t.Fatal("implemented route count", count)
	}
}
func validateSchema(t *testing.T, doc, schema map[string]any, value any) {
	t.Helper()
	{
		ref, ok := schema["$ref"].(string)
		if ok {
			schema = doc["components"].(map[string]any)["schemas"].(map[string]any)[strings.TrimPrefix(ref, "#/components/schemas/")].(map[string]any)
		}
	}
	switch schema["type"] {
	case "array":
		for _, item := range value.([]any) {
			validateSchema(t, doc, schema["items"].(map[string]any), item)
		}
	case "object":
		fields := value.(map[string]any)
		properties := schema["properties"].(map[string]any)
		for _, key := range schema["required"].([]any) {
			{
				_, ok := fields[key.(string)]
				if !ok {
					t.Fatalf("missing required %s", key)
				}
			}
		}
		for key, item := range fields {
			property, ok := properties[key]
			if !ok {
				t.Fatalf("undocumented field %s", key)
			}
			validateSchema(t, doc, property.(map[string]any), item)
		}
	case "string":
		{
			_, ok := value.(string)
			if !ok {
				t.Fatal("expected string", value)
			}
		}
	case "boolean":
		{
			_, ok := value.(bool)
			if !ok {
				t.Fatal("expected boolean", value)
			}
		}
	case "integer":
		{
			_, ok := value.(float64)
			if !ok {
				t.Fatal("expected number", value)
			}
		}
	}
}
func TestTransportRejectionsAndSanitization(t *testing.T) {
	for _, test := range []struct {
		method, path, accept, body string
		status                     int
	}{
		{"POST", "/v1/services", "", "", 405}, {"GET", "/v1/services", "text/html", "", 406}, {"GET", "/v1/services", "", "payload", 400}, {"GET", "/v1/services/bad", "", "", 404}, {"GET", "/v1/services/00000000000000000000000001", "", "", 404}, {"GET", "/v1/services/00000000000000000000000000/extra/path", "", "", 404}, {"GET", "/v1/projects", "", "", 404},
	} {
		r := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		r.Header.Set("Accept", test.accept)
		w := httptest.NewRecorder()
		Handler(fixture{}).ServeHTTP(w, r)
		if w.Code != test.status {
			t.Fatal(test, w.Code)
		}
	}
	for _, test := range []struct {
		err    error
		status int
	}{{errors.New("private credential and path"), 500}, {context.Canceled, 504}, {context.DeadlineExceeded, 504}, {domain.ErrUnavailable, 503}} {
		f := fixture{list: func(context.Context) ([]domain.Service, error) {
			return nil, test.err
		}}
		w := httptest.NewRecorder()
		Handler(f).ServeHTTP(w, httptest.NewRequest("GET", "/v1/services", nil))
		if w.Code != test.status || strings.Contains(w.Body.String(), "private") {
			t.Fatal(w)
		}
	}
}
func TestAdmissionLimit(t *testing.T) {
	entered := make(chan struct{}, 8)
	release := make(chan struct{})
	done := make(chan struct{}, 8)
	handler := Handler(fixture{list: func(ctx context.Context) ([]domain.Service, error) {
		entered <- struct{}{}
		select {
		case <-release:
			return []domain.Service{}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}})
	for range 8 {
		go func() {
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/v1/services", nil))
			done <- struct{}{}
		}()
	}
	for range 8 {
		<-entered
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/v1/services", nil))
	close(release)
	for range 8 {
		<-done
	}
	if w.Code != 503 {
		t.Fatal(w.Code)
	}
}
func TestLoopbackOnlyAndShutdown(t *testing.T) {
	for _, address := range []string{"0.0.0.0:0", "localhost:0", "192.0.2.1:0", "[::]:0", "/tmp/api.sock"} {
		{
			server, err := Listen(address, fixture{})
			if err == nil {
				_ = server.Close()
				t.Fatal(address)
			}
		}
	}
	server, err := Listen("127.0.0.1:0", fixture{})
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Get("http://" + server.Address() + "/v1/health")
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	{
		err = server.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestNegotiationAndResponseLimit(t *testing.T) {
	for _, accept := range []string{"application/json; q=1", "text/html, application/json;q=0.8", "application/*"} {
		request := httptest.NewRequest("GET", "/v1/health", nil)
		request.Header.Set("Accept", accept)
		response := httptest.NewRecorder()
		Handler(fixture{}).ServeHTTP(response, request)
		if response.Code != 200 {
			t.Fatal(accept, response.Code)
		}
	}
	response := httptest.NewRecorder()
	respond(response, 200, strings.Repeat("private", 1<<20))
	if response.Code != 500 || strings.Contains(response.Body.String(), "private") {
		t.Fatal("unbounded response")
	}
}
