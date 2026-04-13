// Package linq is a Go client for the Linq Partner API V3.
//
// See https://apidocs.linqapp.com/ for API reference.
package linq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultBaseURL   = "https://api.linqapp.com/api/partner"
	DefaultUserAgent = "linq-go-sdk/0.1"
)

// Client is the Linq Partner API client.
type Client struct {
	token      string
	baseURL    string
	userAgent  string
	httpClient *http.Client

	Chats        *ChatsService
	Messages     *MessagesService
	Reactions    *ReactionsService
	Attachments  *AttachmentsService
	PhoneNumbers *PhoneNumbersService
	Webhooks     *WebhooksService
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL overrides the API base URL (no trailing slash).
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithHTTPClient overrides the underlying HTTP client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// NewClient returns a Client authenticated with the given bearer token.
func NewClient(token string, opts ...Option) *Client {
	c := &Client{
		token:      token,
		baseURL:    DefaultBaseURL,
		userAgent:  DefaultUserAgent,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.Chats = &ChatsService{c: c}
	c.Messages = &MessagesService{c: c}
	c.Reactions = &ReactionsService{c: c}
	c.Attachments = &AttachmentsService{c: c}
	c.PhoneNumbers = &PhoneNumbersService{c: c}
	c.Webhooks = &WebhooksService{c: c}
	return c
}

// do executes a request. If out is non-nil and the response has a JSON body,
// it is decoded into out. 204/empty responses are handled.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("linq: marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return fmt.Errorf("linq: new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("linq: http: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("linq: read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseError(resp, data)
	}
	if len(data) == 0 || out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("linq: decode response: %w", err)
	}
	return nil
}
