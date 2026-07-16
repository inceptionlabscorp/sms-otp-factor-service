package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	app "github.com/inceptionlabscorp/sms-otp-factor-service/internal/application/smsotp"
	domain "github.com/inceptionlabscorp/sms-otp-factor-service/internal/domain/smsotp"
)

func TestHandlerRejectsMissingToken(t *testing.T) {
	handler := Handler{ServiceToken: "service-token"}
	req := httptest.NewRequest(http.MethodPost, "/v1/sms-otp/send", strings.NewReader(`{}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.Code)
	}
}

func TestHandlerSendVerifyAndValidate(t *testing.T) {
	now := time.Date(2026, 7, 16, 1, 2, 3, 0, time.UTC)
	store := &httpTestStore{challenges: map[string]domain.Challenge{}}
	sms := &httpTestSMS{}
	handler := Handler{
		ServiceToken: "service-token",
		OTP: app.Service{
			Challenges: store,
			SMS:        sms,
			Generator:  fixedGenerator{code: "599371", nonce: "nonce-1"},
			OTPSecret:  "otp-secret",
			Now:        func() time.Time { return now },
		},
		Session: app.SessionService{Secret: "session-secret", Now: func() time.Time { return now }},
	}

	sendBody := bytes.NewBufferString(`{"subject_id":"uid-1","phone_number":"+15555550100"}`)
	sendReq := authedRequest(http.MethodPost, "/v1/sms-otp/send", sendBody)
	sendRes := httptest.NewRecorder()
	handler.ServeHTTP(sendRes, sendReq)
	if sendRes.Code != http.StatusAccepted {
		t.Fatalf("send status = %d body=%s", sendRes.Code, sendRes.Body.String())
	}
	if !strings.Contains(sms.body, "599371") {
		t.Fatalf("sms missing code: %q", sms.body)
	}

	verifyBody := bytes.NewBufferString(`{"subject_id":"uid-1","phone_number":"+15555550100","code":"599371"}`)
	verifyReq := authedRequest(http.MethodPost, "/v1/sms-otp/verify", verifyBody)
	verifyRes := httptest.NewRecorder()
	handler.ServeHTTP(verifyRes, verifyReq)
	if verifyRes.Code != http.StatusOK {
		t.Fatalf("verify status = %d body=%s", verifyRes.Code, verifyRes.Body.String())
	}
	var verifyPayload map[string]any
	if err := json.Unmarshal(verifyRes.Body.Bytes(), &verifyPayload); err != nil {
		t.Fatalf("decode verify: %v", err)
	}
	token, _ := verifyPayload["mfa_token"].(string)
	if token == "" {
		t.Fatalf("missing mfa_token: %#v", verifyPayload)
	}

	validateBody := bytes.NewBufferString(`{"subject_id":"uid-1","token":"` + token + `"}`)
	validateReq := authedRequest(http.MethodPost, "/v1/sms-mfa/session/validate", validateBody)
	validateRes := httptest.NewRecorder()
	handler.ServeHTTP(validateRes, validateReq)
	if validateRes.Code != http.StatusOK {
		t.Fatalf("validate status = %d body=%s", validateRes.Code, validateRes.Body.String())
	}
	var validatePayload map[string]bool
	if err := json.Unmarshal(validateRes.Body.Bytes(), &validatePayload); err != nil {
		t.Fatalf("decode validate: %v", err)
	}
	if !validatePayload["valid"] {
		t.Fatal("expected token to validate")
	}
}

func TestHandlerSupportsBIANAlignedOTPContract(t *testing.T) {
	now := time.Date(2026, 7, 16, 1, 2, 3, 0, time.UTC)
	store := &httpTestStore{challenges: map[string]domain.Challenge{}}
	sms := &httpTestSMS{}
	handler := Handler{
		ServiceToken: "service-token",
		OTP: app.Service{
			Challenges: store,
			SMS:        sms,
			Generator:  fixedGenerator{code: "599371", nonce: "nonce-1"},
			OTPSecret:  "otp-secret",
			Now:        func() time.Time { return now },
		},
		Session: app.SessionService{Secret: "session-secret", Now: func() time.Time { return now }},
	}

	sendReq := authedRequest(http.MethodPost, "/v1/bian/customer-access-entitlement/sms-otp/initiate", bytes.NewBufferString(`{"subject_id":"uid-1","phone_number":"+15555550100"}`))
	sendRes := httptest.NewRecorder()
	handler.ServeHTTP(sendRes, sendReq)
	if sendRes.Code != http.StatusAccepted {
		t.Fatalf("send status = %d body=%s", sendRes.Code, sendRes.Body.String())
	}

	verifyReq := authedRequest(http.MethodPost, "/v1/bian/customer-access-entitlement/sms-otp/execute", bytes.NewBufferString(`{"subject_id":"uid-1","phone_number":"+15555550100","code":"599371"}`))
	verifyRes := httptest.NewRecorder()
	handler.ServeHTTP(verifyRes, verifyReq)
	if verifyRes.Code != http.StatusOK {
		t.Fatalf("verify status = %d body=%s", verifyRes.Code, verifyRes.Body.String())
	}
	var verifyPayload map[string]any
	if err := json.Unmarshal(verifyRes.Body.Bytes(), &verifyPayload); err != nil {
		t.Fatalf("decode verify: %v", err)
	}
	token, _ := verifyPayload["mfa_token"].(string)

	validateReq := authedRequest(http.MethodPost, "/v1/bian/customer-access-entitlement/sms-mfa-session/evaluate", bytes.NewBufferString(`{"subject_id":"uid-1","token":"`+token+`"}`))
	validateRes := httptest.NewRecorder()
	handler.ServeHTTP(validateRes, validateReq)
	if validateRes.Code != http.StatusOK {
		t.Fatalf("validate status = %d body=%s", validateRes.Code, validateRes.Body.String())
	}
}

func authedRequest(method string, path string, body *bytes.Buffer) *http.Request {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Authorization", "Bearer service-token")
	req.Header.Set("Content-Type", "application/json")
	return req
}

type httpTestStore struct {
	challenges map[string]domain.Challenge
}

func (s *httpTestStore) GetChallenge(_ context.Context, key string) (*domain.Challenge, error) {
	challenge, ok := s.challenges[key]
	if !ok {
		return nil, nil
	}
	return &challenge, nil
}

func (s *httpTestStore) PutChallenge(_ context.Context, key string, challenge domain.Challenge) error {
	s.challenges[key] = challenge
	return nil
}

func (s *httpTestStore) DeleteChallenge(_ context.Context, key string) error {
	delete(s.challenges, key)
	return nil
}

type httpTestSMS struct {
	body string
}

func (s *httpTestSMS) SendSMS(_ context.Context, _ string, body string) error {
	s.body = body
	return nil
}

type fixedGenerator struct {
	code  string
	nonce string
}

func (g fixedGenerator) Digits(_ int) (string, error) {
	return g.code, nil
}

func (g fixedGenerator) Nonce() (string, error) {
	return g.nonce, nil
}
