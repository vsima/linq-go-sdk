package linq

import (
	"testing"
)

func TestParseEvent(t *testing.T) {
	body := []byte(`{
	  "api_version":"v3",
	  "event_type":"message.received",
	  "event_id":"evt-1",
	  "created_at":"2026-01-01T00:00:00Z",
	  "trace_id":"trace-1",
	  "partner_id":"p-1",
	  "data":{"id":"m-1","parts":[{"type":"text","value":"hi"}],"created_at":"2026-01-01T00:00:00Z","delivery_status":"delivered","is_read":false}
	}`)

	evt, err := ParseEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	if evt.EventType != EventMessageReceived {
		t.Errorf("EventType = %q", evt.EventType)
	}
	if evt.TraceID != "trace-1" {
		t.Errorf("TraceID = %q", evt.TraceID)
	}
	var msg Message
	if err := evt.DecodeData(&msg); err != nil {
		t.Fatal(err)
	}
	if msg.ID != "m-1" || len(msg.Parts) != 1 || msg.Parts[0].Text.Value != "hi" {
		t.Errorf("decoded message = %+v", msg)
	}
}

func TestParseEventInvalid(t *testing.T) {
	if _, err := ParseEvent([]byte("{not json")); err == nil {
		t.Error("expected error")
	}
}
