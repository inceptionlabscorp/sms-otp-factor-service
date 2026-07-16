package smsotp

import "time"

const (
	DefaultPurpose = "sms_mfa"
	SessionMethod  = "sms"
)

type Challenge struct {
	SubjectID    string    `json:"subject_id"`
	PhoneNumber  string    `json:"phone_number"`
	Purpose      string    `json:"purpose"`
	Hash         string    `json:"hash"`
	Nonce        string    `json:"nonce"`
	ExpiresAt    time.Time `json:"expires_at"`
	SentAt       time.Time `json:"sent_at"`
	RequestCount int       `json:"request_count"`
	Attempts     int       `json:"attempts"`
}

func (c Challenge) Expired(now time.Time) bool {
	return now.UTC().After(c.ExpiresAt)
}

func (c Challenge) AttemptsExhausted(maxAttempts int) bool {
	return c.Attempts >= maxAttempts
}
