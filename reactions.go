package linq

import (
	"context"
	"net/http"
)

// ReactionsService groups reaction endpoints.
type ReactionsService struct{ c *Client }

// ReactionRequest adds or removes a reaction on a message part.
// Set Remove to true to remove; otherwise the reaction is added.
type ReactionRequest struct {
	Type        ReactionType `json:"type"`
	PartIndex   int          `json:"part_index,omitempty"`
	CustomEmoji *string      `json:"custom_emoji,omitempty"`
	Sticker     *Sticker     `json:"sticker,omitempty"`
	Remove      bool         `json:"remove,omitempty"`
}

// Set adds or removes a reaction on a message.
func (s *ReactionsService) Set(ctx context.Context, messageID string, req *ReactionRequest) error {
	return s.c.do(ctx, http.MethodPost, "/v3/messages/"+messageID+"/reactions", nil, req, nil)
}
