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

require AWS_REGION

FUNCTION_NAME="${FUNCTION_NAME:-sms-otp-factor-service}"
ECR_REPOSITORY="${ECR_REPOSITORY:-sms-otp-factor-service-lambda}"
DYNAMODB_TABLE="${DYNAMODB_TABLE:-sms-otp-challenges}"
ROLE_NAME="${ROLE_NAME:-${FUNCTION_NAME}-lambda-execution}"
TAG="${TAG:-$(git rev-parse --short HEAD 2>/dev/null || date +%Y%m%d%H%M%S)}"
SMS_PROVIDER="${SMS_PROVIDER:-amazon_sns}"
AWS_SNS_SMS_TYPE="${AWS_SNS_SMS_TYPE:-Transactional}"
OTP_MESSAGE_TEMPLATE="${OTP_MESSAGE_TEMPLATE:-Your verification code is {{CODE}}. It expires in {{MINUTES}} minutes.}"
MEMORY_SIZE="${MEMORY_SIZE:-512}"
TIMEOUT="${TIMEOUT:-30}"

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

if ! aws ecr describe-repositories --repository-names "$ECR_REPOSITORY" --region "$AWS_REGION" >/dev/null 2>&1; then
  aws ecr create-repository \
    --repository-name "$ECR_REPOSITORY" \
    --image-scanning-configuration scanOnPush=true \
    --encryption-configuration encryptionType=AES256 \
    --region "$AWS_REGION" >/dev/null
fi

echo "Testing and building linux/amd64 Lambda container..."
go test ./...
mkdir -p .build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-w -s" -o .build/server ./cmd/server
docker pull --platform linux/amd64 public.ecr.aws/awsguru/aws-lambda-adapter:0.9.1 >/dev/null
ADAPTER_CONTAINER="$(docker create --platform linux/amd64 public.ecr.aws/awsguru/aws-lambda-adapter:0.9.1 /lambda-adapter)"
docker cp "${ADAPTER_CONTAINER}:/lambda-adapter" .build/lambda-adapter
docker rm "${ADAPTER_CONTAINER}" >/dev/null
chmod +x .build/lambda-adapter

aws ecr get-login-password --region "$AWS_REGION" | docker login --username AWS --password-stdin "${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
docker build --platform linux/amd64 -f Dockerfile.lambda -t "$IMAGE" .
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

cat > "$TMP_DIR/lambda-trust.json" <<'JSON'
{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"lambda.amazonaws.com"},"Action":"sts:AssumeRole"}]}
JSON

if ! aws iam get-role --role-name "$ROLE_NAME" >/dev/null 2>&1; then
  aws iam create-role --role-name "$ROLE_NAME" --assume-role-policy-document "file://${TMP_DIR}/lambda-trust.json" >/dev/null
fi
aws iam attach-role-policy --role-name "$ROLE_NAME" --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole >/dev/null || true
ROLE_ARN="$(aws iam get-role --role-name "$ROLE_NAME" --query Role.Arn --output text)"

cat > "$TMP_DIR/runtime-policy.json" <<JSON
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
      "Action": ["sns:Publish"],
      "Resource": "*"
    }
  ]
}
JSON
aws iam put-role-policy --role-name "$ROLE_NAME" --policy-name "${FUNCTION_NAME}-runtime" --policy-document "file://${TMP_DIR}/runtime-policy.json" >/dev/null

cat > "$TMP_DIR/ecr-policy.json" <<JSON
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "LambdaECRImageRetrieval",
      "Effect": "Allow",
      "Principal": {"Service": "lambda.amazonaws.com"},
      "Action": ["ecr:BatchGetImage", "ecr:GetDownloadUrlForLayer"],
      "Condition": {
        "StringLike": {
          "aws:sourceArn": "arn:aws:lambda:${AWS_REGION}:${ACCOUNT_ID}:function:${FUNCTION_NAME}"
        }
      }
    }
  ]
}
JSON
aws ecr set-repository-policy --repository-name "$ECR_REPOSITORY" --policy-text "file://${TMP_DIR}/ecr-policy.json" --region "$AWS_REGION" >/dev/null

python3 - "$TMP_DIR/env.json" "$DYNAMODB_TABLE" "$AWS_REGION" "$SMS_PROVIDER" "$AWS_SNS_SMS_TYPE" "${AWS_SNS_SENDER_ID:-}" "$OTP_MESSAGE_TEMPLATE" <<'PY'
import json
import os
import sys

