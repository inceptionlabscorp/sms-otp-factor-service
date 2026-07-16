package twilio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	AccountSID       string
	APIKeySID        string
	APIKeySecret     string
	MessagingService string
	HTTPClient       *http.Client
}

func (c Client) SendSMS(ctx context.Context, to string, body string) error {
	if strings.TrimSpace(c.AccountSID) == "" ||
		strings.TrimSpace(c.APIKeySID) == "" ||
		strings.TrimSpace(c.APIKeySecret) == "" ||
		strings.TrimSpace(c.MessagingService) == "" {
		return fmt.Errorf("twilio sms client is not configured")
	}
	form := url.Values{}
	form.Set("To", strings.TrimSpace(to))
	form.Set("MessagingServiceSid", strings.TrimSpace(c.MessagingService))
	form.Set("Body", body)
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", url.PathEscape(strings.TrimSpace(c.AccountSID)))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(strings.TrimSpace(c.APIKeySID), strings.TrimSpace(c.APIKeySecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := c.client().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("twilio send sms failed: %s %s", res.Status, strings.TrimSpace(string(payload)))
	}
	return nil
}

func (c Client) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}
