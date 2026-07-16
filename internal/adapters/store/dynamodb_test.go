package store

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domain "github.com/inceptionlabscorp/sms-otp-factor-service/internal/domain/smsotp"
)

func TestDynamoDBStorePutGetDeleteChallenge(t *testing.T) {
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		if r.Header.Get("Authorization") == "" {
			t.Fatal("missing SigV4 authorization header")
		}
		switch r.Header.Get("X-Amz-Target") {
		case "DynamoDB_20120810.PutItem":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case "DynamoDB_20120810.GetItem":
			w.WriteHeader(http.StatusOK)
			payload, _ := json.Marshal(domain.Challenge{SubjectID: "uid-1", PhoneHash: "phone-hash", Purpose: "sms_mfa"})
			_, _ = w.Write([]byte(`{"Item":{"payload":{"S":` + string(mustJSON(t, string(payload))) + `}}}`))
		case "DynamoDB_20120810.DeleteItem":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Fatalf("unexpected target %q", r.Header.Get("X-Amz-Target"))
		}
	}))
	defer server.Close()

	store := DynamoDBStore{
		Region:          "us-east-1",
		TableName:       "sms-otp-challenges",
		AccessKeyID:     "AKIDEXAMPLE",
		SecretAccessKey: "secret",
		Endpoint:        server.URL,
		Now:             func() time.Time { return time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC) },
	}

	challenge := domain.Challenge{
		SubjectID: "uid-1",
		PhoneHash: "phone-hash",
		Purpose:   "sms_mfa",
		ExpiresAt: time.Date(2026, 7, 16, 12, 5, 0, 0, time.UTC),
	}
	if err := store.PutChallenge(context.Background(), "key-1", challenge); err != nil {
		t.Fatalf("PutChallenge() error = %v", err)
	}
	got, err := store.GetChallenge(context.Background(), "key-1")
	if err != nil {
		t.Fatalf("GetChallenge() error = %v", err)
	}
	if got == nil || got.PhoneHash != "phone-hash" {
		t.Fatalf("challenge = %#v", got)
	}
	if err := store.DeleteChallenge(context.Background(), "key-1"); err != nil {
		t.Fatalf("DeleteChallenge() error = %v", err)
	}
	if strings.Contains(strings.Join(bodies, "\n"), "+15555550100") {
		t.Fatalf("DynamoDB body leaked raw phone: %s", strings.Join(bodies, "\n"))
	}
}

func mustJSON(t *testing.T, value string) []byte {
	t.Helper()
	out, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