path, table, region, sms_provider, sms_type, sender_id, template = sys.argv[1:]
env = {
    "STORE_DRIVER": "dynamodb",
    "DYNAMODB_TABLE": table,
    "SMS_PROVIDER": sms_provider,
    "AWS_SNS_SMS_TYPE": sms_type,
    "AWS_SNS_SENDER_ID": sender_id,
    "OTP_MESSAGE_TEMPLATE": template,
    "SMS_OTP_SERVICE_API_TOKEN": os.environ["SMS_OTP_SERVICE_API_TOKEN"],
    "SMS_OTP_SECRET": os.environ["SMS_OTP_SECRET"],
    "SMS_PHONE_HASH_SECRET": os.environ["SMS_PHONE_HASH_SECRET"],
    "SMS_MFA_SESSION_SECRET": os.environ["SMS_MFA_SESSION_SECRET"],
}
if sms_provider == "twilio":
    env.update({
        "TWILIO_ACCOUNT_SID": os.environ["TWILIO_ACCOUNT_SID"],
        "TWILIO_API_KEY_SID": os.environ["TWILIO_API_KEY_SID"],
        "TWILIO_API_KEY_SECRET": os.environ["TWILIO_API_KEY_SECRET"],
        "TWILIO_MESSAGING_SERVICE_SID": os.environ["TWILIO_MESSAGING_SERVICE_SID"],
    })
with open(path, "w", encoding="utf-8") as handle:
    json.dump({"Variables": env}, handle)
PY

if aws lambda get-function --function-name "$FUNCTION_NAME" --region "$AWS_REGION" >/dev/null 2>&1; then
  aws lambda update-function-code --function-name "$FUNCTION_NAME" --image-uri "$IMAGE" --region "$AWS_REGION" >/dev/null
  aws lambda wait function-updated --function-name "$FUNCTION_NAME" --region "$AWS_REGION"
  aws lambda update-function-configuration \
    --function-name "$FUNCTION_NAME" \
    --role "$ROLE_ARN" \
    --memory-size "$MEMORY_SIZE" \
    --timeout "$TIMEOUT" \
    --environment "file://${TMP_DIR}/env.json" \
    --region "$AWS_REGION" >/dev/null
  aws lambda wait function-updated --function-name "$FUNCTION_NAME" --region "$AWS_REGION"
else
  sleep 10
  aws lambda create-function \
    --function-name "$FUNCTION_NAME" \
    --package-type Image \
    --code ImageUri="$IMAGE" \
    --role "$ROLE_ARN" \
    --architectures x86_64 \
    --memory-size "$MEMORY_SIZE" \
    --timeout "$TIMEOUT" \
    --environment "file://${TMP_DIR}/env.json" \
    --region "$AWS_REGION" >/dev/null
  aws lambda wait function-active --function-name "$FUNCTION_NAME" --region "$AWS_REGION"
fi

if ! aws lambda get-function-url-config --function-name "$FUNCTION_NAME" --region "$AWS_REGION" >/dev/null 2>&1; then
  aws lambda create-function-url-config \
    --function-name "$FUNCTION_NAME" \
    --auth-type NONE \
    --cors AllowOrigins='["*"]',AllowMethods='["GET","POST"]',AllowHeaders='["authorization","content-type"]' \
    --region "$AWS_REGION" >/dev/null
fi
aws lambda add-permission \
  --function-name "$FUNCTION_NAME" \
  --statement-id FunctionURLAllowPublicAccess \
  --action lambda:InvokeFunctionUrl \
  --principal "*" \
  --function-url-auth-type NONE \
  --region "$AWS_REGION" >/dev/null 2>&1 || true
aws lambda add-permission \
  --function-name "$FUNCTION_NAME" \
  --statement-id FunctionURLInvokeAllowPublicAccess \
  --action lambda:InvokeFunction \
  --principal "*" \
  --invoked-via-function-url \
  --region "$AWS_REGION" >/dev/null 2>&1 || true

URL="$(aws lambda get-function-url-config --function-name "$FUNCTION_NAME" --region "$AWS_REGION" --query FunctionUrl --output text)"
echo "AWS Lambda free-tier installation complete."
echo "Function name: ${FUNCTION_NAME}"
echo "Image: ${IMAGE}"
echo "DynamoDB table: ${DYNAMODB_TABLE}"
echo "Function URL: ${URL}"
