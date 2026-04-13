package linq

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// WebhooksService groups webhook-subscription endpoints and event parsing.
type WebhooksService struct{ c *Client }

// WebhookSubscription describes a webhook endpoint and its subscribed events.
type WebhookSubscription struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Events      []string  `json:"events"`
	Secret      string    `json:"secret,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Description string    `json:"description,omitempty"`
}

// ListWebhookSubscriptionsResult is a page of subscriptions.
type ListWebhookSubscriptionsResult struct {
	Subscriptions []WebhookSubscription `json:"subscriptions"`
}

// List returns all webhook subscriptions.
func (s *WebhooksService) List(ctx context.Context) (*ListWebhookSubscriptionsResult, error) {
	var out ListWebhookSubscriptionsResult
	if err := s.c.do(ctx, http.MethodGet, "/v3/webhook-subscriptions", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateWebhookSubscriptionRequest creates a new webhook subscription.
type CreateWebhookSubscriptionRequest struct {
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Description string   `json:"description,omitempty"`
}

// Create registers a new webhook subscription.
func (s *WebhooksService) Create(ctx context.Context, req *CreateWebhookSubscriptionRequest) (*WebhookSubscription, error) {
	var out WebhookSubscription
	if err := s.c.do(ctx, http.MethodPost, "/v3/webhook-subscriptions", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateWebhookSubscriptionRequest updates mutable fields of a subscription.
type UpdateWebhookSubscriptionRequest struct {
	URL         *string  `json:"url,omitempty"`
	Events      []string `json:"events,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
	Description *string  `json:"description,omitempty"`
}

// Update modifies a webhook subscription.
func (s *WebhooksService) Update(ctx context.Context, id string, req *UpdateWebhookSubscriptionRequest) (*WebhookSubscription, error) {
	var out WebhookSubscription
	if err := s.c.do(ctx, http.MethodPut, "/v3/webhook-subscriptions/"+id, nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a webhook subscription.
func (s *WebhooksService) Delete(ctx context.Context, id string) error {
	return s.c.do(ctx, http.MethodDelete, "/v3/webhook-subscriptions/"+id, nil, nil, nil)
}

// EventType enumerates webhook event types.
type EventType string

const (
	EventMessageSent               EventType = "message.sent"
	EventMessageDelivered          EventType = "message.delivered"
	EventMessageFailed             EventType = "message.failed"
	EventMessageReceived           EventType = "message.received"
	EventMessageRead               EventType = "message.read"
	EventReactionSent              EventType = "reaction.sent"
	EventReactionReceived          EventType = "reaction.received"
	EventTypingIndicatorReceived   EventType = "typing_indicator.received"
	EventTypingIndicatorRemoved    EventType = "typing_indicator.removed"
	EventParticipantAdded          EventType = "participant.added"
	EventParticipantRemoved        EventType = "participant.removed"
	EventChatGroupNameUpdated      EventType = "chat.group_name_updated"
	EventChatGroupIconUpdated      EventType = "chat.group_icon_updated"
	EventChatGroupNameUpdateFailed EventType = "chat.group_name_update_failed"
	EventChatGroupIconUpdateFailed EventType = "chat.group_icon_update_failed"
)

// Event is a webhook envelope. Data holds the raw event payload; decode it
// with DecodeData into the matching concrete type for the EventType.
type Event struct {
	APIVersion string          `json:"api_version"`
	EventType  EventType       `json:"event_type"`
	EventID    string          `json:"event_id"`
	CreatedAt  time.Time       `json:"created_at"`
	TraceID    string          `json:"trace_id"`
	PartnerID  string          `json:"partner_id"`
	Data       json.RawMessage `json:"data"`
}

// DecodeData unmarshals Event.Data into out.
func (e *Event) DecodeData(out any) error {
	return json.Unmarshal(e.Data, out)
}

// ParseEvent decodes a webhook request body into an Event.
func ParseEvent(body []byte) (*Event, error) {
	var e Event
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// Webhook header names set by the Linq signing process.
const (
	HeaderWebhookTimestamp = "X-Webhook-Timestamp"
	HeaderWebhookSignature = "X-Webhook-Signature"
)

// DefaultWebhookTolerance is the default maximum age accepted by
// [VerifyWebhook]. Requests older than this are rejected as possible replays.
const DefaultWebhookTolerance = 5 * time.Minute

// Errors returned by [VerifyWebhook] and [VerifyWebhookRequest].
var (
	ErrWebhookMissingHeader    = errors.New("linq: missing webhook signature header")
	ErrWebhookInvalidTimestamp = errors.New("linq: invalid webhook timestamp")
	ErrWebhookStale            = errors.New("linq: webhook timestamp outside tolerance")
	ErrWebhookSignature        = errors.New("linq: webhook signature mismatch")
)

// VerifyWebhook verifies an HMAC-SHA256 webhook signature as specified by Linq.
//
// The signed payload is "{timestamp}.{body}" where body is the raw request
// bytes (do not re-serialize JSON before calling). signature is the value of
// the [HeaderWebhookSignature] header, hex-encoded.
//
// If tolerance is > 0, requests with a timestamp older than tolerance (or more
// than tolerance in the future) are rejected with [ErrWebhookStale].
// Pass 0 to disable the freshness check — not recommended in production.
func VerifyWebhook(body []byte, timestamp, signature, secret string, tolerance time.Duration) error {
	if timestamp == "" || signature == "" {
		return ErrWebhookMissingHeader
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWebhookInvalidTimestamp, err)
	}
	if tolerance > 0 {
		age := time.Since(time.Unix(ts, 0))
		if age < 0 {
			age = -age
		}
		if age > tolerance {
			return ErrWebhookStale
		}
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte{'.'})
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return ErrWebhookSignature
	}
	return nil
}

// VerifyWebhookRequest reads the full body of r, verifies its signature using
// [VerifyWebhook], and returns the raw body bytes for downstream parsing.
//
// On success the caller can pass the returned bytes to [ParseEvent]. The body
// of r is consumed; re-reading r.Body after this call will yield nothing.
func VerifyWebhookRequest(r *http.Request, secret string, tolerance time.Duration) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("linq: read webhook body: %w", err)
	}
	ts := r.Header.Get(HeaderWebhookTimestamp)
	sig := r.Header.Get(HeaderWebhookSignature)
	if err := VerifyWebhook(body, ts, sig, secret, tolerance); err != nil {
		return nil, err
	}
	return body, nil
}
