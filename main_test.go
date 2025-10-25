package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/twilio/twilio-go/client"
)

func TestTwilioWebhookMiddleware(t *testing.T) {
	authToken := "test_auth_token"
	requestValidator = client.NewRequestValidator(authToken)

	handler := twilioWebhookMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	t.Run("rejects request without signature", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("rejects request with invalid signature", func(t *testing.T) {
		form := url.Values{}
		form.Add("From", "+15558675309")
		form.Add("To", "+15551234567")

		req := httptest.NewRequest("POST", "https://example.com/test", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-Twilio-Signature", "invalid_signature")

		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestHandlePhoneCall(t *testing.T) {
	os.Setenv("TWILIO_AUTH_TOKEN", "test_auth_token")
	requestValidator = client.NewRequestValidator("test_auth_token")

	t.Run("returns valid TwiML response", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		handlePhoneCall(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/xml" {
			t.Errorf("expected Content-Type 'application/xml', got '%s'", contentType)
		}

		body := w.Body.String()
		if !strings.Contains(body, "<Response>") {
			t.Error("response should contain <Response> tag")
		}
		if !strings.Contains(body, "<Gather") {
			t.Error("response should contain <Gather> tag")
		}
		if !strings.Contains(body, "<Say") {
			t.Error("response should contain <Say> tag")
		}
		if !strings.Contains(body, "To talk to sales, press 1") {
			t.Error("response should contain IVR menu message")
		}
	})
}
