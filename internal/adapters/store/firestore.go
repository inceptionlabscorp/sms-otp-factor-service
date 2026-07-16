package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	domain "github.com/inceptionlabscorp/sms-otp-factor-service/internal/domain/smsotp"
)

type FirestoreRESTStore struct {
	ProjectID  string
	Collection string
	HTTPClient *http.Client
	Token      func(context.Context) (string, error)
}

func (s *FirestoreRESTStore) GetChallenge(ctx context.Context, key string) (*domain.Challenge, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.documentURL(key), nil)
	if err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, req); err != nil {
		return nil, err
	}
	res, err := s.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return nil, fmt.Errorf("firestore get challenge failed: %s %s", res.Status, strings.TrimSpace(string(body)))
	}
	var doc firestoreDocument
	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return nil, err
	}
	payload := doc.Fields.Payload.StringValue
	if strings.TrimSpace(payload) == "" {
		return nil, nil
	}
	var challenge domain.Challenge
	if err := json.Unmarshal([]byte(payload), &challenge); err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (s *FirestoreRESTStore) PutChallenge(ctx context.Context, key string, challenge domain.Challenge) error {
	payload, err := json.Marshal(challenge)
	if err != nil {
		return err
	}
	body, err := json.Marshal(firestoreDocument{
		Fields: firestoreFields{
			Payload: firestoreString{StringValue: string(payload)},
		},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, s.documentURL(key)+"?updateMask.fieldPaths=payload", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if err := s.authorize(ctx, req); err != nil {
		return err
	}
	res, err := s.client().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("firestore put challenge failed: %s %s", res.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (s *FirestoreRESTStore) DeleteChallenge(ctx context.Context, key string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.documentURL(key), nil)
	if err != nil {
		return err
	}
	if err := s.authorize(ctx, req); err != nil {
		return err
	}
	res, err := s.client().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("firestore delete challenge failed: %s %s", res.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (s *FirestoreRESTStore) documentURL(key string) string {
	projectID := url.PathEscape(strings.TrimSpace(s.ProjectID))
	collection := strings.Trim(strings.TrimSpace(s.Collection), "/")
	if collection == "" {
		collection = "sms_otp_challenges"
	}
	return fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/(default)/documents/%s/%s", projectID, collection, url.PathEscape(key))
}

func (s *FirestoreRESTStore) authorize(ctx context.Context, req *http.Request) error {
	tokenSource := s.Token
	if tokenSource == nil {
		tokenSource = MetadataAccessToken
	}
	token, err := tokenSource(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (s *FirestoreRESTStore) client() *http.Client {
	if s.HTTPClient != nil {
		return s.HTTPClient
	}
	return http.DefaultClient
}

type firestoreDocument struct {
	Fields firestoreFields `json:"fields"`
}

type firestoreFields struct {
	Payload firestoreString `json:"payload"`
}

type firestoreString struct {
	StringValue string `json:"stringValue"`
}

func MetadataAccessToken(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Metadata-Flavor", "Google")
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return "", fmt.Errorf("metadata token failed: %s %s", res.Status, strings.TrimSpace(string(body)))
	}
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", fmt.Errorf("metadata token response missing access_token")
	}
	return payload.AccessToken, nil
}
