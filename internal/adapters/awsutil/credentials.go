package awsutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type CredentialProvider struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	HTTPClient      *http.Client
	Env             func(string) string
}

func (p CredentialProvider) Resolve(ctx context.Context) (Credentials, error) {
	if strings.TrimSpace(p.AccessKeyID) != "" && strings.TrimSpace(p.SecretAccessKey) != "" {
		return Credentials{
			AccessKeyID:     strings.TrimSpace(p.AccessKeyID),
			SecretAccessKey: strings.TrimSpace(p.SecretAccessKey),
			SessionToken:    strings.TrimSpace(p.SessionToken),
		}, nil
	}
	env := p.env
	accessKeyID := strings.TrimSpace(env("AWS_ACCESS_KEY_ID"))
	secretAccessKey := strings.TrimSpace(env("AWS_SECRET_ACCESS_KEY"))
	if accessKeyID != "" && secretAccessKey != "" {
		return Credentials{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
			SessionToken:    strings.TrimSpace(env("AWS_SESSION_TOKEN")),
		}, nil
	}
	return p.resolveContainer(ctx, env)
}

func (p CredentialProvider) resolveContainer(ctx context.Context, env func(string) string) (Credentials, error) {
	uri := strings.TrimSpace(env("AWS_CONTAINER_CREDENTIALS_FULL_URI"))
	if uri == "" {
		relative := strings.TrimSpace(env("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"))
		if relative != "" {
			uri = "http://169.254.170.2" + relative
		}
	}
	if uri == "" {
		return Credentials{}, fmt.Errorf("aws credentials are not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return Credentials{}, err
	}
	if token := strings.TrimSpace(env("AWS_CONTAINER_AUTHORIZATION_TOKEN")); token != "" {
		req.Header.Set("Authorization", token)
	}
	if tokenFile := strings.TrimSpace(env("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE")); tokenFile != "" {
		token, err := os.ReadFile(tokenFile)
		if err != nil {
			return Credentials{}, err
		}
		req.Header.Set("Authorization", strings.TrimSpace(string(token)))
	}
	res, err := p.client().Do(req)
	if err != nil {
		return Credentials{}, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return Credentials{}, fmt.Errorf("aws credential endpoint failed: %s %s", res.Status, strings.TrimSpace(string(body)))
	}
	var payload struct {
		AccessKeyID     string `json:"AccessKeyId"`
		SecretAccessKey string `json:"SecretAccessKey"`
		Token           string `json:"Token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return Credentials{}, err
	}
	if strings.TrimSpace(payload.AccessKeyID) == "" || strings.TrimSpace(payload.SecretAccessKey) == "" {
		return Credentials{}, fmt.Errorf("aws credential endpoint response is missing credentials")
	}
	return Credentials{
		AccessKeyID:     strings.TrimSpace(payload.AccessKeyID),
		SecretAccessKey: strings.TrimSpace(payload.SecretAccessKey),
		SessionToken:    strings.TrimSpace(payload.Token),
	}, nil
}

func (p CredentialProvider) client() *http.Client {
	if p.HTTPClient != nil {
		return p.HTTPClient
	}
	return &http.Client{Timeout: 5 * time.Second}
}

func (p CredentialProvider) env(key string) string {
	if p.Env != nil {
		return p.Env(key)
	}
	return os.Getenv(key)
}
