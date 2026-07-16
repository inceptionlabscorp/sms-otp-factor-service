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

put_secret() {
  local env_name="$1"
  local secret_name="$2"
  require "$env_name"
  local value="${!env_name}"
  if aws secretsmanager describe-secret --secret-id "$secret_name" --region "$AWS_REGION" >/dev/null 2>&1; then
    aws secretsmanager put-secret-value --secret-id "$secret_name" --secret-string "$value" --region "$AWS_REGION" >/dev/null
  else
    aws secretsmanager create-secret --name "$secret_name" --secret-string "$value" --region "$AWS_REGION" >/dev/null
  fi
  aws secretsmanager describe-secret --secret-id "$secret_name" --region "$AWS_REGION" --query ARN --output text
}

require AWS_REGION

SERVICE_NAME="${SERVICE_NAME:-sms-otp-factor-service}"
ECR_REPOSITORY="${ECR_REPOSITORY:-sms-otp-factor-service}"
DYNAMODB_TABLE="${DYNAMODB_TABLE:-sms-otp-challenges}"
TAG="${TAG:-$(git rev-parse --short HEAD 2>/dev/null || date +%Y%m%d%H%M%S)}"
SMS_PROVIDER="${SMS_PROVIDER:-amazon_sns}"
AWS_SNS_SMS_TYPE="${AWS_SNS_SMS_TYPE:-Transactional}"
OTP_MESSAGE_TEMPLATE="${OTP_MESSAGE_TEMPLATE:-Your verification code is {{CODE}}. It expires in {{MINUTES}} minutes.}"
CPU="${CPU:-0.25 vCPU}"
MEMORY="${MEMORY:-0.5 GB}"

for key in SMS_OTP_SERVICE_API_TOKEN SMS_OTP_SECRET SMS_PHONE_HASH_SECRET SMS_MFA_SESSION_SECRET; do
  require "$key"
done

case "$SMS_PROVIDER" in
  amazon_sns|sns|amazon-simple-notification-service)
    ;;
  twilio)
    for key in TWILIO_ACCOUNT_SID TWILIO_API_KEY_SID TWILIO_API_KEY_SECRET TWILIO_MESSAGING_SERVICE_SID; do
      require "$key"
    done
    ;;
  *)
    echo "SMS_PROVIDER must be amazon_sns or twilio" >&2
    exit 1
    ;;
esac

ACCOUNT_ID="$(aws sts get-caller-identity --query Account --output text)"
IMAGE="${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY}:${TAG}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

if ! aws apprunner list-services --region "$AWS_REGION" >/dev/null 2>&1; then
  echo "AWS App Runner is not available for this account/region. Enable or subscribe to App Runner before running this installer." >&2
  exit 1
fi

if ! aws ecr describe-repositories --repository-names "$ECR_REPOSITORY" --region "$AWS_REGION" >/dev/null 2>&1; then
  aws ecr create-repository \
    --repository-name "$ECR_REPOSITORY" \
    --image-scanning-configuration scanOnPush=true \
    --encryption-configuration encryptionType=AES256 \
    --region "$AWS_REGION" >/dev/null
fi

echo "Testing and building linux/amd64 binary..."
go test ./...
mkdir -p .build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-w -s" -o .build/server ./cmd/server

aws ecr get-login-password --region "$AWS_REGION" | docker login --username AWS --password-stdin "${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
docker build --platform linux/amd64 -t "$IMAGE" .
docker push "$IMAGE"

if ! aws dynamodb describe-table --table-name "$DYNAMODB_TABLE" --region "$AWS_REGION" >/dev/null 2>&1; then
  aws dynamodb create-table \
    --table-name "$DYNAMODB_TABLE" \
    --attribute-definitions AttributeName=challenge_key,AttributeType=S \
    --key-schema AttributeName=challenge_key,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    --region "$AWS_REGION" >/dev/null
  aws dynamodb wait table-exists --table-name "$DYNAMODB_TABLE" --region "$AWS_REGION"
  aws dynamodb update-time-to-live \
    --table-name "$DYNAMODB_TABLE" \
    --time-to-live-specification Enabled=true,AttributeName=expires_at \
    --region "$AWS_REGION" >/dev/null
fi

