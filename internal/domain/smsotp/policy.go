package smsotp

import (
	"crypto/sha256"
	"encoding/base64"
	"regexp"
	"strings"
	"time"
)

var e164Pattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

type Policy struct {
	TTL             time.Duration
	RateLimitWindow time.Duration
	RateLimitMax    int
	MaxAttempts     int
	CodeLength      int
	SessionTTL      time.Duration
}

func (p Policy) OTPTTL() time.Duration {
	if p.TTL > 0 {
		return p.TTL
	}
	return 5 * time.Minute
}

func (p Policy) Window() time.Duration {
	if p.RateLimitWindow > 0 {
		return p.RateLimitWindow
	}
	return 10 * time.Minute
}

func (p Policy) MaxRequests() int {
	if p.RateLimitMax > 0 {
		return p.RateLimitMax
	}
	return 3
}

func (p Policy) AllowedAttempts() int {
	if p.MaxAttempts > 0 {
		return p.MaxAttempts
	}
	return 5
}

func (p Policy) Digits() int {
	if p.CodeLength >= 6 && p.CodeLength <= 8 {
		return p.CodeLength
	}
	return 6
}

func (p Policy) MFASessionTTL() time.Duration {
	if p.SessionTTL > 0 {
		return p.SessionTTL
	}
	return 8 * time.Hour
}

func NormalizeSubject(value string) string {
	return strings.TrimSpace(value)
}

func NormalizePhone(value string) string {
	return strings.TrimSpace(value)
}

func ValidPhone(value string) bool {
	return e164Pattern.MatchString(NormalizePhone(value))
}

func ValidOTPCode(value string, length int) bool {
	value = strings.TrimSpace(value)
	if len(value) != length {
		return false
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func NormalizePurpose(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultPurpose
	}
	return value
}

func ChallengeKey(subjectID string, purpose string) string {
	sum := sha256.Sum256([]byte(NormalizeSubject(subjectID) + "|" + NormalizePurpose(purpose)))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
