package linq

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func signPayload(timestamp string, body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte{'.'})
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookValid(t *testing.T) {
	body := []byte(`{"event_type":"message.sent"}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	secret := "s3cret"
	sig := signPayload(ts, body, secret)

	if err := VerifyWebhook(body, ts, sig, secret, DefaultWebhookTolerance); err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
}

func TestVerifyWebhookMissingHeader(t *testing.T) {
	err := VerifyWebhook([]byte("x"), "", "", "s", time.Minute)
	if !errors.Is(err, ErrWebhookMissingHeader) {
		t.Errorf("err = %v", err)
	}
}

func TestVerifyWebhookBadTimestamp(t *testing.T) {
	err := VerifyWebhook([]byte("x"), "not-a-number", "abc", "s", time.Minute)
	if !errors.Is(err, ErrWebhookInvalidTimestamp) {
		t.Errorf("err = %v", err)
	}
}

func TestVerifyWebhookStale(t *testing.T) {
	body := []byte("x")
	ts := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	sig := signPayload(ts, body, "s")
	err := VerifyWebhook(body, ts, sig, "s", 5*time.Minute)
	if !errors.Is(err, ErrWebhookStale) {
		t.Errorf("err = %v", err)
	}
}

func TestVerifyWebhookToleranceDisabled(t *testing.T) {
	body := []byte("x")
	ts := strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10)
	sig := signPayload(ts, body, "s")
	if err := VerifyWebhook(body, ts, sig, "s", 0); err != nil {
		t.Errorf("expected success with tolerance=0: %v", err)
	}
}

func TestVerifyWebhookSignatureMismatch(t *testing.T) {
	body := []byte("x")
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := signPayload(ts, body, "s")
	err := VerifyWebhook(body, ts, sig, "wrong-secret", time.Minute)
	if !errors.Is(err, ErrWebhookSignature) {
		t.Errorf("err = %v", err)
	}
}

func TestVerifyWebhookBodyTampered(t *testing.T) {
	body := []byte(`{"event_type":"message.sent"}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := signPayload(ts, body, "s")
	tampered := []byte(`{"event_type":"message.received"}`)
	err := VerifyWebhook(tampered, ts, sig, "s", time.Minute)
	if !errors.Is(err, ErrWebhookSignature) {
		t.Errorf("err = %v", err)
	}
}

func TestVerifyWebhookRequest(t *testing.T) {
	secret := "s3cret"
	body := []byte(`{"event_type":"message.sent","event_id":"evt-1"}`)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := signPayload(ts, body, secret)

	r := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewReader(body))
	r.Header.Set(HeaderWebhookTimestamp, ts)
	r.Header.Set(HeaderWebhookSignature, sig)

	got, err := VerifyWebhookRequest(r, secret, DefaultWebhookTolerance)
	if err != nil {
		t.Fatalf("VerifyWebhookRequest: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("returned body mismatch")
	}
}

func ExampleVerifyWebhookRequest() {
	// Inside an http.HandlerFunc:
	//
	//   body, err := linq.VerifyWebhookRequest(r, secret, linq.DefaultWebhookTolerance)
	//   if err != nil {
	//       http.Error(w, err.Error(), http.StatusUnauthorized)
	//       return
	//   }
	//   evt, _ := linq.ParseEvent(body)
	//   _ = evt
	fmt.Println("ok")
	// Output: ok
}
