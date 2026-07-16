# Configuration

Configuration is environment-variable based. Secrets must be injected by the runtime or a secret manager, never committed to the repository.

## Core Variables

| Variable | Required | Description |
| --- | --- | --- |
| `PORT` | No | HTTP port. Default: `8080`. |
| `SMS_OTP_SERVICE_API_TOKEN` | Yes | Bearer token required by all `/v1/*` endpoints. |
| `SMS_OTP_SECRET` | Yes | HMAC secret used to hash OTP challenges. Use at least 32 random bytes. |
| `SMS_PHONE_HASH_SECRET` | Yes | Separate HMAC secret used to fingerprint phone numbers in storage. Use at least 32 random bytes. |
| `SMS_MFA_SESSION_SECRET` | Yes | HMAC secret used to sign MFA session tokens. Use at least 32 random bytes. |
| `OTP_MESSAGE_TEMPLATE` | No | SMS body template. Supports `{{CODE}}` and `{{MINUTES}}`. |

Default message template:

```text
Your verification code is {{CODE}}. It expires in {{MINUTES}} minutes.
```

## Store Variables

| Variable | Required | Description |
| --- | --- | --- |
| `STORE_DRIVER` | No | `firestore` or `memory`. Default: `firestore`. |
| `GCP_PROJECT_ID` | For Firestore | Google Cloud project used by the Firestore REST adapter. |
| `GOOGLE_CLOUD_PROJECT` | For Firestore | Alternative project variable used when `GCP_PROJECT_ID` is absent. |
| `FIRESTORE_COLLECTION` | No | Challenge collection. Default: `sms_otp_challenges`. |

Use `memory` only for local development and tests.

## SMS Provider Selection

| Variable | Required | Description |
| --- | --- | --- |
| `SMS_PROVIDER` | No | `twilio` or `amazon_sns`. Default: `twilio`. |

## Twilio Variables

Required when `SMS_PROVIDER=twilio`.

| Variable | Description |
| --- | --- |
| `TWILIO_ACCOUNT_SID` | Twilio account SID, usually starts with `AC`. |
| `TWILIO_API_KEY_SID` | Twilio API key SID, usually starts with `SK`. |
| `TWILIO_API_KEY_SECRET` | Twilio API key secret. |
| `TWILIO_MESSAGING_SERVICE_SID` | Twilio Messaging Service SID, usually starts with `MG`. |

The service uses Twilio raw HTTP APIs and does not require a Twilio SDK.

## Amazon SNS Variables

Required when `SMS_PROVIDER=amazon_sns`.

| Variable | Required | Description |
| --- | --- | --- |
| `AWS_REGION` | Yes | AWS region for SNS, for example `us-east-1`. |
| `AWS_ACCESS_KEY_ID` | Yes | AWS access key with `sns:Publish`. |
| `AWS_SECRET_ACCESS_KEY` | Yes | AWS secret access key. |
| `AWS_SESSION_TOKEN` | No | Temporary credential session token. |
| `AWS_SNS_SMS_TYPE` | No | `Transactional` or `Promotional`. Default: `Transactional`. |
| `AWS_SNS_SENDER_ID` | No | Sender ID where supported by destination country/carrier. |

The adapter signs requests with AWS SigV4.

## Recommended Secret Names

These names are used by the Cloud Run deploy script when the secrets exist:

| Environment variable | Suggested secret name |
| --- | --- |
| `SMS_OTP_SERVICE_API_TOKEN` | `sms-otp-service-api-token` |
| `SMS_OTP_SECRET` | `sms-otp-secret` |
| `SMS_PHONE_HASH_SECRET` | `sms-phone-hash-secret` |
| `SMS_MFA_SESSION_SECRET` | `sms-mfa-session-secret` |
| `TWILIO_ACCOUNT_SID` | `twilio-account-sid` |
| `TWILIO_API_KEY_SID` | `twilio-api-key-sid` |
| `TWILIO_API_KEY_SECRET` | `twilio-api-key-secret` |
| `TWILIO_MESSAGING_SERVICE_SID` | `twilio-messaging-service-sid` |
| `AWS_ACCESS_KEY_ID` | `aws-sns-access-key-id` |
| `AWS_SECRET_ACCESS_KEY` | `aws-sns-secret-access-key` |
| `AWS_SESSION_TOKEN` | `aws-sns-session-token` |
