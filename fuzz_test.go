package linq

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"testing"
	"time"
)

func FuzzParseEvent(f *testing.F) {
	f.Add([]byte(`{"api_version":"v3","event_type":"message.sent","event_id":"x","created_at":"2026-01-01T00:00:00Z","trace_id":"t","partner_id":"p","data":{}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"data":null}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		evt, err := ParseEvent(data)
		if err != nil {
			return
		}
		// A successful parse should never return a nil envelope.
		if evt == nil {
			t.Fatal("ParseEvent: nil event with nil error")
		}
		// DecodeData on an event with empty data should be a no-op or error,
		// but must never panic.
		var v any
		_ = evt.DecodeData(&v)
	})
}

func FuzzVerifyWebhook(f *testing.F) {
	f.Add([]byte(`{"ok":true}`), "1712000000", "deadbeef", "sekret")
	f.Add([]byte{}, "", "", "")
	f.Add([]byte("\x00\x01\x02"), "not-a-number", "zz", "s")

	f.Fuzz(func(t *testing.T, body []byte, ts, sig, secret string) {
		// Must never panic, whatever random garbage we throw in.
		_ = VerifyWebhook(body, ts, sig, secret, time.Minute)

		// Roundtrip: a correctly-signed payload must always verify.
		if secret == "" {
			return
		}
		ts2 := strconv.FormatInt(time.Now().Unix(), 10)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(ts2))
		mac.Write([]byte{'.'})
		mac.Write(body)
		good := hex.EncodeToString(mac.Sum(nil))
		if err := VerifyWebhook(body, ts2, good, secret, time.Minute); err != nil {
			t.Fatalf("correctly-signed payload failed: %v", err)
		}
	})
}

func FuzzMessagePart(f *testing.F) {
	f.Add([]byte(`{"type":"text","value":"hi"}`))
	f.Add([]byte(`{"type":"media","attachment_id":"a"}`))
	f.Add([]byte(`{"type":"link","value":"https://x"}`))
	f.Add([]byte(`{"type":"bogus"}`))
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var p MessagePart
		if err := json.Unmarshal(data, &p); err != nil {
			return
		}
		// A successful unmarshal must produce a marshalable value, and the
		// round-trip must succeed.
		b, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("marshal after successful unmarshal: %v", err)
		}
		var p2 MessagePart
		if err := json.Unmarshal(b, &p2); err != nil {
			t.Fatalf("re-unmarshal: %v (bytes=%s)", err, b)
		}
	})
}
