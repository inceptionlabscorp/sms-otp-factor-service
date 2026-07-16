package smsotp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	domain "github.com/inceptionlabscorp/sms-otp-factor-service/internal/domain/smsotp"
)

type Service struct {
	Challenges      ChallengeRepository
	SMS             SMSGateway
	Generator       CodeGenerator
	OTPSecret       string
	MessageTemplate string
	Now             func() time.Time
	Policy          domain.Policy
}

type SendInput struct {
	SubjectID   string
	PhoneNumber string
	Purpose     string
}

type VerifyInput struct {
	SubjectID   string
	PhoneNumber string
	Purpose     string
	Code        string
}

func (s Service) Send(ctx context.Context, input SendInput) error {
	if err := s.ready(); err != nil {
		return err
	}
	subjectID := domain.NormalizeSubject(input.SubjectID)
	phoneNumber := domain.NormalizePhone(input.PhoneNumber)
	purpose := domain.NormalizePurpose(input.Purpose)
	if subjectID == "" || phoneNumber == "" {
		return domain.ErrNotConfigured
	}

	now := s.now()
	key := domain.ChallengeKey(subjectID, purpose)
	existing, _ := s.Challenges.GetChallenge(ctx, key)
	requestCount := 1
	if existing != nil && now.Sub(existing.SentAt) < s.Policy.Window() {
		if existing.RequestCount >= s.Policy.MaxRequests() {
			return domain.ErrRateLimited
		}
		requestCount = existing.RequestCount + 1
	}

	code, err := s.Generator.Digits(s.Policy.Digits())
	if err != nil {
		return err
	}
	nonce, err := s.Generator.Nonce()
	if err != nil {
		return err
	}
	challenge := domain.Challenge{
		SubjectID:    subjectID,
		PhoneNumber:  phoneNumber,
		Purpose:      purpose,
		Hash:         s.hash(subjectID, phoneNumber, purpose, code, nonce),
		Nonce:        nonce,
		ExpiresAt:    now.Add(s.Policy.OTPTTL()),
		SentAt:       now,
		RequestCount: requestCount,
	}
	if err := s.Challenges.PutChallenge(ctx, key, challenge); err != nil {
		return err
	}
	return s.SMS.SendSMS(ctx, phoneNumber, s.messageBody(code))
}

func (s Service) Verify(ctx context.Context, input VerifyInput) error {
	if err := s.ready(); err != nil {
		return err
	}
	subjectID := domain.NormalizeSubject(input.SubjectID)
	phoneNumber := domain.NormalizePhone(input.PhoneNumber)
	purpose := domain.NormalizePurpose(input.Purpose)
	code := strings.TrimSpace(input.Code)
	if subjectID == "" || phoneNumber == "" || code == "" {
		return domain.ErrInvalidCode
	}

	key := domain.ChallengeKey(subjectID, purpose)
	challenge, err := s.Challenges.GetChallenge(ctx, key)
	if err != nil || challenge == nil || strings.TrimSpace(challenge.Hash) == "" {
		return domain.ErrInvalidCode
	}
	now := s.now()
	if challenge.Expired(now) {
		_ = s.Challenges.DeleteChallenge(ctx, key)
		return domain.ErrExpiredCode
	}
	if challenge.AttemptsExhausted(s.Policy.AllowedAttempts()) {
		_ = s.Challenges.DeleteChallenge(ctx, key)
		return domain.ErrInvalidCode
	}
	if challenge.SubjectID != subjectID || challenge.PhoneNumber != phoneNumber || challenge.Purpose != purpose {
		return domain.ErrInvalidCode
	}
	expected := s.hash(subjectID, phoneNumber, purpose, code, challenge.Nonce)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(challenge.Hash)) != 1 {
		challenge.Attempts++
		_ = s.Challenges.PutChallenge(ctx, key, *challenge)
		return domain.ErrInvalidCode
	}
	return s.Challenges.DeleteChallenge(ctx, key)
}

func (s Service) ready() error {
	if s.Challenges == nil || s.SMS == nil || s.Generator == nil || strings.TrimSpace(s.OTPSecret) == "" {
		return domain.ErrNotConfigured
	}
	return nil
}

func (s Service) hash(subjectID string, phoneNumber string, purpose string, code string, nonce string) string {
	mac := hmac.New(sha256.New, []byte(s.OTPSecret))
	_, _ = mac.Write([]byte(subjectID))
	_, _ = mac.Write([]byte("|"))
	_, _ = mac.Write([]byte(phoneNumber))
	_, _ = mac.Write([]byte("|"))
	_, _ = mac.Write([]byte(purpose))
	_, _ = mac.Write([]byte("|"))
	_, _ = mac.Write([]byte(code))
	_, _ = mac.Write([]byte("|"))
	_, _ = mac.Write([]byte(nonce))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s Service) messageBody(code string) string {
	minutes := fmt.Sprintf("%d", int(s.Policy.OTPTTL().Minutes()))
	template := strings.TrimSpace(s.MessageTemplate)
	if template == "" {
		template = "Your verification code is {{CODE}}. It expires in {{MINUTES}} minutes."
	}
	template = strings.ReplaceAll(template, "{{CODE}}", code)
	template = strings.ReplaceAll(template, "{{MINUTES}}", minutes)
	return template
}
