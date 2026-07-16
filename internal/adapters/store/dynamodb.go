package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/inceptionlabscorp/sms-otp-factor-service/internal/adapters/awsutil"
	domain "github.com/inceptionlabscorp/sms-otp-factor-service/internal/domain/smsotp"
)

const dynamoDBServiceName = "dynamodb"

type DynamoDBStore struct {
	Region          string
	TableName       string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Endpoint        string
	HTTPClient      *http.Client
	Now             func() time.Time
}

func (s DynamoDBStore) GetChallenge(ctx context.Context, key string) (*domain.Challenge, error) {
	body, err := json.Marshal(map[string]any{
		"TableName":      s.tableName(),
		"ConsistentRead": true,
		"Key": map[string]any{
			"challenge_key": map[string]string{"S": key},
		},
	})
	if err != nil {
		return nil, err
	}
	res, err := s.call(ctx, "DynamoDB_20120810.GetItem", body)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Item map[string]map[string]string `json:"Item"`
	}
	if err := json.Unmarshal(res, &payload); err != nil {
		return nil, err
	}
	itemPayload := payload.Item["payload"]["S"]
	if strings.TrimSpace(itemPayload) == "" {
		return nil, nil
	}
	var challenge domain.Challenge
	if err := json.Unmarshal([]byte(itemPayload), &challenge); err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (s DynamoDBStore) PutChallenge(ctx context.Context, key string, challenge domain.Challenge) error {
	payload, err := json.Marshal(challenge)
	if err != nil {
		return err
	}
	body, err := json.Marshal(map[string]any{
		"TableName": s.tableName(),
		"Item": map[string]any{
			"challenge_key": map[string]string{"S": key},
			"payload":       map[string]string{"S": string(payload)},
			"expires_at":    map[string]string{"N": fmt.Sprintf("%d", challenge.ExpiresAt.Unix())},
		},
	})
	if err != nil {
		return err
	}
	_, err = s.call(ctx, "DynamoDB_20120810.PutItem", body)
	return err
}

func (s DynamoDBStore) DeleteChallenge(ctx context.Context, key string) error {
	body, err := json.Marshal(map[string]any{
		"TableName": s.tableName(),
		"Key": map[string]any{
			"challenge_key": map[string]string{"S": key},
		},
	})
	if err != nil {
		return err
	}
	_, err = s.call(ctx, "DynamoDB_20120810.DeleteItem", body)
	return err
}

func (s DynamoDBStore) call(ctx context.Context, target string, body []byte) ([]byte, error) {
	if strings.TrimSpace(s.Region) == "" || strings.TrimSpace(s.tableName()) == "" {
		return nil, fmt.Errorf("dynamodb store is not configured")
	}
	credentials, err := awsutil.CredentialProvider{
		AccessKeyID:     s.AccessKeyID,
		SecretAccessKey: s.SecretAccessKey,
		SessionToken:    s.SessionToken,
		HTTPClient:      s.HTTPClient,
	}.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-amz-json-1.0")
	req.Header.Set("X-Amz-Target", target)
	awsutil.Sign(req, string(body), dynamoDBServiceName, strings.TrimSpace(s.Region), credentials, s.now())
	res, err := s.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	response, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("dynamodb operation failed: %s %s", res.Status, strings.TrimSpace(string(response)))
	}
	return response, nil
}

func (s DynamoDBStore) endpoint() string {
	if endpoint := strings.TrimSpace(s.Endpoint); endpoint != "" {
		return endpoint
	}
	return fmt.Sprintf("https://dynamodb.%s.amazonaws.com/", strings.TrimSpace(s.Region))
}

func (s DynamoDBStore) tableName() string {
	if table := strings.TrimSpace(s.TableName); table != "" {
		return table
	}
	return "sms-otp-challenges"
}

func (s DynamoDBStore) client() *http.Client {
	if s.HTTPClient != nil {
		return s.HTTPClient
	}
	return http.DefaultClient
}

func (s DynamoDBStore) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}
