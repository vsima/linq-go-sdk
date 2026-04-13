package linq

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode is the Linq-specific error code returned in the error body.
type ErrorCode int

const (
	ErrCodeServer           ErrorCode = 1001
	ErrCodeUnauthorized     ErrorCode = 1002
	ErrCodeNotFound         ErrorCode = 1004
	ErrCodeInvalidParameter ErrorCode = 1005
	ErrCodeRateLimited      ErrorCode = 1007
)

// APIError represents a non-2xx response from the Linq API.
type APIError struct {
	StatusCode int       `json:"-"`
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	RetryAfter int       `json:"retry_after,omitempty"`
	TraceID    string    `json:"-"`
	RequestID  string    `json:"-"`
	Raw        []byte    `json:"-"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("linq: %d %s (code=%d)", e.StatusCode, e.Message, e.Code)
	}
	return fmt.Sprintf("linq: http %d", e.StatusCode)
}

// IsNotFound reports whether err is a 404 from the API.
func IsNotFound(err error) bool { return hasStatus(err, http.StatusNotFound) }

// IsUnauthorized reports whether err is a 401 from the API.
func IsUnauthorized(err error) bool { return hasStatus(err, http.StatusUnauthorized) }

// IsRateLimited reports whether err is a 429 from the API.
func IsRateLimited(err error) bool { return hasStatus(err, http.StatusTooManyRequests) }

func hasStatus(err error, code int) bool {
	var ae *APIError
	return errors.As(err, &ae) && ae.StatusCode == code
}

func parseError(resp *http.Response, body []byte) error {
	ae := &APIError{
		StatusCode: resp.StatusCode,
		TraceID:    resp.Header.Get("X-Trace-Id"),
		RequestID:  resp.Header.Get("X-Request-Id"),
		Raw:        body,
	}
	var wrapper struct {
		Success bool      `json:"success"`
		Error   *APIError `json:"error"`
		TraceID string    `json:"trace_id"`
	}
	if err := json.Unmarshal(body, &wrapper); err == nil && wrapper.Error != nil {
		ae.Code = wrapper.Error.Code
		ae.Message = wrapper.Error.Message
		ae.RetryAfter = wrapper.Error.RetryAfter
		if wrapper.TraceID != "" {
			ae.TraceID = wrapper.TraceID
		}
	}
	return ae
}
