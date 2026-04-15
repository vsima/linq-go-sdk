package linq

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientHeaders(t *testing.T) {
	var gotAuth, gotUA, gotAccept, gotCT string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept")
		gotCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"chats":[]}`))
	})
	if _, err := c.Chats.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if !strings.HasPrefix(gotUA, "linq-go-sdk/") {
		t.Errorf("User-Agent = %q", gotUA)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q", gotAccept)
	}
	if gotCT != "" {
		t.Errorf("Content-Type on GET = %q, want empty", gotCT)
	}
}

func TestWithUserAgent(t *testing.T) {
	var gotUA string
	handler := func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`{"chats":[]}`))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(ts.Close)
	c := NewClient("t",
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
		WithUserAgent("custom/1.0"),
	)
	_, _ = c.Chats.List(context.Background(), nil)
	if gotUA != "custom/1.0" {
		t.Errorf("User-Agent = %q", gotUA)
	}
}

func TestWithBaseURLTrimsSlash(t *testing.T) {
	c := NewClient("t", WithBaseURL("https://example.com/api/"))
	if c.baseURL != "https://example.com/api" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

func TestContextCancelled(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.Chats.List(ctx, nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestJSONContentTypeOnPOST(t *testing.T) {
	var gotCT string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	})
	_ = c.Chats.AddParticipant(context.Background(), "chat-id", "+15551234567")
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q", gotCT)
	}
}