SERVICE_TOKEN_SECRET_ARN="$(put_secret SMS_OTP_SERVICE_API_TOKEN "${SERVICE_NAME}/sms-otp-service-api-token")"
OTP_SECRET_ARN="$(put_secret SMS_OTP_SECRET "${SERVICE_NAME}/sms-otp-secret")"
PHONE_HASH_SECRET_ARN="$(put_secret SMS_PHONE_HASH_SECRET "${SERVICE_NAME}/sms-phone-hash-secret")"
SESSION_SECRET_ARN="$(put_secret SMS_MFA_SESSION_SECRET "${SERVICE_NAME}/sms-mfa-session-secret")"

TWILIO_ACCOUNT_SECRET_ARN=""
TWILIO_KEY_SID_SECRET_ARN=""
TWILIO_KEY_SECRET_SECRET_ARN=""
TWILIO_MESSAGING_SECRET_ARN=""
if [ "$SMS_PROVIDER" = "twilio" ]; then
  TWILIO_ACCOUNT_SECRET_ARN="$(put_secret TWILIO_ACCOUNT_SID "${SERVICE_NAME}/twilio-account-sid")"
  TWILIO_KEY_SID_SECRET_ARN="$(put_secret TWILIO_API_KEY_SID "${SERVICE_NAME}/twilio-api-key-sid")"
  TWILIO_KEY_SECRET_SECRET_ARN="$(put_secret TWILIO_API_KEY_SECRET "${SERVICE_NAME}/twilio-api-key-secret")"
  TWILIO_MESSAGING_SECRET_ARN="$(put_secret TWILIO_MESSAGING_SERVICE_SID "${SERVICE_NAME}/twilio-messaging-service-sid")"
fi

ACCESS_ROLE_NAME="${SERVICE_NAME}-apprunner-ecr-access"
INSTANCE_ROLE_NAME="${SERVICE_NAME}-apprunner-instance"

cat > "$TMP_DIR/apprunner-access-trust.json" <<'JSON'
{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"build.apprunner.amazonaws.com"},"Action":"sts:AssumeRole"}]}
JSON

cat > "$TMP_DIR/apprunner-instance-trust.json" <<'JSON'
{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"tasks.apprunner.amazonaws.com"},"Action":"sts:AssumeRole"}]}
JSON

if ! aws iam get-role --role-name "$ACCESS_ROLE_NAME" >/dev/null 2>&1; then
  aws iam create-role --role-name "$ACCESS_ROLE_NAME" --assume-role-policy-document "file://${TMP_DIR}/apprunner-access-trust.json" >/dev/null
fi
aws iam attach-role-policy --role-name "$ACCESS_ROLE_NAME" --policy-arn arn:aws:iam::aws:policy/service-role/AWSAppRunnerServicePolicyForECRAccess >/dev/null || true
ACCESS_ROLE_ARN="$(aws iam get-role --role-name "$ACCESS_ROLE_NAME" --query Role.Arn --output text)"

if ! aws iam get-role --role-name "$INSTANCE_ROLE_NAME" >/dev/null 2>&1; then
  aws iam create-role --role-name "$INSTANCE_ROLE_NAME" --assume-role-policy-document "file://${TMP_DIR}/apprunner-instance-trust.json" >/dev/null
fi
INSTANCE_ROLE_ARN="$(aws iam get-role --role-name "$INSTANCE_ROLE_NAME" --query Role.Arn --output text)"

cat > "$TMP_DIR/instance-policy.json" <<JSON
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:DeleteItem"],
      "Resource": "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/${DYNAMODB_TABLE}"
    },
    {
      "Effect": "Allow",
      "Action": ["secretsmanager:GetSecretValue"],
      "Resource": [
        "${SERVICE_TOKEN_SECRET_ARN}",
        "${OTP_SECRET_ARN}",
        "${PHONE_HASH_SECRET_ARN}",
        "${SESSION_SECRET_ARN}"$(if [ "$SMS_PROVIDER" = "twilio" ]; then printf ',\n        "%s",\n        "%s",\n        "%s",\n        "%s"' "$TWILIO_ACCOUNT_SECRET_ARN" "$TWILIO_KEY_SID_SECRET_ARN" "$TWILIO_KEY_SECRET_SECRET_ARN" "$TWILIO_MESSAGING_SECRET_ARN"; fi)
      ]
    },
    {
      "Effect": "Allow",
      "Action": ["sns:Publish"],
      "Resource": "*"
    }
  ]
}
JSON
aws iam put-role-policy --role-name "$INSTANCE_ROLE_NAME" --policy-name "${SERVICE_NAME}-runtime" --policy-document "file://${TMP_DIR}/instance-policy.json" >/dev/null

