package linq

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestAPIErrorDecoding(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"success":false,"error":{"status":429,"code":1007,"message":"slow down","retry_after":3},"trace_id":"abc-123"}`))
	})
	_, err := c.Chats.Get(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error")
	}
	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("not an APIError: %T", err)
	}
	if ae.StatusCode != 429 {
		t.Errorf("StatusCode = %d", ae.StatusCode)
	}
	if ae.Code != ErrCodeRateLimited {
		t.Errorf("Code = %d", ae.Code)
	}
	if ae.Message != "slow down" {
		t.Errorf("Message = %q", ae.Message)
	}
	if ae.RetryAfter != 3 {
		t.Errorf("RetryAfter = %d", ae.RetryAfter)
	}
	if ae.TraceID != "abc-123" {
		t.Errorf("TraceID = %q", ae.TraceID)
	}
	if !IsRateLimited(err) {
		t.Error("IsRateLimited should be true")
	}
	if IsNotFound(err) || IsUnauthorized(err) {
		t.Error("other predicates should be false")
	}
}

func TestAPIErrorNotFound(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"success":false,"error":{"status":404,"code":1004,"message":"nope"}}`))
	})
	_, err := c.Chats.Get(context.Background(), "x")
	if !IsNotFound(err) {
		t.Errorf("expected IsNotFound, got %v", err)
	}
}

func TestAPIErrorUnauthorized(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	_, err := c.Chats.Get(context.Background(), "x")
	if !IsUnauthorized(err) {
		t.Errorf("expected IsUnauthorized, got %v", err)
	}
}

func TestAPIErrorNonJSONBody(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream boom"))
	})
	_, err := c.Chats.Get(context.Background(), "x")
	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if ae.StatusCode != http.StatusBadGateway {
		t.Errorf("StatusCode = %d", ae.StatusCode)
	}
}

func TestAPIErrorString(t *testing.T) {
	ae := &APIError{StatusCode: 404, Code: 1004, Message: "not found"}
	if ae.Error() == "" {
		t.Error("Error() returned empty")
	}
}
