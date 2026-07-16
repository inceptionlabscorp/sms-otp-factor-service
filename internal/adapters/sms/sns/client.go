package sns

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/inceptionlabscorp/sms-otp-factor-service/internal/adapters/awsutil"
)

const (
	serviceName = "sns"
)

type Client struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	SMSType         string
	SenderID        string
	Endpoint        string
	HTTPClient      *http.Client
	Now             func() time.Time
}

func (c Client) SendSMS(ctx context.Context, to string, body string) error {
	if strings.TrimSpace(c.Region) == "" ||
		(strings.TrimSpace(c.AccessKeyID) == "" && strings.TrimSpace(c.SecretAccessKey) != "") ||
		(strings.TrimSpace(c.AccessKeyID) != "" && strings.TrimSpace(c.SecretAccessKey) == "") {
		return fmt.Errorf("amazon sns sms client is not configured")
	}
	credentials, err := c.credentials(ctx)
	if err != nil {
		return err
	}
	endpoint := c.endpoint()
	form := url.Values{}
	form.Set("Action", "Publish")
	form.Set("Version", "2010-03-31")
	form.Set("PhoneNumber", strings.TrimSpace(to))
	form.Set("Message", body)
	smsType := strings.TrimSpace(c.SMSType)
	if smsType == "" {
		smsType = "Transactional"
	}
	form.Set("MessageAttributes.entry.1.Name", "AWS.SNS.SMS.SMSType")
	form.Set("MessageAttributes.entry.1.Value.DataType", "String")
	form.Set("MessageAttributes.entry.1.Value.StringValue", smsType)
	if senderID := strings.TrimSpace(c.SenderID); senderID != "" {
		form.Set("MessageAttributes.entry.2.Name", "AWS.SNS.SMS.SenderID")
		form.Set("MessageAttributes.entry.2.Value.DataType", "String")
		form.Set("MessageAttributes.entry.2.Value.StringValue", senderID)
	}

	payload := form.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/xml, application/xml")
	awsutil.Sign(req, payload, serviceName, strings.TrimSpace(c.Region), credentials, c.now())
	res, err := c.client().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		response, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return fmt.Errorf("amazon sns publish sms failed: %s %s", res.Status, strings.TrimSpace(string(response)))
	}
	io.Copy(io.Discard, res.Body)
	return nil
}

func (c Client) endpoint() string {
	if endpoint := strings.TrimSpace(c.Endpoint); endpoint != "" {
		return endpoint
	}
	return fmt.Sprintf("https://sns.%s.amazonaws.com/", strings.TrimSpace(c.Region))
}

func (c Client) now() time.Time {
	if c.Now != nil {
		return c.Now().UTC()
	}
	return time.Now().UTC()
}

func (c Client) credentials(ctx context.Context) (awsutil.Credentials, error) {
	return awsutil.CredentialProvider{
		AccessKeyID:     c.AccessKeyID,
		SecretAccessKey: c.SecretAccessKey,
		SessionToken:    c.SessionToken,
		HTTPClient:      c.HTTPClient,
	}.Resolve(ctx)
}

func (c Client) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}
