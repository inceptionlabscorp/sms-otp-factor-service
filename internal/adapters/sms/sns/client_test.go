package sns

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientSendSMSSignsAndPublishesToSNS(t *testing.T) {
	var gotBody string
	var gotAuth string
	var gotDate string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		gotAuth = r.Header.Get("Authorization")
		gotDate = r.Header.Get("X-Amz-Date")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(`<PublishResponse><PublishResult><MessageId>mid</MessageId></PublishResult></PublishResponse>`))
	}))
	defer server.Close()

	client := Client{
		Region:          "us-east-1",
		AccessKeyID:     "AKIDEXAMPLE",
		SecretAccessKey: "secret",
		SMSType:         "Transactional",
		SenderID:        "Verify",
		Endpoint:        server.URL,
		Now:             func() time.Time { return time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC) },
	}

	if err := client.SendSMS(t.Context(), "+15555550100", "Your verification code is 599371."); err != nil {
		t.Fatalf("SendSMS() error = %v", err)
	}

	if !strings.Contains(gotBody, "Action=Publish") {
		t.Fatalf("body missing Publish action: %s", gotBody)
	}
	if !strings.Contains(gotBody, "PhoneNumber=%2B15555550100") {
		t.Fatalf("body missing phone number: %s", gotBody)
	}
	if !strings.Contains(gotBody, "MessageAttributes.entry.1.Name=AWS.SNS.SMS.SMSType") {
		t.Fatalf("body missing sms type attribute: %s", gotBody)
	}
	if !strings.Contains(gotBody, "MessageAttributes.entry.2.Name=AWS.SNS.SMS.SenderID") {
		t.Fatalf("body missing sender id attribute: %s", gotBody)
	}
	if !strings.HasPrefix(gotAuth, "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20260716/us-east-1/sns/aws4_request") {
		t.Fatalf("authorization header = %q", gotAuth)
	}
	if gotDate != "20260716T120000Z" {
		t.Fatalf("x-amz-date = %q", gotDate)
	}
}

func TestClientSendSMSRequiresConfiguration(t *testing.T) {
	if err := (Client{}).SendSMS(t.Context(), "+15555550100", "body"); err == nil {
		t.Fatal("expected missing config error")
	}
}
