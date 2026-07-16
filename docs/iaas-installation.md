# IaaS Installation Scripts

Español: [Scripts de Instalacion IaaS](es/iaas-installation.md)

The repository includes provider installation scripts for production-style deployments.

These scripts provision real cloud resources and deploy the container image. They do not store secret values in the repository.

## GCP

Script:

```bash
./scripts/install-gcp.sh
```

Provisioned resources:

- Artifact Registry Docker repository.
- Firestore database for OTP challenges.
- Secret Manager secrets and versions.
- Dedicated service account.
- Cloud Run service.
- Least-privilege IAM bindings for Firestore and Secret Manager.

Required variables:

```bash
export GCP_PROJECT_ID="example-gcp-project"
export REGION="us-central1"
export SMS_PROVIDER="twilio"
export SMS_OTP_SERVICE_API_TOKEN="32+ character random value"
export SMS_OTP_SECRET="32+ character random value"
export SMS_PHONE_HASH_SECRET="32+ character random value"
export SMS_MFA_SESSION_SECRET="32+ character random value"
export TWILIO_ACCOUNT_SID="AC..."
export TWILIO_API_KEY_SID="SK..."
export TWILIO_API_KEY_SECRET="..."
export TWILIO_MESSAGING_SERVICE_SID="MG..."
```

Secure defaults:

- `INGRESS=internal-and-cloud-load-balancing`
- `ALLOW_UNAUTHENTICATED=false`
- `STORE_DRIVER=firestore`

To expose the service publicly for controlled testing:

```bash
export ALLOW_UNAUTHENTICATED=true
export INGRESS=all
```

The API still requires `Authorization: Bearer <SMS_OTP_SERVICE_API_TOKEN>` on `/v1/*`.

## AWS

Script:

```bash
./scripts/install-aws.sh
```

Provisioned resources:

- ECR repository with scan-on-push.
- DynamoDB table with on-demand billing, default server-side encryption, and TTL on `expires_at`.
- Secrets Manager secrets.
- App Runner ECR access role.
- App Runner instance role.
- IAM policy for DynamoDB, Secrets Manager, and SNS publish.
- App Runner service.

Preflight requirement: AWS App Runner must be enabled/subscribed for the target account and region. The installer checks this before creating resources.

Required variables for Amazon SNS:

```bash
export AWS_REGION="us-east-1"
export SMS_PROVIDER="amazon_sns"
export SMS_OTP_SERVICE_API_TOKEN="32+ character random value"
export SMS_OTP_SECRET="32+ character random value"
export SMS_PHONE_HASH_SECRET="32+ character random value"
export SMS_MFA_SESSION_SECRET="32+ character random value"
```

Required variables for Twilio on AWS:

```bash
export AWS_REGION="us-east-1"
export SMS_PROVIDER="twilio"
export SMS_OTP_SERVICE_API_TOKEN="32+ character random value"
export SMS_OTP_SECRET="32+ character random value"
export SMS_PHONE_HASH_SECRET="32+ character random value"
export SMS_MFA_SESSION_SECRET="32+ character random value"
export TWILIO_ACCOUNT_SID="AC..."
export TWILIO_API_KEY_SID="SK..."
export TWILIO_API_KEY_SECRET="..."
export TWILIO_MESSAGING_SERVICE_SID="MG..."
```

Secure defaults:

- `STORE_DRIVER=dynamodb`
- Amazon SNS uses the App Runner instance role rather than static AWS access keys.
- Secrets are injected from Secrets Manager into App Runner.

## Operational Notes

- Run scripts from a secured workstation or CI runner with approved cloud credentials.
- Review generated IAM permissions before using in regulated production.
- Prefer private ingress or a service-to-service gateway in front of Cloud Run/App Runner.
- Rotate the service token and HMAC secrets using your organization's key-management process.
- Keep SMS OTP as a restricted step-up factor; use phishing-resistant MFA for privileged and high-risk workflows.
