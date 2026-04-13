package linq

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ChatsService groups chat endpoints.
type ChatsService struct{ c *Client }

// ListChatsParams filters the List call.
type ListChatsParams struct {
	From   string // E.164 phone number
	To     string // participant handle (phone or email)
	Limit  int    // 1-100
	Cursor string
}

// ListChatsResult is a paginated page of chats.
type ListChatsResult struct {
	Chats      []Chat `json:"chats"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// List returns a paginated set of chats.
func (s *ChatsService) List(ctx context.Context, p *ListChatsParams) (*ListChatsResult, error) {
	q := url.Values{}
	if p != nil {
		if p.From != "" {
			q.Set("from", p.From)
		}
		if p.To != "" {
			q.Set("to", p.To)
		}
		if p.Limit > 0 {
			q.Set("limit", strconv.Itoa(p.Limit))
		}
		if p.Cursor != "" {
			q.Set("cursor", p.Cursor)
		}
	}
	var out ListChatsResult
	if err := s.c.do(ctx, http.MethodGet, "/v3/chats", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateChatMessage is the initial message sent when a chat is created.
type CreateChatMessage struct {
	Parts            []MessagePart  `json:"parts"`
	Effect           *MessageEffect `json:"effect,omitempty"`
	PreferredService *ServiceType   `json:"preferred_service,omitempty"`
	ReplyTo          *ReplyTo       `json:"reply_to,omitempty"`
}

// CreateChatRequest is the body for creating a chat with an initial message.
type CreateChatRequest struct {
	From    string            `json:"from"`
	To      []string          `json:"to"`
	Message CreateChatMessage `json:"message"`
}

// CreateChatResult is the 201 response from creating a chat.
//
// The initial message lives at Chat.Message, not at the top level.
type CreateChatResult struct {
	Chat Chat `json:"chat"`
}

// Create creates a new chat with an initial message.
func (s *ChatsService) Create(ctx context.Context, req *CreateChatRequest) (*CreateChatResult, error) {
	var out CreateChatResult
	if err := s.c.do(ctx, http.MethodPost, "/v3/chats", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a chat by ID.
func (s *ChatsService) Get(ctx context.Context, chatID string) (*Chat, error) {
	var out Chat
	if err := s.c.do(ctx, http.MethodGet, "/v3/chats/"+chatID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateChatRequest updates mutable properties of a chat.
type UpdateChatRequest struct {
	DisplayName   *string `json:"display_name,omitempty"`
	GroupChatIcon *string `json:"group_chat_icon,omitempty"`
}

// Update updates the display name or icon of a group chat.
func (s *ChatsService) Update(ctx context.Context, chatID string, req *UpdateChatRequest) error {
	return s.c.do(ctx, http.MethodPut, "/v3/chats/"+chatID, nil, req, nil)
}

// AddParticipant adds a handle (phone or email) to a group chat.
func (s *ChatsService) AddParticipant(ctx context.Context, chatID, handle string) error {
	body := map[string]string{"handle": handle}
	return s.c.do(ctx, http.MethodPost, "/v3/chats/"+chatID+"/participants", nil, body, nil)
}

// RemoveParticipant removes a handle from a group chat.
func (s *ChatsService) RemoveParticipant(ctx context.Context, chatID, handle string) error {
	body := map[string]string{"handle": handle}
	return s.c.do(ctx, http.MethodDelete, "/v3/chats/"+chatID+"/participants", nil, body, nil)
}

// Leave leaves a group chat.
func (s *ChatsService) Leave(ctx context.Context, chatID string) error {
	return s.c.do(ctx, http.MethodPost, "/v3/chats/"+chatID+"/leave", nil, nil, nil)
}

// StartTyping sends a typing-started indicator.
func (s *ChatsService) StartTyping(ctx context.Context, chatID string) error {
	return s.c.do(ctx, http.MethodPost, "/v3/chats/"+chatID+"/typing", nil, nil, nil)
}

// StopTyping sends a typing-stopped indicator.
func (s *ChatsService) StopTyping(ctx context.Context, chatID string) error {
	return s.c.do(ctx, http.MethodDelete, "/v3/chats/"+chatID+"/typing", nil, nil, nil)
}

// MarkRead marks all messages in the chat as read.
func (s *ChatsService) MarkRead(ctx context.Context, chatID string) error {
	return s.c.do(ctx, http.MethodPost, "/v3/chats/"+chatID+"/read", nil, nil, nil)
}

// ShareContactCard shares the authenticated partner's contact card in the chat.
func (s *ChatsService) ShareContactCard(ctx context.Context, chatID string) error {
	return s.c.do(ctx, http.MethodPost, "/v3/chats/"+chatID+"/share_contact_card", nil, nil, nil)
}

// SendVoiceMemo sends a voice memo to a chat. The attachment must be pre-uploaded.
func (s *ChatsService) SendVoiceMemo(ctx context.Context, chatID string, req *SendVoiceMemoRequest) (*Message, error) {
	var out Message
	path := fmt.Sprintf("/v3/chats/%s/voicememo", chatID)
	if err := s.c.do(ctx, http.MethodPost, path, nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendVoiceMemoRequest is the body for SendVoiceMemo.
type SendVoiceMemoRequest struct {
	AttachmentID     string       `json:"attachment_id"`
	PreferredService *ServiceType `json:"preferred_service,omitempty"`
}
