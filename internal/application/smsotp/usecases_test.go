package smsotp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	domain "github.com/inceptionlabscorp/sms-otp-factor-service/internal/domain/smsotp"
)

const (
	testOTPSecret       = "0123456789abcdef0123456789abcdef"
	testPhoneHashSecret = "abcdef0123456789abcdef0123456789"
	testSessionSecret   = "session-secret-0123456789abcdef01"
)

func TestServiceSendAndVerify(t *testing.T) {
	now := time.Date(2026, 7, 16, 1, 2, 3, 0, time.UTC)
	store := &testStore{}
	sms := &testSMS{}
	service := Service{
		Challenges:      store,
		SMS:             sms,
		Generator:       fixedGenerator{code: "585021", nonce: "nonce-1"},
		OTPSecret:       testOTPSecret,
		PhoneHashSecret: testPhoneHashSecret,
		Now:             func() time.Time { return now },
	}

	err := service.Send(context.Background(), SendInput{SubjectID: "uid-1", PhoneNumber: "+15555550100"})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if sms.to != "+15555550100" {
		t.Fatalf("sms to = %q", sms.to)
	}
	if !strings.Contains(sms.body, "585021") {
		t.Fatalf("sms body does not contain generated code: %q", sms.body)
	}
	if strings.Contains(sms.body, "LegacyBrand") {
		t.Fatalf("sms body leaked product branding: %q", sms.body)
	}
	key := domain.ChallengeKey("uid-1", domain.DefaultPurpose)
	if store.challenges[key].Hash == "585021" {
		t.Fatal("challenge stored OTP in clear text")
	}
	payload, err := json.Marshal(store.challenges[key])
	if err != nil {
		t.Fatalf("marshal challenge: %v", err)
	}
	if strings.Contains(string(payload), "+15555550100") {
		t.Fatalf("challenge stored phone number in clear text: %s", string(payload))
	}

	if err := service.Verify(context.Background(), VerifyInput{SubjectID: "uid-1", PhoneNumber: "+15555550100", Code: "585021"}); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if _, ok := store.challenges[key]; ok {
		t.Fatal("challenge was not cleared after verify")
	}
}

