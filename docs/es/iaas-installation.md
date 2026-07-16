# Scripts de Instalacion IaaS

English: [IaaS installation scripts](../iaas-installation.md)

El repositorio incluye scripts de instalacion por proveedor para despliegues productivos.

Estos scripts provisionan recursos cloud reales y despliegan la imagen de contenedor. No guardan valores secretos en el repositorio.

## GCP

Script:

```bash
./scripts/install-gcp.sh
```

Recursos provisionados:

- Repositorio Docker en Artifact Registry.
- Base Firestore para challenges OTP.
- Secretos y versiones en Secret Manager.
- Service account dedicada.
- Servicio Cloud Run.
- Bindings IAM de minimo privilegio para Firestore y Secret Manager.

Variables requeridas:

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

Defaults seguros:

- `INGRESS=internal-and-cloud-load-balancing`
- `ALLOW_UNAUTHENTICATED=false`
- `STORE_DRIVER=firestore`

Para exponer el servicio publicamente en pruebas controladas:

```bash
export ALLOW_UNAUTHENTICATED=true
export INGRESS=all
```

La API sigue exigiendo `Authorization: Bearer <SMS_OTP_SERVICE_API_TOKEN>` en `/v1/*`.

## AWS

Script:

```bash
./scripts/install-aws.sh
```

Recursos provisionados:

- Repositorio ECR con scan-on-push.
- Tabla DynamoDB con billing on-demand, cifrado server-side y TTL sobre `expires_at`.
- Secretos en Secrets Manager.
- Rol de acceso ECR para App Runner.
- Rol de instancia para App Runner.
- Politica IAM para DynamoDB, Secrets Manager y SNS publish.
- Servicio App Runner.

Variables requeridas para Amazon SNS:

```bash
export AWS_REGION="us-east-1"
export SMS_PROVIDER="amazon_sns"
export SMS_OTP_SERVICE_API_TOKEN="32+ character random value"
export SMS_OTP_SECRET="32+ character random value"
export SMS_PHONE_HASH_SECRET="32+ character random value"
export SMS_MFA_SESSION_SECRET="32+ character random value"
```

Variables requeridas para Twilio en AWS:

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

Defaults seguros:

- `STORE_DRIVER=dynamodb`
- Amazon SNS usa el role de instancia App Runner en vez de access keys estaticas.
- Los secretos se inyectan desde Secrets Manager hacia App Runner.

## Notas Operativas

- Ejecuta scripts desde una estacion o runner CI con credenciales cloud aprobadas.
- Revisa permisos IAM generados antes de usarlos en produccion regulada.
- Prefiere ingress privado o un gateway service-to-service delante de Cloud Run/App Runner.
- Rota service token y secretos HMAC con el proceso de key management de tu organizacion.
- Mantén SMS OTP como factor step-up restringido; usa MFA phishing-resistant para flujos privilegiados y de alto riesgo.
