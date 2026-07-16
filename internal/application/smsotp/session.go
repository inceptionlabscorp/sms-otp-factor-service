package smsotp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	domain "github.com/inceptionlabscorp/sms-otp-factor-service/internal/domain/smsotp"
)

type SessionService struct {
	Secret string
	Now    func() time.Time
	Policy domain.Policy
}

type SessionClaims struct {
	SubjectID string `json:"sub"`
	Method    string `json:"method"`
	ExpiresAt int64  `json:"exp"`
}

func (s SessionService) Sign(subjectID string) (string, int, error) {
	subjectID = domain.NormalizeSubject(subjectID)
	secret := strings.TrimSpace(s.Secret)
	if subjectID == "" || len(secret) < 32 {
		return "", 0, domain.ErrInvalidSession
	}
	ttl := s.Policy.MFASessionTTL()
	claims := SessionClaims{
		SubjectID: subjectID,
		Method:    domain.SessionMethod,
		ExpiresAt: s.now().Add(ttl).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", 0, err
	}
	payloadPart := base64.RawURLEncoding.EncodeToString(payload)
	signaturePart := sign(payloadPart, secret)
	return payloadPart + "." + signaturePart, int(ttl.Seconds()), nil
}

func (s SessionService) Validate(token string, subjectID string) bool {
	token = strings.TrimSpace(token)
	subjectID = domain.NormalizeSubject(subjectID)
	secret := strings.TrimSpace(s.Secret)
	if token == "" || subjectID == "" || len(secret) < 32 {
		return false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	expected := sign(parts[0], secret)
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	var claims SessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false
	}
	return claims.SubjectID == subjectID && claims.Method == domain.SessionMethod && s.now().Unix() < claims.ExpiresAt
}

func (s SessionService) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func sign(payloadPart string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("sms-mfa-session|"))
	_, _ = mac.Write([]byte(payloadPart))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
