# SMS OTP Factor Service

> Microservicio HTTP privado de OTP por SMS para el segundo factor: lo llama un backend de confianza, no el navegador ni la app.
>
> Private SMS OTP microservice for second-factor verification. A trusted backend calls it; browsers and mobile apps do not.

English index: [English](#english).

Este índice sigue la plantilla de Inception Labs. Cada dato sale de este repositorio. Lo que no está en el árbol figura como **PENDIENTE**.

## 1. Ficha

| Campo | Valor |
|---|---|
| Repositorio | [`inceptionlabscorp/sms-otp-factor-service`](https://github.com/inceptionlabscorp/sms-otp-factor-service) |
| Tipo | API (microservicio HTTP privado) |
| Estado | PENDIENTE |
| Cliente | PENDIENTE |
| Producto o sistema | SMS OTP Factor Service, bounded context `smsotp` |
| Owner técnico | PENDIENTE |
| URL de producción | PENDIENTE |
| URL de staging o prueba | PENDIENTE |
| Proyecto cloud | PENDIENTE |
| Repos relacionados | PENDIENTE |
| Versión actual | 0.5.6 ([`CHANGELOG.md`](CHANGELOG.md), [`docs/openapi.yaml`](docs/openapi.yaml)). No hay tags git. |
| Última revisión del README | 2026-10-21 |

## 2. Propósito y alcance

Resuelve el ciclo del factor OTP por SMS para un backend que ya autenticó al sujeto y ya decidió qué teléfono está autorizado: crear el challenge, enviar el SMS, verificar el código y firmar o validar un `mfa_token` de vida corta. El contrato público lo describen [`docs/architecture.md`](docs/architecture.md) y [`docs/api-integration.md`](docs/api-integration.md).

Límites de este repo:

- No hay UI, ni proveedor de identidad, ni política de titularidad del teléfono.
- El servicio no verifica la identidad de la persona ni decide si el teléfono le pertenece ([`docs/nist-security-profile.md`](docs/nist-security-profile.md), Non-Goals).
- Los clientes web o móviles no llaman al servicio. `/v1/*` exige el bearer de servicio.
- Los alias `/v1/bian/*` están alineados con lenguaje BIAN (Customer Access Entitlement). No son una API certificada BIAN ([`docs/openapi.yaml`](docs/openapi.yaml)).
- No expone ni consume MCP. Ver la sección 10.

## 3. Módulos y funcionalidades

| Módulo | Ruta o entrypoint | Qué hace | Datos (tablas o colecciones) | Estado |
|---|---|---|---|---|
| Arranque | [`cmd/server/main.go`](cmd/server/main.go) | Lee el entorno, cablea store, SMS y sesión, escucha `PORT` (default `8080`) | — | implementado |
| API HTTP | [`internal/adapters/httpapi/handler.go`](internal/adapters/httpapi/handler.go) | `GET /health` y los `POST /v1/*` | — | contrato interno estable ([`docs/openapi.yaml`](docs/openapi.yaml)) |
| Dominio `smsotp` | [`internal/domain/smsotp/`](internal/domain/smsotp/policy.go) | Challenge, política, E.164, propósito | — | implementado |
| Casos de uso | [`internal/application/smsotp/usecases.go`](internal/application/smsotp/usecases.go) | Enviar y verificar OTP (HMAC, fingerprint de teléfono) | — | implementado |
| Sesión MFA | [`internal/application/smsotp/session.go`](internal/application/smsotp/session.go) | Firma y valida `mfa_token` | el token no se persiste | implementado |
| Store Firestore | [`internal/adapters/store/firestore.go`](internal/adapters/store/firestore.go) | REST sobre la base `(default)` | colección `FIRESTORE_COLLECTION`, default `sms_otp_challenges` | implementado |
| Store DynamoDB | [`internal/adapters/store/dynamodb.go`](internal/adapters/store/dynamodb.go) | Store AWS firmado con SigV4 | tabla `DYNAMODB_TABLE`; default de scripts `sms-otp-challenges` | implementado |
| Store memoria | [`internal/adapters/store/memory.go`](internal/adapters/store/memory.go) | Store del proceso | ninguno | solo local y pruebas |
| Twilio | [`internal/adapters/sms/twilio/client.go`](internal/adapters/sms/twilio/client.go) | SMS por Twilio Messaging Service (HTTP, sin SDK) | — | implementado |
| Amazon SNS | [`internal/adapters/sms/sns/client.go`](internal/adapters/sms/sns/client.go) | SMS por Amazon SNS (SigV4) | — | implementado |

El challenge persiste `subject_id`, `phone_hash`, `purpose`, `hash`, `nonce`, `expires_at`, `sent_at`, `request_count` y `attempts` ([`internal/domain/smsotp/challenge.go`](internal/domain/smsotp/challenge.go)). No persiste el teléfono ni el OTP en claro.

La política por defecto, si no se inyecta otra, es TTL del OTP 5 min, ventana 10 min, máximo 3 envíos, 5 intentos, código de 6 dígitos y sesión MFA de 8 h ([`internal/domain/smsotp/policy.go`](internal/domain/smsotp/policy.go)). El propósito por defecto es `sms_mfa`.

## 4. Arquitectura y stack

| Capa | Tecnología y versión |
|---|---|
| Frontend | No aplica. Este repositorio no tiene UI. |
| Backend / API | Go `1.23.0` ([`go.mod`](go.mod)). HTTP de la biblioteca estándar. `go.mod` no declara dependencias de terceros. |
| Datos | Firestore por REST, DynamoDB o memoria. No hay SDK de Firebase. |
| Hosting / runtime | Imagen [`Dockerfile`](Dockerfile) (`gcr.io/distroless/static:nonroot`) y [`Dockerfile.lambda`](Dockerfile.lambda) (Lambda Web Adapter). Despliegue por script a Cloud Run, App Runner o Lambda. No hay `firebase.json`, Wrangler ni `package.json`. |
| Integraciones externas | Twilio Messaging Service y Amazon SNS. Secretos en Secret Manager (GCP), Secrets Manager (App Runner) o variables cifradas de Lambda. |

Descripción de arquitectura ISO/IEC/IEEE 42010 con vistas C4: [`docs/architecture.md`](docs/architecture.md). Espejo: [`docs/es/architecture.md`](docs/es/architecture.md). No existe `docs/arquitectura.md`.

## 5. Ambientes

El repositorio no fija project id, cuenta, URL ni rama de despliegue. Los scripts exigen variables del operador y etiquetan la imagen con `git rev-parse --short HEAD` o, si no hay git, con un timestamp.

| Ambiente | URL | Proyecto cloud | Rama o tag que despliega | Notas |
|---|---|---|---|---|
| Desarrollo | `http://localhost:8080` (default de `PORT`) | No aplica. `STORE_DRIVER=memory` no usa cloud. | No aplica. Se ejecuta en local con `go run ./cmd/server`. | [`.env.example`](.env.example) parte de `STORE_DRIVER=memory`. |
| Staging / prueba | PENDIENTE | PENDIENTE | PENDIENTE | El repo no define un ambiente de staging. |
| Producción | PENDIENTE | PENDIENTE | PENDIENTE. No hay tag git ni workflow. | La URL la imprimen los scripts al terminar y no queda en el repo. [`docs/openapi.yaml`](docs/openapi.yaml) usa `https://sms-otp-factor-service.example.com` como servidor de ejemplo. |

Defaults de región que sí están en los scripts, no son un ambiente desplegado:

- GCP: `REGION=us-central1`, Firestore `FIRESTORE_LOCATION=nam5`, servicio `sms-otp-factor-service` ([`scripts/install-gcp.sh`](scripts/install-gcp.sh)).
- AWS: `AWS_REGION` es obligatorio y no tiene default. La documentación de ejemplo usa `us-east-1` ([`docs/iaas-installation.md`](docs/iaas-installation.md)). La cuenta sale de `aws sts get-caller-identity` dentro del script.

## 6. Desarrollo local

[`docs/installation.md`](docs/installation.md) pide Go 1.22 o más nuevo. [`go.mod`](go.mod) declara `go 1.23.0`.

```bash
git clone https://github.com/inceptionlabscorp/sms-otp-factor-service.git
cd sms-otp-factor-service
cp .env.example .env
go test ./...
go run ./cmd/server
curl http://localhost:8080/health
```

Nombres de variables, sin valores: [`docs/configuration.md`](docs/configuration.md) y [`.env.example`](.env.example). Obligatorias al arrancar: `SMS_OTP_SERVICE_API_TOKEN`, `SMS_OTP_SECRET`, `SMS_PHONE_HASH_SECRET`, `SMS_MFA_SESSION_SECRET` (mínimo 32 caracteres). Según store y proveedor también `GCP_PROJECT_ID` o `GOOGLE_CLOUD_PROJECT`, `AWS_REGION`, `DYNAMODB_TABLE`, `TWILIO_ACCOUNT_SID`, `TWILIO_API_KEY_SID`, `TWILIO_API_KEY_SECRET`, `TWILIO_MESSAGING_SERVICE_SID`.

`STORE_DRIVER` en [`cmd/server/main.go`](cmd/server/main.go) admite `firestore` (default del proceso), `dynamodb` o `memory`. `SMS_PROVIDER` admite `twilio` (default) o `amazon_sns`.

Docker local: ver [`docs/installation.md`](docs/installation.md) (`docker build` / `docker run` con `--env-file .env`).

## 7. Despliegue y rollback

No hay workflow de GitHub que despliegue. Quién aprueba un despliegue: **PENDIENTE**.

| Destino | Script | Qué hace |
|---|---|---|
| Cloud Run, infra nueva | [`scripts/install-gcp.sh`](scripts/install-gcp.sh) | Pruebas, imagen en Artifact Registry (`sms-otp`), Firestore `(default)`, Secret Manager, cuenta de servicio, IAM y Cloud Run. Default `INGRESS=internal-and-cloud-load-balancing` y `ALLOW_UNAUTHENTICATED=false`. `STORE_DRIVER=firestore`. |
| Cloud Run, infra ya existente | [`scripts/deploy-prod.sh`](scripts/deploy-prod.sh) | Pruebas, imagen en `gcr.io/<GCP_PROJECT_ID>/sms-otp-factor-service:<sha>-local` y `gcloud run deploy` con `--allow-unauthenticated`. Monta un secreto de Secret Manager solo si ya existe. Exige `GCP_PROJECT_ID`. |
| App Runner | [`scripts/install-aws.sh`](scripts/install-aws.sh) | Preflight de App Runner, ECR, DynamoDB on-demand con TTL en `expires_at`, Secrets Manager, roles e instancia. `STORE_DRIVER=dynamodb`. Default de proveedor `amazon_sns`. |
| Lambda (free tier documentado) | [`scripts/install-aws-lambda-free-tier.sh`](scripts/install-aws-lambda-free-tier.sh) | ECR, función con imagen, Function URL (`AuthType` `NONE` a nivel de URL), DynamoDB e IAM. Los secretos van como variables cifradas de Lambda. `/v1/*` sigue exigiendo el bearer. |

Detalle de recursos: [`docs/iaas-installation.md`](docs/iaas-installation.md) y [`docs/es/iaas-installation.md`](docs/es/iaas-installation.md). Los cuatro scripts ejecutan `go test ./...` antes de construir.

Rollback de producto, en [`docs/architecture.md`](docs/architecture.md) sección 15: desactivar la exigencia de SMS OTP en el backend de confianza, usar otro segundo factor aprobado si existe, y dejar el servicio fuera del camino crítico. El espejo [`docs/es/architecture.md`](docs/es/architecture.md) no trae esa sección. Rollback de una revisión de Cloud Run, App Runner o Lambda: **PENDIENTE** (no está en los scripts ni en los docs).

Secretos, solo nombres y dónde viven:

| Variable | Nombre en GCP Secret Manager | Nombre en AWS Secrets Manager (App Runner) |
|---|---|---|
| `SMS_OTP_SERVICE_API_TOKEN` | `sms-otp-service-api-token` | `<SERVICE_NAME>/sms-otp-service-api-token` |
| `SMS_OTP_SECRET` | `sms-otp-secret` | `<SERVICE_NAME>/sms-otp-secret` |
| `SMS_PHONE_HASH_SECRET` | `sms-phone-hash-secret` | `<SERVICE_NAME>/sms-phone-hash-secret` |
| `SMS_MFA_SESSION_SECRET` | `sms-mfa-session-secret` | `<SERVICE_NAME>/sms-mfa-session-secret` |
| `TWILIO_ACCOUNT_SID` | `twilio-account-sid` | `<SERVICE_NAME>/twilio-account-sid` |
| `TWILIO_API_KEY_SID` | `twilio-api-key-sid` | `<SERVICE_NAME>/twilio-api-key-sid` |
| `TWILIO_API_KEY_SECRET` | `twilio-api-key-secret` | `<SERVICE_NAME>/twilio-api-key-secret` |
| `TWILIO_MESSAGING_SERVICE_SID` | `twilio-messaging-service-sid` | `<SERVICE_NAME>/twilio-messaging-service-sid` |
| `AWS_ACCESS_KEY_ID` | `aws-sns-access-key-id` | No aplica en App Runner: el rol de instancia publica en SNS. |
| `AWS_SECRET_ACCESS_KEY` | `aws-sns-secret-access-key` | No aplica en App Runner. |
| `AWS_SESSION_TOKEN` | `aws-sns-session-token` | No aplica en App Runner. |

`<SERVICE_NAME>` default `sms-otp-factor-service` ([`scripts/install-aws.sh`](scripts/install-aws.sh)). En Lambda los mismos nombres de variable quedan en el entorno de la función ([`scripts/install-aws-lambda-free-tier.sh`](scripts/install-aws-lambda-free-tier.sh)). No hay GitHub Environments ni Secret Manager de otro producto en este repo.

## 8. API y endpoints

Contrato: [`docs/openapi.yaml`](docs/openapi.yaml) y [`docs/api-integration.md`](docs/api-integration.md). Catálogo de errores: [`docs/error-catalog.md`](docs/error-catalog.md).

| Método | Ruta | Uso | Auth |
|---|---|---|---|
| `GET` | `/health` | Salud. Cuerpo con `status` y `service`. | ninguna |
| `POST` | `/v1/sms-otp/send` | Crea el challenge y envía el SMS. Responde `202` con `status: sent`. | `Authorization: Bearer <SMS_OTP_SERVICE_API_TOKEN>` |
| `POST` | `/v1/sms-otp/verify` | Verifica el código y devuelve `mfa_token`, `expires_in` y `method`. | mismo bearer |
| `POST` | `/v1/sms-mfa/session/validate` | Valida el `mfa_token` para un `subject_id`. Devuelve `valid`. | mismo bearer |
| `POST` | `/v1/admin/sms-otp/send` | Alias legado de send. | mismo bearer |
| `POST` | `/v1/admin/sms-otp/verify` | Alias legado de verify. | mismo bearer |
| `POST` | `/v1/admin/sms-mfa/session/validate` | Alias legado de validate. | mismo bearer |
| `POST` | `/v1/bian/customer-access-entitlement/sms-otp/initiate` | Alias BIAN-aligned de send. | mismo bearer |
| `POST` | `/v1/bian/customer-access-entitlement/sms-otp/execute` | Alias BIAN-aligned de verify. | mismo bearer |
| `POST` | `/v1/bian/customer-access-entitlement/sms-mfa-session/evaluate` | Alias BIAN-aligned de validate. | mismo bearer |

Cuerpo máximo 4096 bytes. El JSON rechaza campos desconocidos ([`internal/adapters/httpapi/handler.go`](internal/adapters/httpapi/handler.go)).

## 9. Controles de CI

No existe el directorio `.github/workflows`. No hay otro workflow de CI en el repositorio.

| Workflow | Disparador | Qué verifica | Bloquea el merge |
|---|---|---|---|
| No aplica: no hay archivos en `.github/workflows` | — | Nada en CI. `go test ./...` corre en local y dentro de los scripts de despliegue. | No hay un workflow que bloquee el merge. |

Tampoco hay un control que exija actualizar «Última revisión del README».

## 10. Herramientas MCP

- Este proyecto **no expone MCP**. No hay servidor, URL, transporte, scopes ni `mcpTools.ts`.
- Este proyecto **no consume MCP**. No hay cliente ni configuración de servidores MCP.

| Herramienta | Lectura/escritura | Scope | Qué hace |
|---|---|---|---|
| No aplica | — | — | El repositorio no define herramientas MCP. |

## 11. Observabilidad y operación

- Salud: `GET /health` devuelve `status: ok` y `service: sms-otp-factor-service`. No incluye datos de desafío.
- El proceso solo registra en log que está escuchando ([`cmd/server/main.go`](cmd/server/main.go)). El perfil NIST prohíbe loguear OTP, token, teléfono completo o secretos, y deja la auditoría de decisiones en el backend de confianza ([`docs/nist-security-profile.md`](docs/nist-security-profile.md)).
- El instalador Lambda concede al rol de ejecución permiso de CloudWatch Logs, además de DynamoDB y `sns:Publish` ([`docs/iaas-installation.md`](docs/iaas-installation.md)). Eso es logging de plataforma, no un monitor de producto.
- Uptime, alertas, on-call y SIEM: **PENDIENTE**. No existe `docs/operacion.md`.

## 12. Seguridad y datos

Trata un identificador opaco `subject_id` que envía el backend y el teléfono solo para calcular `phone_hash` y para entregarlo al proveedor SMS. El challenge guarda el HMAC del OTP y el HMAC del teléfono, con secretos distintos. No es la fuente de verdad de usuarios, roles ni teléfonos autorizados ([`SECURITY.md`](SECURITY.md)).

- Retención: **PENDIENTE** como política. El perfil NIST la exige al operador y este repo no la documenta con plazos. Los instaladores AWS crean TTL de DynamoDB sobre `expires_at`. Verificar un código válido o vencido borra el challenge (CR-03 en [`docs/architecture.md`](docs/architecture.md)).
- RBAC de usuarios: no aplica dentro del servicio. Hay un solo token de servicio a servicio, sin roles. El control de acceso de personas vive en el backend que llama.
- Auditoría: no hay colección ni log de auditoría en este repo. El perfil NIST asigna ese control al backend de confianza.
- Aviso de vulnerabilidad: [`SECURITY.md`](SECURITY.md). La rama `main` es la que recibe correcciones. El canal pedido es GitHub Security Advisories o el canal de seguridad de la organización, sin abrir un issue público.
- SMS OTP no es resistente a phishing. El perfil NIST lo trata como factor fuera de banda restringido.

## 13. Documentación relacionada

| Documento | Contenido |
|---|---|
| `AGENTS.md` | No aplica. El archivo no está en el repositorio. |
| [`CHANGELOG.md`](CHANGELOG.md) | Cambios por versión. Última entrada: 0.5.6, 2026-07-17. |
| [`SECURITY.md`](SECURITY.md) | Versiones soportadas y reporte de vulnerabilidades. |
| [`LICENSE`](LICENSE) | Apache-2.0. |
| [`docs/architecture.md`](docs/architecture.md) | ISO/IEC/IEEE 42010 y C4, en inglés, con rollback de producto. |
| [`docs/es/architecture.md`](docs/es/architecture.md) | Espejo en español. No incluye la sección Rollback. |
| [`docs/installation.md`](docs/installation.md) | Instalación local, Docker y punteros de despliegue. |
| [`docs/configuration.md`](docs/configuration.md) | Variables de entorno y nombres de secretos. |
| [`docs/iaas-installation.md`](docs/iaas-installation.md) | Recursos que crean los scripts GCP y AWS. |
| [`docs/api-integration.md`](docs/api-integration.md) | Responsabilidades del backend y ejemplos de API. |
| [`docs/error-catalog.md`](docs/error-catalog.md) | Códigos de error estables, agnósticos del proveedor. |
| [`docs/nist-security-profile.md`](docs/nist-security-profile.md) | Mapa a controles NIST. No es certificación ni auditoría. |
| [`docs/openapi.yaml`](docs/openapi.yaml) | OpenAPI 3.0.3, versión de contrato 0.5.6. |
| [`docs/es/`](docs/es/installation.md) | Espejos en español de la documentación pública. |

## 14. Pendientes conocidos

- **PENDIENTE.** Estado de ciclo de vida (activo, mantenimiento, legado o archivado). [`docs/architecture.md`](docs/architecture.md) solo dice «Public reference implementation». [`SECURITY.md`](SECURITY.md) dice que `main` recibe correcciones. La versión publicada en el changelog es 0.5.6 (2026-07-17).
- **PENDIENTE.** Cliente o razón social. El changelog 0.3.0 dice que se quitaron nombres y dominios internos del contrato público. Este árbol no nombra un cliente.
- **PENDIENTE.** Owner técnico. No hay `CODEOWNERS` ni dueño en el repo.
- **PENDIENTE.** URL de producción y URL de staging. No hay un ambiente de staging. La URL de Cloud Run, App Runner o Lambda no está commiteada.
- **PENDIENTE.** Proyecto cloud por ambiente: ni project id de GCP ni cuenta o zona de AWS. `GCP_PROJECT_ID` es obligatorio y no tiene default.
- **PENDIENTE.** Repos relacionados y consumidores. Este repo no nombra quién lo llama.
- **PENDIENTE.** Quién aprueba un despliegue. No hay GitHub Environment ni regla de aprobación.
- **PENDIENTE.** Rollback de la revisión de Cloud Run, App Runner o Lambda. Solo está el rollback de producto en la sección 15 de [`docs/architecture.md`](docs/architecture.md).
- **PENDIENTE.** Monitoreo, alertas, on-call y runbook. No existe `docs/operacion.md`.
- **PENDIENTE.** Política de retención (plazos) de challenges y de logs de entrega.
- **PENDIENTE.** CI. No existe `.github/workflows`, así que nada bloquea el merge ni comprueba este README.
- Desfase de documentación, verificable en el repo: [`docs/installation.md`](docs/installation.md) dice Go 1.22 o más nuevo y [`go.mod`](go.mod) declara `go 1.23.0`. [`docs/api-integration.md`](docs/api-integration.md) muestra `"expires_in": 900`; con la política por defecto de [`internal/domain/smsotp/policy.go`](internal/domain/smsotp/policy.go) la sesión dura 8 h (`28800` segundos) porque [`cmd/server/main.go`](cmd/server/main.go) no inyecta otra `Policy`. [`docs/configuration.md`](docs/configuration.md) describe `STORE_DRIVER` como `firestore` o `memory`; el proceso también acepta `dynamodb`.

---

# English

[Español](#sms-otp-factor-service)

Same index, same sources. Anything not present in this repository is **PENDIENTE**.

## 1. Record

| Field | Value |
|---|---|
| Repository | [`inceptionlabscorp/sms-otp-factor-service`](https://github.com/inceptionlabscorp/sms-otp-factor-service) |
| Type | API (private HTTP microservice) |
| Status | PENDIENTE |
| Client | PENDIENTE |
| Product or system | SMS OTP Factor Service, bounded context `smsotp` |
| Technical owner | PENDIENTE |
| Production URL | PENDIENTE |
| Staging or test URL | PENDIENTE |
| Cloud project | PENDIENTE |
| Related repos | PENDIENTE |
| Current version | 0.5.6 ([`CHANGELOG.md`](CHANGELOG.md), [`docs/openapi.yaml`](docs/openapi.yaml)). No git tags. |
| README last reviewed | 2026-10-21 |

## 2. Purpose and scope

The service owns SMS OTP challenge creation, SMS delivery, OTP verification, and short-lived `mfa_token` signing and validation. The trusted backend authenticates the subject and authorizes the phone number before calling. See [`docs/architecture.md`](docs/architecture.md) and [`docs/api-integration.md`](docs/api-integration.md).

Boundaries:

- No UI, no identity provider, and no phone-ownership policy.
- It does not verify a person's identity and it does not decide that a phone number belongs to a user ([`docs/nist-security-profile.md`](docs/nist-security-profile.md), Non-Goals).
- Browsers and mobile apps must not call it. `/v1/*` requires the service bearer token.
- `/v1/bian/*` aliases use BIAN-style Customer Access Entitlement language. They are not a certified BIAN API ([`docs/openapi.yaml`](docs/openapi.yaml)).
- It neither exposes nor consumes MCP. See section 10.

## 3. Modules and behavior

| Module | Path or entrypoint | What it does | Data (tables or collections) | Status |
|---|---|---|---|---|
| Bootstrap | [`cmd/server/main.go`](cmd/server/main.go) | Reads env, wires store, SMS, and session, listens on `PORT` (default `8080`) | — | implemented |
| HTTP API | [`internal/adapters/httpapi/handler.go`](internal/adapters/httpapi/handler.go) | `GET /health` and `POST /v1/*` | — | stable internal contract ([`docs/openapi.yaml`](docs/openapi.yaml)) |
| `smsotp` domain | [`internal/domain/smsotp/`](internal/domain/smsotp/policy.go) | Challenge, policy, E.164, purpose | — | implemented |
| Use cases | [`internal/application/smsotp/usecases.go`](internal/application/smsotp/usecases.go) | Send and verify OTP (HMAC, phone fingerprint) | — | implemented |
| MFA session | [`internal/application/smsotp/session.go`](internal/application/smsotp/session.go) | Signs and validates `mfa_token` | token is not persisted | implemented |
| Firestore store | [`internal/adapters/store/firestore.go`](internal/adapters/store/firestore.go) | REST against the `(default)` database | collection `FIRESTORE_COLLECTION`, default `sms_otp_challenges` | implemented |
| DynamoDB store | [`internal/adapters/store/dynamodb.go`](internal/adapters/store/dynamodb.go) | AWS store signed with SigV4 | table `DYNAMODB_TABLE`; script default `sms-otp-challenges` | implemented |
| Memory store | [`internal/adapters/store/memory.go`](internal/adapters/store/memory.go) | In-process store | none | local and tests only |
| Twilio | [`internal/adapters/sms/twilio/client.go`](internal/adapters/sms/twilio/client.go) | SMS via Twilio Messaging Service (HTTP, no SDK) | — | implemented |
| Amazon SNS | [`internal/adapters/sms/sns/client.go`](internal/adapters/sms/sns/client.go) | SMS via Amazon SNS (SigV4) | — | implemented |

A challenge stores `subject_id`, `phone_hash`, `purpose`, `hash`, `nonce`, `expires_at`, `sent_at`, `request_count`, and `attempts` ([`internal/domain/smsotp/challenge.go`](internal/domain/smsotp/challenge.go)). It does not store the raw phone number or the raw OTP.

Default policy, when none is injected: OTP TTL 5 min, window 10 min, max 3 sends, 5 attempts, 6-digit code, MFA session 8 h ([`internal/domain/smsotp/policy.go`](internal/domain/smsotp/policy.go)). Default purpose is `sms_mfa`.

## 4. Architecture and stack

| Layer | Technology and version |
|---|---|
| Frontend | Not applicable. This repository has no UI. |
| Backend / API | Go `1.23.0` ([`go.mod`](go.mod)). Standard-library HTTP. `go.mod` declares no third-party dependencies. |
| Data | Firestore over REST, DynamoDB, or memory. No Firebase SDK. |
| Hosting / runtime | [`Dockerfile`](Dockerfile) (`gcr.io/distroless/static:nonroot`) and [`Dockerfile.lambda`](Dockerfile.lambda) (Lambda Web Adapter). Scripts deploy to Cloud Run, App Runner, or Lambda. No `firebase.json`, Wrangler, or `package.json`. |
| External integrations | Twilio Messaging Service and Amazon SNS. Secrets in Secret Manager (GCP), Secrets Manager (App Runner), or Lambda encrypted environment variables. |

ISO/IEC/IEEE 42010 architecture description with C4 views: [`docs/architecture.md`](docs/architecture.md). Spanish mirror: [`docs/es/architecture.md`](docs/es/architecture.md). There is no `docs/arquitectura.md`.

## 5. Environments

This repository does not pin a project id, account, URL, or deploy branch. Scripts require operator variables and tag the image with `git rev-parse --short HEAD`, or a timestamp when git is unavailable.

| Environment | URL | Cloud project | Branch or tag that deploys | Notes |
|---|---|---|---|---|
| Development | `http://localhost:8080` (`PORT` default) | Not applicable. `STORE_DRIVER=memory` uses no cloud. | Not applicable. Run locally with `go run ./cmd/server`. | [`.env.example`](.env.example) starts from `STORE_DRIVER=memory`. |
| Staging / test | PENDIENTE | PENDIENTE | PENDIENTE | No staging environment is defined. |
| Production | PENDIENTE | PENDIENTE | PENDIENTE. No git tag and no workflow. | Scripts print the URL and do not commit it. [`docs/openapi.yaml`](docs/openapi.yaml) uses `https://sms-otp-factor-service.example.com` as an example server. |

Region defaults that exist in scripts are not a deployed environment:

- GCP: `REGION=us-central1`, Firestore `FIRESTORE_LOCATION=nam5`, service `sms-otp-factor-service` ([`scripts/install-gcp.sh`](scripts/install-gcp.sh)).
- AWS: `AWS_REGION` is required and has no default. The docs use `us-east-1` as an example ([`docs/iaas-installation.md`](docs/iaas-installation.md)). The account comes from `aws sts get-caller-identity` inside the script.

## 6. Local development

[`docs/installation.md`](docs/installation.md) asks for Go 1.22 or newer. [`go.mod`](go.mod) declares `go 1.23.0`.

```bash
git clone https://github.com/inceptionlabscorp/sms-otp-factor-service.git
cd sms-otp-factor-service
cp .env.example .env
go test ./...
go run ./cmd/server
curl http://localhost:8080/health
```

Variable names only: [`docs/configuration.md`](docs/configuration.md) and [`.env.example`](.env.example). Required at startup: `SMS_OTP_SERVICE_API_TOKEN`, `SMS_OTP_SECRET`, `SMS_PHONE_HASH_SECRET`, `SMS_MFA_SESSION_SECRET` (at least 32 characters). Depending on store and provider, also `GCP_PROJECT_ID` or `GOOGLE_CLOUD_PROJECT`, `AWS_REGION`, `DYNAMODB_TABLE`, `TWILIO_ACCOUNT_SID`, `TWILIO_API_KEY_SID`, `TWILIO_API_KEY_SECRET`, `TWILIO_MESSAGING_SERVICE_SID`.

`STORE_DRIVER` in [`cmd/server/main.go`](cmd/server/main.go) accepts `firestore` (process default), `dynamodb`, or `memory`. `SMS_PROVIDER` accepts `twilio` (default) or `amazon_sns`.

Local Docker: [`docs/installation.md`](docs/installation.md) (`docker build` / `docker run` with `--env-file .env`).

## 7. Deploy and rollback

No GitHub workflow deploys this service. Who approves a deploy: **PENDIENTE**.

| Target | Script | What it does |
|---|---|---|
| Cloud Run, new infra | [`scripts/install-gcp.sh`](scripts/install-gcp.sh) | Tests, Artifact Registry image (`sms-otp`), Firestore `(default)`, Secret Manager, service account, IAM, and Cloud Run. Defaults `INGRESS=internal-and-cloud-load-balancing` and `ALLOW_UNAUTHENTICATED=false`. `STORE_DRIVER=firestore`. |
| Cloud Run, existing infra | [`scripts/deploy-prod.sh`](scripts/deploy-prod.sh) | Tests, image `gcr.io/<GCP_PROJECT_ID>/sms-otp-factor-service:<sha>-local`, and `gcloud run deploy` with `--allow-unauthenticated`. Mounts a Secret Manager secret only when it already exists. Requires `GCP_PROJECT_ID`. |
| App Runner | [`scripts/install-aws.sh`](scripts/install-aws.sh) | App Runner preflight, ECR, on-demand DynamoDB with TTL on `expires_at`, Secrets Manager, roles, and the service. `STORE_DRIVER=dynamodb`. Default provider `amazon_sns`. |
| Lambda (documented free tier) | [`scripts/install-aws-lambda-free-tier.sh`](scripts/install-aws-lambda-free-tier.sh) | ECR, container-image function, Function URL (URL `AuthType` `NONE`), DynamoDB, and IAM. Secrets are Lambda encrypted environment variables. `/v1/*` still requires the bearer token. |

Resource list: [`docs/iaas-installation.md`](docs/iaas-installation.md) and [`docs/es/iaas-installation.md`](docs/es/iaas-installation.md). All four scripts run `go test ./...` before the build.

Product rollback, [`docs/architecture.md`](docs/architecture.md) section 15: turn off SMS OTP enforcement in the trusted backend, fall back to another approved second factor when one exists, and keep this service off the critical path. The Spanish mirror [`docs/es/architecture.md`](docs/es/architecture.md) does not include that section. Rollback of a Cloud Run, App Runner, or Lambda revision: **PENDIENTE** (not in the scripts or the docs).

Secrets, names and where they live:

| Variable | GCP Secret Manager name | AWS Secrets Manager name (App Runner) |
|---|---|---|
| `SMS_OTP_SERVICE_API_TOKEN` | `sms-otp-service-api-token` | `<SERVICE_NAME>/sms-otp-service-api-token` |
| `SMS_OTP_SECRET` | `sms-otp-secret` | `<SERVICE_NAME>/sms-otp-secret` |
| `SMS_PHONE_HASH_SECRET` | `sms-phone-hash-secret` | `<SERVICE_NAME>/sms-phone-hash-secret` |
| `SMS_MFA_SESSION_SECRET` | `sms-mfa-session-secret` | `<SERVICE_NAME>/sms-mfa-session-secret` |
| `TWILIO_ACCOUNT_SID` | `twilio-account-sid` | `<SERVICE_NAME>/twilio-account-sid` |
| `TWILIO_API_KEY_SID` | `twilio-api-key-sid` | `<SERVICE_NAME>/twilio-api-key-sid` |
| `TWILIO_API_KEY_SECRET` | `twilio-api-key-secret` | `<SERVICE_NAME>/twilio-api-key-secret` |
| `TWILIO_MESSAGING_SERVICE_SID` | `twilio-messaging-service-sid` | `<SERVICE_NAME>/twilio-messaging-service-sid` |
| `AWS_ACCESS_KEY_ID` | `aws-sns-access-key-id` | Not used on App Runner: the instance role publishes to SNS. |
| `AWS_SECRET_ACCESS_KEY` | `aws-sns-secret-access-key` | Not used on App Runner. |
| `AWS_SESSION_TOKEN` | `aws-sns-session-token` | Not used on App Runner. |

`<SERVICE_NAME>` defaults to `sms-otp-factor-service` ([`scripts/install-aws.sh`](scripts/install-aws.sh)). On Lambda the same variable names are function environment variables ([`scripts/install-aws-lambda-free-tier.sh`](scripts/install-aws-lambda-free-tier.sh)). This repo has no GitHub Environments.

## 8. API and endpoints

Contract: [`docs/openapi.yaml`](docs/openapi.yaml) and [`docs/api-integration.md`](docs/api-integration.md). Error catalog: [`docs/error-catalog.md`](docs/error-catalog.md).

| Method | Path | Use | Auth |
|---|---|---|---|
| `GET` | `/health` | Health. Body includes `status` and `service`. | none |
| `POST` | `/v1/sms-otp/send` | Creates the challenge and sends the SMS. `202` with `status: sent`. | `Authorization: Bearer <SMS_OTP_SERVICE_API_TOKEN>` |
| `POST` | `/v1/sms-otp/verify` | Verifies the code and returns `mfa_token`, `expires_in`, and `method`. | same bearer |
| `POST` | `/v1/sms-mfa/session/validate` | Validates `mfa_token` for a `subject_id`. Returns `valid`. | same bearer |
| `POST` | `/v1/admin/sms-otp/send` | Legacy alias of send. | same bearer |
| `POST` | `/v1/admin/sms-otp/verify` | Legacy alias of verify. | same bearer |
| `POST` | `/v1/admin/sms-mfa/session/validate` | Legacy alias of validate. | same bearer |
| `POST` | `/v1/bian/customer-access-entitlement/sms-otp/initiate` | BIAN-aligned alias of send. | same bearer |
| `POST` | `/v1/bian/customer-access-entitlement/sms-otp/execute` | BIAN-aligned alias of verify. | same bearer |
| `POST` | `/v1/bian/customer-access-entitlement/sms-mfa-session/evaluate` | BIAN-aligned alias of validate. | same bearer |

Maximum body 4096 bytes. The JSON decoder rejects unknown fields ([`internal/adapters/httpapi/handler.go`](internal/adapters/httpapi/handler.go)).

## 9. CI controls

The directory `.github/workflows` does not exist. There is no other CI workflow in the repository.

| Workflow | Trigger | What it checks | Blocks merge |
|---|---|---|---|
| Not applicable: no files under `.github/workflows` | — | Nothing in CI. `go test ./...` runs locally and inside the deploy scripts. | No workflow blocks merge. |

Nothing in CI checks «README last reviewed».

## 10. MCP tools

- This project **does not expose MCP**. There is no server, URL, transport, scope, or `mcpTools.ts`.
- This project **does not consume MCP**. There is no MCP client and no MCP server configuration.

| Tool | Read/write | Scope | What it does |
|---|---|---|---|
| Not applicable | — | — | The repository defines no MCP tools. |

## 11. Observability and operations

- Health: `GET /health` returns `status: ok` and `service: sms-otp-factor-service`. It does not include challenge data.
- The process only logs that it is listening ([`cmd/server/main.go`](cmd/server/main.go)). The NIST profile forbids logging OTPs, tokens, full phone numbers, or secrets, and places decision audit on the trusted backend ([`docs/nist-security-profile.md`](docs/nist-security-profile.md)).
- The Lambda installer grants the execution role CloudWatch Logs, plus DynamoDB and `sns:Publish` ([`docs/iaas-installation.md`](docs/iaas-installation.md)). That is platform logging, not a product monitor.
- Uptime, alerts, on-call, and SIEM: **PENDIENTE**. `docs/operacion.md` does not exist.

## 12. Security and data

The service handles an opaque `subject_id` supplied by the backend, and the phone number only to compute `phone_hash` and to hand it to the SMS provider. The challenge stores an OTP HMAC and a phone HMAC, with separate secrets. It is not the source of truth for users, roles, or authorized phone numbers ([`SECURITY.md`](SECURITY.md)).

- Retention: **PENDIENTE** as a policy. The NIST profile requires the operator to document one, and this repo does not set retention periods. AWS installers create a DynamoDB TTL on `expires_at`. Verification of a valid or expired code deletes the challenge (CR-03 in [`docs/architecture.md`](docs/architecture.md)).
- User RBAC: not applicable inside the service. There is one service-to-service token and no roles. Person-level access control lives in the calling backend.
- Audit: this repo has no audit collection and no audit log. The NIST profile assigns that control to the trusted backend.
- Vulnerability reporting: [`SECURITY.md`](SECURITY.md). The `main` branch receives security fixes. Report through GitHub Security Advisories or the organization security channel, not a public issue.
- SMS OTP is not phishing-resistant. The NIST profile treats it as a restricted out-of-band factor.

## 13. Related documentation

| Document | Contents |
|---|---|
| `AGENTS.md` | Not applicable. The file is not in the repository. |
| [`CHANGELOG.md`](CHANGELOG.md) | Changes by version. Latest entry: 0.5.6, 2026-07-17. |
| [`SECURITY.md`](SECURITY.md) | Supported versions and vulnerability reporting. |
| [`LICENSE`](LICENSE) | Apache-2.0. |
| [`docs/architecture.md`](docs/architecture.md) | ISO/IEC/IEEE 42010 and C4, including product rollback. |
| [`docs/es/architecture.md`](docs/es/architecture.md) | Spanish mirror. It does not include the Rollback section. |
| [`docs/installation.md`](docs/installation.md) | Local install, Docker, and deploy pointers. |
| [`docs/configuration.md`](docs/configuration.md) | Environment variables and secret names. |
| [`docs/iaas-installation.md`](docs/iaas-installation.md) | Resources created by the GCP and AWS scripts. |
| [`docs/api-integration.md`](docs/api-integration.md) | Backend responsibilities and API examples. |
| [`docs/error-catalog.md`](docs/error-catalog.md) | Stable, provider-agnostic error codes. |
| [`docs/nist-security-profile.md`](docs/nist-security-profile.md) | NIST control mapping. Not a certification or an audit. |
| [`docs/openapi.yaml`](docs/openapi.yaml) | OpenAPI 3.0.3, contract version 0.5.6. |
| [`docs/es/`](docs/es/installation.md) | Spanish mirrors of the public docs. |

## 14. Known pending items

- **PENDIENTE.** Lifecycle status (active, maintenance, legacy, or archived). [`docs/architecture.md`](docs/architecture.md) only says "Public reference implementation". [`SECURITY.md`](SECURITY.md) says `main` receives security fixes. The changelog version is 0.5.6 (2026-07-17).
- **PENDIENTE.** Client or legal name. Changelog 0.3.0 says internal names and domains were removed from the public contract. This tree names no client.
- **PENDIENTE.** Technical owner. There is no `CODEOWNERS` and no owner in the repo.
- **PENDIENTE.** Production URL and staging URL. No staging environment is defined. The Cloud Run, App Runner, or Lambda URL is not committed.
- **PENDIENTE.** Cloud project per environment: no GCP project id and no AWS account or zone. `GCP_PROJECT_ID` is required and has no default.
- **PENDIENTE.** Related repositories and consumers. This repo does not name its callers.
- **PENDIENTE.** Who approves a deploy. There is no GitHub Environment and no approval rule.
- **PENDIENTE.** Rollback of a Cloud Run, App Runner, or Lambda revision. Only the product rollback in section 15 of [`docs/architecture.md`](docs/architecture.md) is documented.
- **PENDIENTE.** Monitoring, alerts, on-call, and a runbook. `docs/operacion.md` does not exist.
- **PENDIENTE.** Retention periods for challenges and delivery logs.
- **PENDIENTE.** CI. `.github/workflows` does not exist, so nothing blocks merge and nothing checks this README.
- Documentation drift verified in the repo: [`docs/installation.md`](docs/installation.md) says Go 1.22 or newer, while [`go.mod`](go.mod) declares `go 1.23.0`. [`docs/api-integration.md`](docs/api-integration.md) shows `"expires_in": 900`; the default policy in [`internal/domain/smsotp/policy.go`](internal/domain/smsotp/policy.go) is an 8 h session (`28800` seconds) because [`cmd/server/main.go`](cmd/server/main.go) does not inject another `Policy`. [`docs/configuration.md`](docs/configuration.md) describes `STORE_DRIVER` as `firestore` or `memory`; the process also accepts `dynamodb`.
