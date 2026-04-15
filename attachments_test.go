package linq

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAttachmentsCreateAndUpload(t *testing.T) {
	// Upload target captures PUT body.
	var gotUpload []byte
	upSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("upload method = %s", r.Method)
		}
		gotUpload = readAll(t, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer upSrv.Close()

	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/attachments" {
			t.Errorf("path = %s", r.URL.Path)
		}
		resp := fmt.Sprintf(`{"attachment_id":"att-1","upload_url":%q,"download_url":"https://example/d","expires_at":%q}`,
			upSrv.URL, time.Now().Add(15*time.Minute).Format(time.RFC3339))
		_, _ = w.Write([]byte(resp))
	})

	att, err := c.Attachments.Create(context.Background(), &CreateAttachmentRequest{FileName: "a.png", MimeType: "image/png"})
	if err != nil {
		t.Fatal(err)
	}
	if att.AttachmentID != "att-1" {
		t.Errorf("attachment_id = %q", att.AttachmentID)
	}

	payload := []byte("PNG-BYTES")
	if err := c.Attachments.Upload(context.Background(), att, "image/png", bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotUpload, payload) {
		t.Errorf("uploaded body = %q", gotUpload)
	}
}

func TestAttachmentsGet(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/attachments/att-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"attachment_id":"att-1","status":"ready","download_url":"https://x/d"}`))
	})
	att, err := c.Attachments.Get(context.Background(), "att-1")
	if err != nil {
		t.Fatal(err)
	}
	if att.Status != "ready" || att.DownloadURL != "https://x/d" {
		t.Errorf("att = %+v", att)
	}
}

func TestAttachmentsUploadFailure(t *testing.T) {
	upSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("denied"))
	}))
	defer upSrv.Close()
	c := NewClient("t")
	att := &Attachment{UploadURL: upSrv.URL}
	err := c.Attachments.Upload(context.Background(), att, "text/plain", bytes.NewReader([]byte("x")))
	if err == nil {
		t.Fatal("expected error")
	}
}
