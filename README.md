# SMS OTP Factor Service

Provider-agnostic SMS OTP microservice for second-factor verification flows.

The service owns OTP challenge creation, SMS delivery, OTP verification, and short-lived MFA session token validation. It is designed as a private backend service: your trusted backend decides who the subject is and which phone number is authorized, then calls this service over an authenticated service-to-service API.

## Features

- Tactical DDD bounded context: `smsotp`.
- Hexagonal architecture: domain/application inside, HTTP/store/SMS providers outside.
- OTP challenges stored as HMAC hashes with nonce, expiry, attempts and cooldown.
- Pluggable SMS providers: Twilio Messaging Service or Amazon Simple Notification Service.
- Configurable SMS message template.
- Stable internal API plus BIAN-aligned aliases for banking-style integration language.
- No UI, no identity provider coupling, no phone-number ownership policy.

## Documentation

- [Installation](docs/installation.md)
- [Configuration](docs/configuration.md)
- [API integration](docs/api-integration.md)
- [Architecture](docs/architecture.md)
- [OpenAPI contract](docs/openapi.yaml)

## Quick Start

```bash
cp .env.example .env
export SMS_OTP_SERVICE_API_TOKEN="replace-with-a-long-random-token"
export SMS_OTP_SECRET="replace-with-at-least-32-random-bytes"
export SMS_MFA_SESSION_SECRET="replace-with-at-least-32-random-bytes"
export STORE_DRIVER=memory
export SMS_PROVIDER=twilio
export TWILIO_ACCOUNT_SID="AC..."
export TWILIO_API_KEY_SID="SK..."
export TWILIO_API_KEY_SECRET="..."
export TWILIO_MESSAGING_SERVICE_SID="MG..."
go run ./cmd/server
```

Health check:

```bash
curl http://localhost:8080/health
```

Send OTP:

```bash
curl -X POST http://localhost:8080/v1/sms-otp/send \
  -H "Authorization: Bearer ${SMS_OTP_SERVICE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"subject_id":"user-123","phone_number":"+15555550100"}'
```

Verify OTP:

```bash
curl -X POST http://localhost:8080/v1/sms-otp/verify \
  -H "Authorization: Bearer ${SMS_OTP_SERVICE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"subject_id":"user-123","phone_number":"+15555550100","code":"123456"}'
```

The verify response returns an `mfa_token`. Your backend can validate it with:

```bash
curl -X POST http://localhost:8080/v1/sms-mfa/session/validate \
  -H "Authorization: Bearer ${SMS_OTP_SERVICE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"subject_id":"user-123","token":"<mfa_token>"}'
```

## Security Model

- Do not expose this service directly to browsers or mobile apps.
- Keep `SMS_OTP_SERVICE_API_TOKEN`, `SMS_OTP_SECRET`, provider credentials, and session secrets outside the repository.
- The caller must authorize the subject and phone number before calling `/send`.
- The service must not be the source of truth for users, roles, or authorized phone numbers.
- Logs must not include OTP codes, session tokens, full phone numbers, or secrets.

## Development

```bash
go test ./...
go build ./cmd/server
git diff --check
```

## License

Apache-2.0. See [LICENSE](LICENSE).
