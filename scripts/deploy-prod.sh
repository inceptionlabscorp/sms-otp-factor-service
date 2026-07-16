#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

: "${GCP_PROJECT_ID:?GCP_PROJECT_ID is required}"
PROJECT_ID="${GCP_PROJECT_ID}"
REGION="${REGION:-us-central1}"
SERVICE="${SERVICE:-sms-otp-factor-service}"
TAG="${TAG:-$(git rev-parse --short HEAD 2>/dev/null || date +%Y%m%d%H%M%S)}-local"
IMAGE="${IMAGE:-gcr.io/${PROJECT_ID}/${SERVICE}:${TAG}}"
SERVICE_ACCOUNT="${SERVICE_ACCOUNT:-}"

echo "Testing..."
go test ./...

echo "Building linux/amd64 binary..."
mkdir -p .build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-w -s" -o .build/server ./cmd/server

gcloud config set project "$PROJECT_ID"
gcloud auth configure-docker gcr.io --quiet

echo "Building Docker image: ${IMAGE}"
docker build --platform linux/amd64 -t "${IMAGE}" .
docker push "${IMAGE}"

ENV_VARS="GCP_PROJECT_ID=${PROJECT_ID},GOOGLE_CLOUD_PROJECT=${PROJECT_ID},STORE_DRIVER=firestore,FIRESTORE_COLLECTION=${FIRESTORE_COLLECTION:-sms_otp_challenges},SMS_PROVIDER=${SMS_PROVIDER:-twilio},AWS_REGION=${AWS_REGION:-},AWS_SNS_SMS_TYPE=${AWS_SNS_SMS_TYPE:-Transactional},AWS_SNS_SENDER_ID=${AWS_SNS_SENDER_ID:-},OTP_MESSAGE_TEMPLATE=${OTP_MESSAGE_TEMPLATE:-Your verification code is {{CODE}}. It expires in {{MINUTES}} minutes.}"

SET_SECRETS=""
for mapping in \
  "SMS_OTP_SERVICE_API_TOKEN=sms-otp-service-api-token" \
  "SMS_OTP_SECRET=sms-otp-secret" \
  "SMS_MFA_SESSION_SECRET=sms-mfa-session-secret" \
  "TWILIO_ACCOUNT_SID=twilio-account-sid" \
  "TWILIO_API_KEY_SID=twilio-api-key-sid" \
  "TWILIO_API_KEY_SECRET=twilio-api-key-secret" \
  "TWILIO_MESSAGING_SERVICE_SID=twilio-messaging-service-sid" \
  "AWS_ACCESS_KEY_ID=aws-sns-access-key-id" \
  "AWS_SECRET_ACCESS_KEY=aws-sns-secret-access-key" \
  "AWS_SESSION_TOKEN=aws-sns-session-token"
do
  env_name="${mapping%%=*}"
  secret_name="${mapping#*=}"
  if gcloud secrets describe "${secret_name}" --project="${PROJECT_ID}" >/dev/null 2>&1; then
    if [ -n "$SET_SECRETS" ]; then
      SET_SECRETS="${SET_SECRETS},"
    fi
    SET_SECRETS="${SET_SECRETS}${env_name}=${secret_name}:latest"
  fi
done

DEPLOY_ARGS=(
  run deploy "${SERVICE}"
  --project "${PROJECT_ID}"
  --region "${REGION}"
  --image "${IMAGE}"
  --allow-unauthenticated
  --set-env-vars "${ENV_VARS}"
)

if [ -n "$SERVICE_ACCOUNT" ]; then
  DEPLOY_ARGS+=(--service-account "$SERVICE_ACCOUNT")
fi

if [ -n "$SET_SECRETS" ]; then
  DEPLOY_ARGS+=(--set-secrets "$SET_SECRETS")
fi

gcloud "${DEPLOY_ARGS[@]}"

URL="$(gcloud run services describe "${SERVICE}" --region "${REGION}" --project "${PROJECT_ID}" --format='value(status.url)')"
echo "Service URL: ${URL}"
curl -fsS "${URL}/health"
echo
