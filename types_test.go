package linq

import (
	"encoding/json"
	"testing"
)

func TestMessagePartRoundTripText(t *testing.T) {
	original := NewTextPart("hello")
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded MessagePart
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Text == nil || decoded.Text.Value != "hello" {
		t.Errorf("round-trip failed: %+v", decoded)
	}
}

func TestMessagePartRoundTripMedia(t *testing.T) {
	original := NewMediaPartByID("att-1")
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(b), `"type":"media"`) {
		t.Errorf("missing type discriminator: %s", b)
	}
	var decoded MessagePart
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Media == nil || decoded.Media.AttachmentID == nil || *decoded.Media.AttachmentID != "att-1" {
		t.Errorf("round-trip failed: %+v", decoded)
	}
}

func TestMessagePartRoundTripLink(t *testing.T) {
	original := NewLinkPart("https://example.com")
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded MessagePart
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Link == nil || decoded.Link.Value != "https://example.com" {
		t.Errorf("round-trip failed: %+v", decoded)
	}
}

func TestMessagePartUnknownType(t *testing.T) {
	var p MessagePart
	if err := json.Unmarshal([]byte(`{"type":"bogus"}`), &p); err == nil {
		t.Error("expected error on unknown type")
	}
}

func TestMessagePartEmptyMarshal(t *testing.T) {
	var p MessagePart
	if _, err := json.Marshal(p); err == nil {
		t.Error("expected error marshalling empty part")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
