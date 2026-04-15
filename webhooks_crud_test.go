package linq

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestWebhooksList(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/webhook-subscriptions" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"subscriptions":[{"id":"s1","target_url":"https://x/h","subscribed_events":["message.sent"],"is_active":true}]}`))
	})
	res, err := c.Webhooks.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Subscriptions) != 1 || res.Subscriptions[0].ID != "s1" || res.Subscriptions[0].TargetURL != "https://x/h" {
		t.Errorf("subs = %+v", res.Subscriptions)
	}
}

func TestWebhooksCreate(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		var body CreateWebhookSubscriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.TargetURL != "https://x/h" {
			t.Errorf("target_url = %q", body.TargetURL)
		}
		if len(body.SubscribedEvents) != 1 || body.SubscribedEvents[0] != "message.sent" {
			t.Errorf("subscribed_events = %+v", body.SubscribedEvents)
		}
		_, _ = w.Write([]byte(`{"id":"s1","target_url":"https://x/h","subscribed_events":["message.sent"],"signing_secret":"shh","is_active":true}`))
	})
	sub, err := c.Webhooks.Create(context.Background(), &CreateWebhookSubscriptionRequest{
		TargetURL:        "https://x/h",
		SubscribedEvents: []string{"message.sent"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sub.ID != "s1" || sub.SigningSecret != "shh" {
		t.Errorf("sub = %+v", sub)
	}
}

func TestWebhooksUpdate(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/v3/webhook-subscriptions/s1" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		var body UpdateWebhookSubscriptionRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.IsActive == nil || *body.IsActive {
			t.Errorf("is_active = %v", body.IsActive)
		}
		_, _ = w.Write([]byte(`{"id":"s1","target_url":"https://x/h","subscribed_events":["message.sent"],"is_active":false}`))
	})
	active := false
	sub, err := c.Webhooks.Update(context.Background(), "s1", &UpdateWebhookSubscriptionRequest{IsActive: &active})
	if err != nil {
		t.Fatal(err)
	}
	if sub.IsActive {
		t.Errorf("expected is_active=false, got %v", sub.IsActive)
	}
}

func TestWebhooksDelete(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v3/webhook-subscriptions/s1" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	if err := c.Webhooks.Delete(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}
}

func TestSendVoiceMemo(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/chats/c1/voicememo" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		var body SendVoiceMemoRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.AttachmentID != "att-1" {
			t.Errorf("attachment_id = %q", body.AttachmentID)
		}
		_, _ = w.Write([]byte(`{"id":"m1","parts":[],"created_at":"2026-01-01T00:00:00Z","delivery_status":"queued","is_read":false}`))
	})
	msg, err := c.Chats.SendVoiceMemo(context.Background(), "c1", &SendVoiceMemoRequest{AttachmentID: "att-1"})
	if err != nil {
		t.Fatal(err)
	}
	if msg.ID != "m1" {
		t.Errorf("msg = %+v", msg)
	}
}

func TestNewMediaPartByURL(t *testing.T) {
	p := NewMediaPartByURL("https://cdn/cat.png")
	if p.Media == nil || p.Media.URL == nil || *p.Media.URL != "https://cdn/cat.png" {
		t.Errorf("part = %+v", p)
	}
}
