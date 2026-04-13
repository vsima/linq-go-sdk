package linq

import (
	"context"
	"encoding/json"
	"net/http"
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
	EventMessageSent              EventType = "message.sent"
	EventMessageDelivered         EventType = "message.delivered"
	EventMessageFailed            EventType = "message.failed"
	EventMessageReceived          EventType = "message.received"
	EventMessageRead              EventType = "message.read"
	EventReactionSent             EventType = "reaction.sent"
	EventReactionReceived         EventType = "reaction.received"
	EventTypingIndicatorReceived  EventType = "typing_indicator.received"
	EventTypingIndicatorRemoved   EventType = "typing_indicator.removed"
	EventParticipantAdded         EventType = "participant.added"
	EventParticipantRemoved       EventType = "participant.removed"
	EventChatGroupNameUpdated     EventType = "chat.group_name_updated"
	EventChatGroupIconUpdated     EventType = "chat.group_icon_updated"
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
