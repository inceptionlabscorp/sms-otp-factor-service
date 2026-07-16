#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

require() {
  local name="$1"
  if [ -z "${!name:-}" ]; then
    echo "Missing required environment variable: ${name}" >&2
    exit 1
  fi
}

ensure_secret() {
  local env_name="$1"
  local secret_name="$2"
  require "$env_name"
  if ! gcloud secrets describe "$secret_name" --project "$GCP_PROJECT_ID" >/dev/null 2>&1; then
    gcloud secrets create "$secret_name" --project "$GCP_PROJECT_ID" --replication-policy automatic
  fi
  printf '%s' "${!env_name}" | gcloud secrets versions add "$secret_name" --project "$GCP_PROJECT_ID" --data-file=-
  gcloud secrets add-iam-policy-binding "$secret_name" \
    --project "$GCP_PROJECT_ID" \
    --member "serviceAccount:${SERVICE_ACCOUNT_EMAIL}" \
    --role roles/secretmanager.secretAccessor >/dev/null
}

require GCP_PROJECT_ID

REGION="${REGION:-us-central1}"
ARTIFACT_REGION="${ARTIFACT_REGION:-$REGION}"
SERVICE="${SERVICE:-sms-otp-factor-service}"
REPOSITORY="${REPOSITORY:-sms-otp}"
TAG="${TAG:-$(git rev-parse --short HEAD 2>/dev/null || date +%Y%m%d%H%M%S)}"
FIRESTORE_LOCATION="${FIRESTORE_LOCATION:-nam5}"
FIRESTORE_COLLECTION="${FIRESTORE_COLLECTION:-sms_otp_challenges}"
SMS_PROVIDER="${SMS_PROVIDER:-twilio}"
INGRESS="${INGRESS:-internal-and-cloud-load-balancing}"
ALLOW_UNAUTHENTICATED="${ALLOW_UNAUTHENTICATED:-false}"
OTP_MESSAGE_TEMPLATE="${OTP_MESSAGE_TEMPLATE:-Your verification code is {{CODE}}. It expires in {{MINUTES}} minutes.}"
SERVICE_ACCOUNT_NAME="${SERVICE_ACCOUNT_NAME:-sms-otp-factor-service}"
SERVICE_ACCOUNT_EMAIL="${SERVICE_ACCOUNT_NAME}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"
IMAGE="${ARTIFACT_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/${REPOSITORY}/${SERVICE}:${TAG}"

for key in SMS_OTP_SERVICE_API_TOKEN SMS_OTP_SECRET SMS_PHONE_HASH_SECRET SMS_MFA_SESSION_SECRET; do
  require "$key"
done

case "$SMS_PROVIDER" in
  twilio)
    for key in TWILIO_ACCOUNT_SID TWILIO_API_KEY_SID TWILIO_API_KEY_SECRET TWILIO_MESSAGING_SERVICE_SID; do
      require "$key"
    done
    ;;
  amazon_sns|sns|amazon-simple-notification-service)
    require AWS_REGION
    for key in AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY; do
      require "$key"
    done
    ;;
  *)
    echo "SMS_PROVIDER must be twilio or amazon_sns" >&2
    exit 1
    ;;
esac

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

for api in run.googleapis.com artifactregistry.googleapis.com secretmanager.googleapis.com firestore.googleapis.com iam.googleapis.com; do
  gcloud services enable "$api" --project "$GCP_PROJECT_ID" >/dev/null
done

if ! gcloud iam service-accounts describe "$SERVICE_ACCOUNT_EMAIL" --project "$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud iam service-accounts create "$SERVICE_ACCOUNT_NAME" \
    --project "$GCP_PROJECT_ID" \
    --display-name "SMS OTP Factor Service"
fi

gcloud projects add-iam-policy-binding "$GCP_PROJECT_ID" \
  --member "serviceAccount:${SERVICE_ACCOUNT_EMAIL}" \
  --role roles/datastore.user >/dev/null

if ! gcloud artifacts repositories describe "$REPOSITORY" --location "$ARTIFACT_REGION" --project "$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud artifacts repositories create "$REPOSITORY" \
    --project "$GCP_PROJECT_ID" \
    --location "$ARTIFACT_REGION" \
    --repository-format docker \
    --description "SMS OTP Factor Service containers"
fi

