package linq

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestChatsList(t *testing.T) {
	var gotQuery string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/chats" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"chats":[{"id":"c1","is_archived":false,"is_group":false,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}],"next_cursor":"n1"}`))
	})
	res, err := c.Chats.List(context.Background(), &ListChatsParams{From: "+15551234567", Limit: 10, Cursor: "abc"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Chats) != 1 || res.Chats[0].ID != "c1" {
		t.Errorf("chats = %+v", res.Chats)
	}
	if res.NextCursor != "n1" {
		t.Errorf("NextCursor = %q", res.NextCursor)
	}
	if !contains(gotQuery, "from=%2B15551234567") || !contains(gotQuery, "limit=10") || !contains(gotQuery, "cursor=abc") {
		t.Errorf("query = %q", gotQuery)
	}
}

func TestChatsCreate(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/chats" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		var body CreateChatRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.From != "+15551234567" || len(body.To) != 1 {
			t.Errorf("body = %+v", body)
		}
		if len(body.Message.Parts) != 1 || body.Message.Parts[0].Text == nil {
			t.Errorf("parts = %+v", body.Message.Parts)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"chat":{"id":"c1","is_archived":false,"is_group":false,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"},"message":{"id":"m1","parts":[],"created_at":"2026-01-01T00:00:00Z","delivery_status":"queued","is_read":false}}`))
	})
	res, err := c.Chats.Create(context.Background(), &CreateChatRequest{
		From:    "+15551234567",
		To:      []string{"+15557654321"},
		Message: CreateChatMessage{Parts: []MessagePart{NewTextPart("hi")}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Chat.ID != "c1" || res.Message.ID != "m1" {
		t.Errorf("res = %+v", res)
	}
}

func TestChatsParticipants(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/chats/c1/participants" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["handle"] != "+15557654321" {
			t.Errorf("handle = %q", body["handle"])
		}
		w.WriteHeader(http.StatusAccepted)
	})
	if err := c.Chats.AddParticipant(context.Background(), "c1", "+15557654321"); err != nil {
		t.Fatal(err)
	}
	if err := c.Chats.RemoveParticipant(context.Background(), "c1", "+15557654321"); err != nil {
		t.Fatal(err)
	}
}

func TestChatsTypingAndRead(t *testing.T) {
	methods := []struct {
		method string
		path   string
		call   func(c *Client) error
	}{
		{http.MethodPost, "/v3/chats/c1/typing", func(c *Client) error { return c.Chats.StartTyping(context.Background(), "c1") }},
		{http.MethodDelete, "/v3/chats/c1/typing", func(c *Client) error { return c.Chats.StopTyping(context.Background(), "c1") }},
		{http.MethodPost, "/v3/chats/c1/read", func(c *Client) error { return c.Chats.MarkRead(context.Background(), "c1") }},
		{http.MethodPost, "/v3/chats/c1/leave", func(c *Client) error { return c.Chats.Leave(context.Background(), "c1") }},
		{http.MethodPost, "/v3/chats/c1/share_contact_card", func(c *Client) error { return c.Chats.ShareContactCard(context.Background(), "c1") }},
	}
	for _, tc := range methods {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tc.method || r.URL.Path != tc.path {
					t.Errorf("got %s %s", r.Method, r.URL.Path)
				}
				w.WriteHeader(http.StatusNoContent)
			})
			if err := tc.call(c); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestChatsUpdate(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/v3/chats/c1" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		var body UpdateChatRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.DisplayName == nil || *body.DisplayName != "New Name" {
			t.Errorf("display_name = %v", body.DisplayName)
		}
		w.WriteHeader(http.StatusAccepted)
	})
	name := "New Name"
	if err := c.Chats.Update(context.Background(), "c1", &UpdateChatRequest{DisplayName: &name}); err != nil {
		t.Fatal(err)
	}
}
