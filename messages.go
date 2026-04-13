package linq

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// MessagesService groups message endpoints.
type MessagesService struct{ c *Client }

// SendMessageRequest is the body for sending a message into a chat.
type SendMessageRequest struct {
	Parts            []MessagePart  `json:"parts"`
	Effect           *MessageEffect `json:"effect,omitempty"`
	PreferredService *ServiceType   `json:"preferred_service,omitempty"`
	ReplyTo          *ReplyTo       `json:"reply_to,omitempty"`
}

// Send sends a message to the given chat.
func (s *MessagesService) Send(ctx context.Context, chatID string, req *SendMessageRequest) (*Message, error) {
	var out Message
	if err := s.c.do(ctx, http.MethodPost, "/v3/chats/"+chatID+"/messages", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListMessagesParams filters the List call.
type ListMessagesParams struct {
	Limit  int
	Cursor string
}

// ListMessagesResult is a paginated page of messages.
type ListMessagesResult struct {
	Messages   []Message `json:"messages"`
	NextCursor string    `json:"next_cursor,omitempty"`
}

// List returns messages in a chat.
func (s *MessagesService) List(ctx context.Context, chatID string, p *ListMessagesParams) (*ListMessagesResult, error) {
	q := url.Values{}
	if p != nil {
		if p.Limit > 0 {
			q.Set("limit", strconv.Itoa(p.Limit))
		}
		if p.Cursor != "" {
			q.Set("cursor", p.Cursor)
		}
	}
	var out ListMessagesResult
	if err := s.c.do(ctx, http.MethodGet, "/v3/chats/"+chatID+"/messages", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a single message by ID.
func (s *MessagesService) Get(ctx context.Context, messageID string) (*Message, error) {
	var out Message
	if err := s.c.do(ctx, http.MethodGet, "/v3/messages/"+messageID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete deletes a message.
func (s *MessagesService) Delete(ctx context.Context, messageID string) error {
	return s.c.do(ctx, http.MethodDelete, "/v3/messages/"+messageID, nil, nil, nil)
}

// ThreadResult is the threaded-reply list for a given message.
type ThreadResult struct {
	Messages []Message `json:"messages"`
}

// Thread fetches the threaded reply chain for a message.
func (s *MessagesService) Thread(ctx context.Context, messageID string) (*ThreadResult, error) {
	var out ThreadResult
	if err := s.c.do(ctx, http.MethodGet, "/v3/messages/"+messageID+"/thread", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
