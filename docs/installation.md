# Installation

Español: [Instalacion](es/installation.md)

## Requirements

- Go 1.22 or newer.
- A trusted backend that can call this service over HTTP.
- One SMS provider:
  - Twilio Messaging Service, or
  - Amazon Simple Notification Service.
- A challenge store:
  - `memory` for local development and tests.
  - `firestore` for production deployments on Google Cloud.

## Local Installation

```bash
git clone https://github.com/inceptionlabscorp/sms-otp-factor-service.git
cd sms-otp-factor-service
cp .env.example .env
```

Set environment variables from `.env.example`, then run:

```bash
go test ./...
go run ./cmd/server
```

The service listens on `PORT`, defaulting to `8080`.

## Docker

```bash
docker build -t sms-otp-factor-service:local .
docker run --rm -p 8080:8080 \
  --env-file .env \
  sms-otp-factor-service:local
```

## Google Cloud Run

For a full GCP install that provisions Artifact Registry, Firestore, Secret Manager, IAM and Cloud Run, use:

```bash
./scripts/install-gcp.sh
```

The legacy deployment script only builds and deploys Cloud Run against existing infrastructure:

```bash
export GCP_PROJECT_ID="example-gcp-project"
export REGION="us-central1"
export SERVICE="sms-otp-factor-service"
export SERVICE_ACCOUNT="sms-otp-factor-service@example-gcp-project.iam.gserviceaccount.com"
./scripts/deploy-prod.sh
```

`GCP_PROJECT_ID` is intentionally required. The script does not contain production project defaults.

## AWS App Runner

For a full AWS install that provisions ECR, DynamoDB, Secrets Manager, IAM and App Runner, use:

```bash
./scripts/install-aws.sh
```

AWS production deployments use `STORE_DRIVER=dynamodb`.

## Production Checklist

- Use `STORE_DRIVER=firestore`.
- Store secrets in Secret Manager or an equivalent secret manager.
- Use separate random values for `SMS_OTP_SECRET`, `SMS_PHONE_HASH_SECRET`, and `SMS_MFA_SESSION_SECRET`.
- Restrict network access to the trusted backend where possible.
- Rotate `SMS_OTP_SERVICE_API_TOKEN` and HMAC secrets with an operational runbook.
- Keep provider credentials scoped to SMS sending only.
- Confirm logs do not include OTP codes, tokens, full phone numbers, or secrets.
