package sns

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	serviceName = "sns"
	aws4Request = "aws4_request"
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
		strings.TrimSpace(c.AccessKeyID) == "" ||
		strings.TrimSpace(c.SecretAccessKey) == "" {
		return fmt.Errorf("amazon sns sms client is not configured")
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
	if err := c.sign(req, payload); err != nil {
		return err
	}
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

func (c Client) sign(req *http.Request, payload string) error {
	now := c.now()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	region := strings.TrimSpace(c.Region)
	scope := dateStamp + "/" + region + "/" + serviceName + "/" + aws4Request

	req.Header.Set("X-Amz-Date", amzDate)
	if token := strings.TrimSpace(c.SessionToken); token != "" {
		req.Header.Set("X-Amz-Security-Token", token)
	}

	signedHeaders := "content-type;host;x-amz-date"
	canonicalHeaders := "content-type:" + req.Header.Get("Content-Type") + "\n" +
		"host:" + req.URL.Host + "\n" +
		"x-amz-date:" + amzDate + "\n"
	if strings.TrimSpace(c.SessionToken) != "" {
		signedHeaders = "content-type;host;x-amz-date;x-amz-security-token"
		canonicalHeaders += "x-amz-security-token:" + strings.TrimSpace(c.SessionToken) + "\n"
	}

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI(req.URL),
		"",
		canonicalHeaders,
		signedHeaders,
		sha256Hex(payload),
	}, "\n")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		sha256Hex(canonicalRequest),
	}, "\n")
	signature := hex.EncodeToString(hmacSHA256(signingKey(strings.TrimSpace(c.SecretAccessKey), dateStamp, region), stringToSign))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+strings.TrimSpace(c.AccessKeyID)+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+signature)
	return nil
}

func canonicalURI(u *url.URL) string {
	if u.EscapedPath() == "" {
		return "/"
	}
	return u.EscapedPath()
}

func signingKey(secret string, dateStamp string, region string) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secret), dateStamp)
	dateRegionKey := hmacSHA256(dateKey, region)
	dateRegionServiceKey := hmacSHA256(dateRegionKey, serviceName)
	return hmacSHA256(dateRegionServiceKey, aws4Request)
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(data))
	return mac.Sum(nil)
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (c Client) now() time.Time {
	if c.Now != nil {
		return c.Now().UTC()
	}
	return time.Now().UTC()
}

func (c Client) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}
