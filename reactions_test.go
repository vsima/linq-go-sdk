package linq

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestReactionsSet(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/messages/m1/reactions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var body ReactionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Type != ReactionCustom {
			t.Errorf("type = %q", body.Type)
		}
		if body.CustomEmoji == nil || *body.CustomEmoji != "🔥" {
			t.Errorf("emoji = %v", body.CustomEmoji)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	emoji := "🔥"
	if err := c.Reactions.Set(context.Background(), "m1", &ReactionRequest{Type: ReactionCustom, CustomEmoji: &emoji}); err != nil {
		t.Fatal(err)
	}
}
