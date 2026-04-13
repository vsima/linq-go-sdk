package linq

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AttachmentsService groups attachment endpoints.
type AttachmentsService struct{ c *Client }

// CreateAttachmentRequest requests a presigned upload URL.
type CreateAttachmentRequest struct {
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size,omitempty"`
}

// Attachment is the presigned-upload response.
type Attachment struct {
	AttachmentID string    `json:"attachment_id"`
	UploadURL    string    `json:"upload_url"`
	DownloadURL  string    `json:"download_url"`
	ExpiresAt    time.Time `json:"expires_at"`
	Status       string    `json:"status,omitempty"`
}

// Create requests a presigned URL for uploading an attachment.
func (s *AttachmentsService) Create(ctx context.Context, req *CreateAttachmentRequest) (*Attachment, error) {
	var out Attachment
	if err := s.c.do(ctx, http.MethodPost, "/v3/attachments", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns the current status of an attachment.
func (s *AttachmentsService) Get(ctx context.Context, attachmentID string) (*Attachment, error) {
	var out Attachment
	if err := s.c.do(ctx, http.MethodGet, "/v3/attachments/"+attachmentID, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Upload PUTs the given content to the presigned URL. It reads body fully to
// set Content-Length, which S3-style presigned PUTs require.
func (s *AttachmentsService) Upload(ctx context.Context, att *Attachment, mimeType string, body io.Reader) error {
	buf, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("linq: read upload body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, att.UploadURL, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("linq: new upload request: %w", err)
	}
	req.ContentLength = int64(len(buf))
	if mimeType != "" {
		req.Header.Set("Content-Type", mimeType)
	}
	resp, err := s.c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("linq: upload: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("linq: upload failed: %d %s", resp.StatusCode, string(data))
	}
	return nil
}