python3 - "$TMP_DIR/service.json" "$IMAGE" "$ACCESS_ROLE_ARN" "$INSTANCE_ROLE_ARN" "$SERVICE_NAME" "$DYNAMODB_TABLE" "$AWS_REGION" "$SMS_PROVIDER" "$AWS_SNS_SMS_TYPE" "${AWS_SNS_SENDER_ID:-}" "$OTP_MESSAGE_TEMPLATE" "$CPU" "$MEMORY" "$SERVICE_TOKEN_SECRET_ARN" "$OTP_SECRET_ARN" "$PHONE_HASH_SECRET_ARN" "$SESSION_SECRET_ARN" "$TWILIO_ACCOUNT_SECRET_ARN" "$TWILIO_KEY_SID_SECRET_ARN" "$TWILIO_KEY_SECRET_SECRET_ARN" "$TWILIO_MESSAGING_SECRET_ARN" <<'PY'
import json
import sys

(
    path, image, access_role, instance_role, service_name, table, region,
    sms_provider, sms_type, sender_id, template, cpu, memory,
    service_token, otp_secret, phone_hash_secret, session_secret,
    twilio_account, twilio_key_sid, twilio_key_secret, twilio_messaging,
) = sys.argv[1:]

env = {
    "STORE_DRIVER": "dynamodb",
    "DYNAMODB_TABLE": table,
    "AWS_REGION": region,
    "SMS_PROVIDER": sms_provider,
    "AWS_SNS_SMS_TYPE": sms_type,
    "AWS_SNS_SENDER_ID": sender_id,
    "OTP_MESSAGE_TEMPLATE": template,
}
secrets = {
    "SMS_OTP_SERVICE_API_TOKEN": service_token,
    "SMS_OTP_SECRET": otp_secret,
    "SMS_PHONE_HASH_SECRET": phone_hash_secret,
    "SMS_MFA_SESSION_SECRET": session_secret,
}
if sms_provider == "twilio":
    secrets.update({
        "TWILIO_ACCOUNT_SID": twilio_account,
        "TWILIO_API_KEY_SID": twilio_key_sid,
        "TWILIO_API_KEY_SECRET": twilio_key_secret,
        "TWILIO_MESSAGING_SERVICE_SID": twilio_messaging,
    })

payload = {
    "ServiceName": service_name,
    "SourceConfiguration": {
        "AuthenticationConfiguration": {"AccessRoleArn": access_role},
        "AutoDeploymentsEnabled": True,
        "ImageRepository": {
            "ImageIdentifier": image,
            "ImageRepositoryType": "ECR",
            "ImageConfiguration": {
                "Port": "8080",
                "RuntimeEnvironmentVariables": env,
                "RuntimeEnvironmentSecrets": secrets,
            },
        },
    },
    "InstanceConfiguration": {
        "Cpu": cpu,
        "Memory": memory,
        "InstanceRoleArn": instance_role,
    },
    "HealthCheckConfiguration": {
        "Protocol": "HTTP",
        "Path": "/health",
        "Interval": 10,
        "Timeout": 5,
        "HealthyThreshold": 1,
        "UnhealthyThreshold": 5,
    },
}
with open(path, "w", encoding="utf-8") as handle:
    json.dump(payload, handle)
PY

SERVICE_ARN="$(aws apprunner list-services --region "$AWS_REGION" --query "ServiceSummaryList[?ServiceName=='${SERVICE_NAME}'].ServiceArn | [0]" --output text)"
if [ "$SERVICE_ARN" = "None" ] || [ -z "$SERVICE_ARN" ]; then
  aws apprunner create-service --cli-input-json "file://${TMP_DIR}/service.json" --region "$AWS_REGION" >/dev/null
else
  python3 - "$TMP_DIR/service.json" "$TMP_DIR/update-service.json" "$SERVICE_ARN" <<'PY'
import json
import sys

source, target, arn = sys.argv[1:]
with open(source, encoding="utf-8") as handle:
    payload = json.load(handle)
payload["ServiceArn"] = arn
payload.pop("ServiceName", None)
with open(target, "w", encoding="utf-8") as handle:
    json.dump(payload, handle)
PY
  aws apprunner update-service --cli-input-json "file://${TMP_DIR}/update-service.json" --region "$AWS_REGION" >/dev/null
fi

echo "AWS installation requested."
echo "App Runner service: ${SERVICE_NAME}"
echo "Image: ${IMAGE}"
echo "DynamoDB table: ${DYNAMODB_TABLE}"
