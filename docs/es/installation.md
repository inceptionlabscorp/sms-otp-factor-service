# Instalacion

English: [Installation](../installation.md)

## Requisitos

- Go 1.22 o superior.
- Un backend confiable que pueda llamar este servicio por HTTP.
- Un proveedor SMS: Twilio Messaging Service o Amazon Simple Notification Service.
- Un store de challenges:
  - `memory` solo para desarrollo local y pruebas.
  - `firestore` para despliegues productivos en Google Cloud.
  - `dynamodb` para despliegues productivos en AWS.

## Instalacion Local

```bash
git clone https://github.com/inceptionlabscorp/sms-otp-factor-service.git
cd sms-otp-factor-service
cp .env.example .env
```

Configura las variables de `.env.example` y ejecuta:

```bash
go test ./...
go run ./cmd/server
```

El servicio escucha en `PORT`; el valor por defecto es `8080`.

## Docker

```bash
docker build -t sms-otp-factor-service:local .
docker run --rm -p 8080:8080 \
  --env-file .env \
  sms-otp-factor-service:local
```

## Google Cloud Run

Para instalar en GCP provisionando Artifact Registry, Firestore, Secret Manager, IAM y Cloud Run:

```bash
./scripts/install-gcp.sh
```

El script legacy `deploy-prod.sh` solo construye y despliega Cloud Run contra infraestructura existente:

```bash
export GCP_PROJECT_ID="example-gcp-project"
export REGION="us-central1"
export SERVICE="sms-otp-factor-service"
export SERVICE_ACCOUNT="sms-otp-factor-service@example-gcp-project.iam.gserviceaccount.com"
./scripts/deploy-prod.sh
```

`GCP_PROJECT_ID` es obligatorio a proposito. El script no contiene defaults productivos.

## AWS App Runner

Para instalar en AWS provisionando ECR, DynamoDB, Secrets Manager, IAM y App Runner:

```bash
./scripts/install-aws.sh
```

Los despliegues productivos en AWS usan `STORE_DRIVER=dynamodb`.

## Checklist Productivo

- Usa `STORE_DRIVER=firestore` en GCP o `STORE_DRIVER=dynamodb` en AWS.
- Guarda secretos en Secret Manager, AWS Secrets Manager, Vault o equivalente.
- Usa valores aleatorios separados para `SMS_OTP_SECRET`, `SMS_PHONE_HASH_SECRET` y `SMS_MFA_SESSION_SECRET`.
- Restringe acceso de red al backend confiable cuando sea posible.
- Rota `SMS_OTP_SERVICE_API_TOKEN` y secretos HMAC con un runbook operativo.
- Mantén credenciales de proveedor limitadas a envio SMS.
- Confirma que logs no incluyan OTP, tokens, telefonos completos ni secretos.
