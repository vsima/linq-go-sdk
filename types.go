package linq

import (
	"encoding/json"
	"fmt"
	"time"
)

// ServiceType identifies the delivery protocol for a chat or message.
type ServiceType string

const (
	ServiceIMessage ServiceType = "iMessage"
	ServiceSMS      ServiceType = "SMS"
	ServiceRCS      ServiceType = "RCS"
)

// DeliveryStatus is the lifecycle state of an outbound message.
type DeliveryStatus string

const (
	DeliveryPending   DeliveryStatus = "pending"
	DeliveryQueued    DeliveryStatus = "queued"
	DeliverySent      DeliveryStatus = "sent"
	DeliveryDelivered DeliveryStatus = "delivered"
	DeliveryFailed    DeliveryStatus = "failed"
)

// HandleStatus is the membership state of a participant in a chat.
type HandleStatus string

const (
	HandleActive  HandleStatus = "active"
	HandleLeft    HandleStatus = "left"
	HandleRemoved HandleStatus = "removed"
)

// ReactionType enumerates built-in reaction kinds. Use ReactionCustom with
// CustomEmoji, or ReactionSticker with a Sticker payload.
type ReactionType string

const (
	ReactionLove      ReactionType = "love"
	ReactionLike      ReactionType = "like"
	ReactionDislike   ReactionType = "dislike"
	ReactionLaugh     ReactionType = "laugh"
	ReactionEmphasize ReactionType = "emphasize"
	ReactionQuestion  ReactionType = "question"
	ReactionCustom    ReactionType = "custom"
	ReactionSticker   ReactionType = "sticker"
)

// Chat is a conversation.
type Chat struct {
	ID          string       `json:"id"`
	DisplayName *string      `json:"display_name,omitempty"`
	Service     *ServiceType `json:"service,omitempty"`
	Handles     []ChatHandle `json:"handles,omitempty"`
	IsArchived  bool         `json:"is_archived"`
	IsGroup     bool         `json:"is_group"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// ChatHandle identifies a participant in a chat.
type ChatHandle struct {
	ID       string        `json:"id"`
	Handle   string        `json:"handle"`
	Service  ServiceType   `json:"service"`
	Status   *HandleStatus `json:"status,omitempty"`
	JoinedAt time.Time     `json:"joined_at"`
	LeftAt   *time.Time    `json:"left_at,omitempty"`
	IsMe     *bool         `json:"is_me,omitempty"`
}

// Message is a chat message.
type Message struct {
	ID               string         `json:"id"`
	Service          *ServiceType   `json:"service,omitempty"`
	PreferredService *ServiceType   `json:"preferred_service,omitempty"`
	Parts            []MessagePart  `json:"parts"`
	CreatedAt        time.Time      `json:"created_at"`
	SentAt           *time.Time     `json:"sent_at,omitempty"`
	DeliveredAt      *time.Time     `json:"delivered_at,omitempty"`
	DeliveryStatus   DeliveryStatus `json:"delivery_status"`
	IsRead           bool           `json:"is_read"`
	Effect           *MessageEffect `json:"effect,omitempty"`
	FromHandle       *ChatHandle    `json:"from_handle,omitempty"`
	ReplyTo          *ReplyTo       `json:"reply_to,omitempty"`
}

// MessageEffect is a screen or bubble effect applied to a message.
// Type must be "screen" or "bubble".
type MessageEffect struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// ReplyTo targets a message (and optional part index) for threaded replies.
type ReplyTo struct {
	MessageID string `json:"message_id"`
	PartIndex int    `json:"part_index,omitempty"`
}

// Reaction is a reaction attached to a message part.
type Reaction struct {
	IsMe        bool         `json:"is_me"`
	Handle      ChatHandle   `json:"handle"`
	Type        ReactionType `json:"type"`
	CustomEmoji *string      `json:"custom_emoji,omitempty"`
	Sticker     *Sticker     `json:"sticker,omitempty"`
}

// Sticker is a custom sticker used in a reaction.
type Sticker struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileName string `json:"file_name"`
}

// MessagePart is a discriminated union: Text, Media, or Link.
// Exactly one of the pointer fields on the wrapper is set after unmarshal.
type MessagePart struct {
	Text  *TextPart
	Media *MediaPart
	Link  *LinkPart
}

// TextPart is a textual segment of a message.
type TextPart struct {
	Value           string           `json:"value"`
	TextDecorations []TextDecoration `json:"text_decorations,omitempty"`
	Reactions       []Reaction       `json:"reactions,omitempty"`
}

// MediaPart references an attachment by URL or attachment_id.
type MediaPart struct {
	URL          *string    `json:"url,omitempty"`
	AttachmentID *string    `json:"attachment_id,omitempty"`
	Reactions    []Reaction `json:"reactions,omitempty"`
}

// LinkPart is a link preview segment.
type LinkPart struct {
	Value     string     `json:"value"`
	Reactions []Reaction `json:"reactions,omitempty"`
}

// TextDecoration applies a style or animation to a character range [Start,End).
// Set Style or Animation, not both.
type TextDecoration struct {
	Range     [2]int  `json:"range"`
	Style     *string `json:"style,omitempty"`
	Animation *string `json:"animation,omitempty"`
}

// MarshalJSON encodes the active variant with its type discriminator.
func (p MessagePart) MarshalJSON() ([]byte, error) {
	switch {
	case p.Text != nil:
		return json.Marshal(struct {
			Type string `json:"type"`
			*TextPart
		}{"text", p.Text})
	case p.Media != nil:
		return json.Marshal(struct {
			Type string `json:"type"`
			*MediaPart
		}{"media", p.Media})
	case p.Link != nil:
		return json.Marshal(struct {
			Type string `json:"type"`
			*LinkPart
		}{"link", p.Link})
	}
	return nil, fmt.Errorf("linq: empty MessagePart")
}

// UnmarshalJSON decodes a tagged part and populates the matching variant.
func (p *MessagePart) UnmarshalJSON(data []byte) error {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return err
	}
	switch head.Type {
	case "text":
		var t TextPart
		if err := json.Unmarshal(data, &t); err != nil {
			return err
		}
		p.Text = &t
	case "media":
		var m MediaPart
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		p.Media = &m
	case "link":
		var l LinkPart
		if err := json.Unmarshal(data, &l); err != nil {
			return err
		}
		p.Link = &l
	default:
		return fmt.Errorf("linq: unknown MessagePart type %q", head.Type)
	}
	return nil
}

// NewTextPart is a convenience for building a text MessagePart.
func NewTextPart(value string) MessagePart {
	return MessagePart{Text: &TextPart{Value: value}}
}

// NewMediaPartByID builds a media MessagePart from an uploaded attachment.
func NewMediaPartByID(attachmentID string) MessagePart {
	return MessagePart{Media: &MediaPart{AttachmentID: &attachmentID}}
}

// NewMediaPartByURL builds a media MessagePart from an HTTPS URL.
func NewMediaPartByURL(url string) MessagePart {
	return MessagePart{Media: &MediaPart{URL: &url}}
}

// NewLinkPart builds a link MessagePart.
func NewLinkPart(url string) MessagePart {
	return MessagePart{Link: &LinkPart{Value: url}}
}
