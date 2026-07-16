package smsotp

import (
	"context"

	domain "github.com/inceptionlabscorp/sms-otp-factor-service/internal/domain/smsotp"
)

type ChallengeRepository interface {
	GetChallenge(ctx context.Context, key string) (*domain.Challenge, error)
	PutChallenge(ctx context.Context, key string, challenge domain.Challenge) error
	DeleteChallenge(ctx context.Context, key string) error
}

type SMSGateway interface {
	SendSMS(ctx context.Context, to string, body string) error
}

type CodeGenerator interface {
	Digits(length int) (string, error)
	Nonce() (string, error)
}
