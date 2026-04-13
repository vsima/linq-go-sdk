package linq

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer spins up an httptest server that invokes handler, and returns
// a Client pointed at it.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient("test-token", WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	return c, srv
}

// readAll is a small helper so tests stay terse.
func readAll(t *testing.T, r io.Reader) []byte {
	t.Helper()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return b
}