if ! gcloud firestore databases describe --database="(default)" --project "$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud firestore databases create --database="(default)" --location="$FIRESTORE_LOCATION" --project "$GCP_PROJECT_ID"
fi

echo "Testing and building linux/amd64 binary..."
go test ./...
mkdir -p .build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-w -s" -o .build/server ./cmd/server

gcloud auth configure-docker "${ARTIFACT_REGION}-docker.pkg.dev" --quiet
docker build --platform linux/amd64 -t "$IMAGE" .
docker push "$IMAGE"

ensure_secret SMS_OTP_SERVICE_API_TOKEN sms-otp-service-api-token
ensure_secret SMS_OTP_SECRET sms-otp-secret
ensure_secret SMS_PHONE_HASH_SECRET sms-phone-hash-secret
ensure_secret SMS_MFA_SESSION_SECRET sms-mfa-session-secret

SET_SECRETS="SMS_OTP_SERVICE_API_TOKEN=sms-otp-service-api-token:latest,SMS_OTP_SECRET=sms-otp-secret:latest,SMS_PHONE_HASH_SECRET=sms-phone-hash-secret:latest,SMS_MFA_SESSION_SECRET=sms-mfa-session-secret:latest"

if [ "$SMS_PROVIDER" = "twilio" ]; then
  ensure_secret TWILIO_ACCOUNT_SID twilio-account-sid
  ensure_secret TWILIO_API_KEY_SID twilio-api-key-sid
  ensure_secret TWILIO_API_KEY_SECRET twilio-api-key-secret
  ensure_secret TWILIO_MESSAGING_SERVICE_SID twilio-messaging-service-sid
  SET_SECRETS="${SET_SECRETS},TWILIO_ACCOUNT_SID=twilio-account-sid:latest,TWILIO_API_KEY_SID=twilio-api-key-sid:latest,TWILIO_API_KEY_SECRET=twilio-api-key-secret:latest,TWILIO_MESSAGING_SERVICE_SID=twilio-messaging-service-sid:latest"
else
  ensure_secret AWS_ACCESS_KEY_ID aws-sns-access-key-id
  ensure_secret AWS_SECRET_ACCESS_KEY aws-sns-secret-access-key
  if [ -n "${AWS_SESSION_TOKEN:-}" ]; then
    ensure_secret AWS_SESSION_TOKEN aws-sns-session-token
    SET_SECRETS="${SET_SECRETS},AWS_SESSION_TOKEN=aws-sns-session-token:latest"
  fi
  SET_SECRETS="${SET_SECRETS},AWS_ACCESS_KEY_ID=aws-sns-access-key-id:latest,AWS_SECRET_ACCESS_KEY=aws-sns-secret-access-key:latest"
fi

ENV_VARS="GCP_PROJECT_ID=${GCP_PROJECT_ID},GOOGLE_CLOUD_PROJECT=${GCP_PROJECT_ID},STORE_DRIVER=firestore,FIRESTORE_COLLECTION=${FIRESTORE_COLLECTION},SMS_PROVIDER=${SMS_PROVIDER},AWS_REGION=${AWS_REGION:-},AWS_SNS_SMS_TYPE=${AWS_SNS_SMS_TYPE:-Transactional},AWS_SNS_SENDER_ID=${AWS_SNS_SENDER_ID:-},OTP_MESSAGE_TEMPLATE=${OTP_MESSAGE_TEMPLATE}"

DEPLOY_ARGS=(
  run deploy "$SERVICE"
  --project "$GCP_PROJECT_ID"
  --region "$REGION"
  --image "$IMAGE"
  --service-account "$SERVICE_ACCOUNT_EMAIL"
  --ingress "$INGRESS"
  --set-env-vars "$ENV_VARS"
  --set-secrets "$SET_SECRETS"
)

if [ "$ALLOW_UNAUTHENTICATED" = "true" ]; then
  DEPLOY_ARGS+=(--allow-unauthenticated)
else
  DEPLOY_ARGS+=(--no-allow-unauthenticated)
fi

gcloud "${DEPLOY_ARGS[@]}"

URL="$(gcloud run services describe "$SERVICE" --region "$REGION" --project "$GCP_PROJECT_ID" --format='value(status.url)')"
echo "GCP installation complete."
echo "Cloud Run URL: ${URL}"
echo "Ingress: ${INGRESS}"
