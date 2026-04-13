package linq

import (
	"context"
	"net/http"
)

// PhoneNumbersService groups phone-number endpoints.
type PhoneNumbersService struct{ c *Client }

// PhoneNumber is a sender phone number associated with the partner token.
type PhoneNumber struct {
	Number    string   `json:"number"`
	Services  []string `json:"services,omitempty"`
	Label     string   `json:"label,omitempty"`
	IsDefault bool     `json:"is_default,omitempty"`
}

// ListPhoneNumbersResult is the response for List.
type ListPhoneNumbersResult struct {
	PhoneNumbers []PhoneNumber `json:"phone_numbers"`
}

// List returns the phone numbers the partner token may send from.
func (s *PhoneNumbersService) List(ctx context.Context) (*ListPhoneNumbersResult, error) {
	var out ListPhoneNumbersResult
	if err := s.c.do(ctx, http.MethodGet, "/v3/phonenumbers", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
