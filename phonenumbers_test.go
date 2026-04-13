package linq

import (
	"context"
	"net/http"
	"testing"
)

func TestPhoneNumbersList(t *testing.T) {
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/phonenumbers" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"phone_numbers":[{"number":"+15551234567","is_default":true}]}`))
	})
	res, err := c.PhoneNumbers.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.PhoneNumbers) != 1 || res.PhoneNumbers[0].Number != "+15551234567" {
		t.Errorf("numbers = %+v", res.PhoneNumbers)
	}
}