func TestServiceUsesConfiguredMessageTemplate(t *testing.T) {
	now := time.Date(2026, 7, 16, 1, 2, 3, 0, time.UTC)
	sms := &testSMS{}
	service := Service{
		Challenges:      &testStore{},
		SMS:             sms,
		Generator:       fixedGenerator{code: "585021", nonce: "nonce-1"},
		OTPSecret:       testOTPSecret,
		PhoneHashSecret: testPhoneHashSecret,
		MessageTemplate: "Codigo {{CODE}}, vence en {{MINUTES}} minutos.",
		Now:             func() time.Time { return now },
	}

	if err := service.Send(context.Background(), SendInput{SubjectID: "uid-1", PhoneNumber: "+15555550100"}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if sms.body != "Codigo 585021, vence en 5 minutos." {
		t.Fatalf("sms body = %q", sms.body)
	}
}

func TestServiceUsesDefaultMessageTemplate(t *testing.T) {
	now := time.Date(2026, 7, 16, 1, 2, 3, 0, time.UTC)
	sms := &testSMS{}
	service := Service{
		Challenges:      &testStore{},
		SMS:             sms,
		Generator:       fixedGenerator{code: "585021", nonce: "nonce-1"},
		OTPSecret:       testOTPSecret,
		PhoneHashSecret: testPhoneHashSecret,
		Now:             func() time.Time { return now },
	}

	if err := service.Send(context.Background(), SendInput{SubjectID: "uid-1", PhoneNumber: "+15555550100"}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if sms.body != "Your verification code is 585021. It expires in 5 minutes." {
		t.Fatalf("sms body = %q", sms.body)
	}
	if strings.Contains(sms.body, "{{") || strings.Contains(sms.body, "}}") {
		t.Fatalf("sms body contains unreplaced template markers: %q", sms.body)
	}
}

func TestServiceRateLimits(t *testing.T) {
	now := time.Date(2026, 7, 16, 1, 2, 3, 0, time.UTC)
	service := Service{
		Challenges:      &testStore{},
		SMS:             &testSMS{},
		Generator:       fixedGenerator{code: "585021", nonce: "nonce-1"},
		OTPSecret:       testOTPSecret,
		PhoneHashSecret: testPhoneHashSecret,
		Now:             func() time.Time { return now },
		Policy:          domain.Policy{RateLimitMax: 1},
	}
	input := SendInput{SubjectID: "uid-1", PhoneNumber: "+15555550100"}
	if err := service.Send(context.Background(), input); err != nil {
		t.Fatalf("first Send() error = %v", err)
	}
	if err := service.Send(context.Background(), input); err != domain.ErrRateLimited {
		t.Fatalf("second Send() error = %v, want ErrRateLimited", err)
	}
}

func TestServiceRejectsWeakOrMissingPhoneHashSecret(t *testing.T) {
	service := Service{
		Challenges: &testStore{},
		SMS:        &testSMS{},
		Generator:  fixedGenerator{code: "585021", nonce: "nonce-1"},
		OTPSecret:  testOTPSecret,
	}
	err := service.Send(context.Background(), SendInput{SubjectID: "uid-1", PhoneNumber: "+15555550100"})
	if err != domain.ErrNotConfigured {
		t.Fatalf("Send() error = %v, want ErrNotConfigured", err)
	}
}

func TestServiceRejectsInvalidPhoneAndCodeShape(t *testing.T) {
	service := Service{
		Challenges:      &testStore{},
		SMS:             &testSMS{},
		Generator:       fixedGenerator{code: "585021", nonce: "nonce-1"},
		OTPSecret:       testOTPSecret,
		PhoneHashSecret: testPhoneHashSecret,
	}
	if err := service.Send(context.Background(), SendInput{SubjectID: "uid-1", PhoneNumber: "555-0100"}); err != domain.ErrInvalidInput {
		t.Fatalf("Send() error = %v, want ErrInvalidInput", err)
	}
	if err := service.Verify(context.Background(), VerifyInput{SubjectID: "uid-1", PhoneNumber: "+15555550100", Code: "abc123"}); err != domain.ErrInvalidCode {
		t.Fatalf("Verify() error = %v, want ErrInvalidCode", err)
	}
}

func TestSessionService(t *testing.T) {
	now := time.Date(2026, 7, 16, 1, 2, 3, 0, time.UTC)
	service := SessionService{Secret: testSessionSecret, Now: func() time.Time { return now }}
	token, expiresIn, err := service.Sign("uid-1")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if expiresIn != int((8 * time.Hour).Seconds()) {
		t.Fatalf("expiresIn = %d", expiresIn)
	}
	if !service.Validate(token, "uid-1") {
		t.Fatal("expected token to validate")
	}
	if service.Validate(token, "other") {
		t.Fatal("token validated for another subject")
	}
	expired := SessionService{Secret: testSessionSecret, Now: func() time.Time { return now.Add(9 * time.Hour) }}
	if expired.Validate(token, "uid-1") {
		t.Fatal("expired token validated")
	}
}

type testStore struct {
	challenges map[string]domain.Challenge
}

func (s *testStore) GetChallenge(_ context.Context, key string) (*domain.Challenge, error) {
	if s.challenges == nil {
		s.challenges = map[string]domain.Challenge{}
	}
	challenge, ok := s.challenges[key]
	if !ok {
		return nil, nil
	}
	return &challenge, nil
}

func (s *testStore) PutChallenge(_ context.Context, key string, challenge domain.Challenge) error {
	if s.challenges == nil {
		s.challenges = map[string]domain.Challenge{}
	}
	s.challenges[key] = challenge
	return nil
}

func (s *testStore) DeleteChallenge(_ context.Context, key string) error {
	delete(s.challenges, key)
	return nil
}

type testSMS struct {
	body string
	to   string
}

func (s *testSMS) SendSMS(_ context.Context, to string, body string) error {
	s.to = to
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
