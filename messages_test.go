package linq

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestMessagesSend(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/chats/c1/messages" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		var body SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Effect == nil || body.Effect.Name != "confetti" {
			t.Errorf("effect = %+v", body.Effect)
		}
		_, _ = w.Write([]byte(`{"id":"m1","parts":[{"type":"text","value":"hi"}],"created_at":"2026-01-01T00:00:00Z","delivery_status":"queued","is_read":false}`))
	})
	msg, err := c.Messages.Send(context.Background(), "c1", &SendMessageRequest{
		Parts:  []MessagePart{NewTextPart("hi")},
		Effect: &MessageEffect{Type: "screen", Name: "confetti"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg.ID != "m1" || len(msg.Parts) != 1 || msg.Parts[0].Text.Value != "hi" {
		t.Errorf("msg = %+v", msg)
	}
}

func TestMessagesGetAndDelete(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/v3/messages/m1" {
				t.Errorf("path = %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"id":"m1","parts":[],"created_at":"2026-01-01T00:00:00Z","delivery_status":"sent","is_read":false}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})
	msg, err := c.Messages.Get(context.Background(), "m1")
	if err != nil {
		t.Fatal(err)
	}
	if msg.DeliveryStatus != DeliverySent {
		t.Errorf("status = %q", msg.DeliveryStatus)
	}
	if err := c.Messages.Delete(context.Background(), "m1"); err != nil {
		t.Fatal(err)
	}
}

func TestMessagesThread(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/messages/m1/thread" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"messages":[{"id":"m2","parts":[],"created_at":"2026-01-01T00:00:00Z","delivery_status":"delivered","is_read":true}]}`))
	})
	res, err := c.Messages.Thread(context.Background(), "m1")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Messages) != 1 || res.Messages[0].ID != "m2" {
		t.Errorf("thread = %+v", res.Messages)
	}
}

func TestMessagesList(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/chats/c1/messages" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("limit = %q", r.URL.Query().Get("limit"))
		}
		_, _ = w.Write([]byte(`{"messages":[],"next_cursor":""}`))
	})
	if _, err := c.Messages.List(context.Background(), "c1", &ListMessagesParams{Limit: 5}); err != nil {
		t.Fatal(err)
	}
}
