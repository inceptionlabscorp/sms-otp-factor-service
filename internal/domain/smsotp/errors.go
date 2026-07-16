package smsotp

import "errors"

var (
	ErrNotConfigured  = errors.New("sms otp service is not configured")
	ErrInvalidInput   = errors.New("invalid sms otp input")
	ErrRateLimited    = errors.New("too many sms otp requests")
	ErrInvalidCode    = errors.New("invalid sms otp code")
	ErrExpiredCode    = errors.New("expired sms otp code")
	ErrInvalidSession = errors.New("invalid sms mfa session")
)
