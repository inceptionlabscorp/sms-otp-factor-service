# SMS OTP Factor Service

English | [Español](#servicio-de-factor-otp-sms)

Provider-agnostic SMS OTP microservice for second-factor verification flows.

The service owns OTP challenge creation, SMS delivery, OTP verification, and short-lived MFA session token validation. It is designed as a private backend service: your trusted backend decides who the subject is and which phone number is authorized, then calls this service over an authenticated service-to-service API.

## Features

- Tactical DDD bounded context: `smsotp`.
- Hexagonal architecture: domain/application inside, HTTP/store/SMS providers outside.
- OTP challenges stored as HMAC hashes with nonce, expiry, attempts and cooldown.
- Pluggable SMS providers: Twilio Messaging Service or Amazon Simple Notification Service.
- Configurable SMS message template.
- Stable internal API plus BIAN-aligned aliases for banking-style integration language.
- No UI, no identity provider coupling, no phone-number ownership policy.

## Documentation

- [Installation](docs/installation.md) | [Instalacion](docs/es/installation.md)
- [Configuration](docs/configuration.md) | [Configuracion](docs/es/configuration.md)
- [IaaS installation scripts](docs/iaas-installation.md) | [Scripts IaaS](docs/es/iaas-installation.md)
- [API integration](docs/api-integration.md) | [Integracion API](docs/es/api-integration.md)
- [Provider-agnostic error catalog](docs/error-catalog.md) | [Catalogo de errores](docs/es/error-catalog.md)
- [Architecture](docs/architecture.md) | [Arquitectura](docs/es/architecture.md)
- [NIST security profile](docs/nist-security-profile.md) | [Perfil de seguridad NIST](docs/es/nist-security-profile.md)
- [OpenAPI contract](docs/openapi.yaml)

## Quick Start

```bash
cp .env.example .env
export SMS_OTP_SERVICE_API_TOKEN="replace-with-a-long-random-token"
export SMS_OTP_SECRET="replace-with-at-least-32-random-bytes"
export SMS_PHONE_HASH_SECRET="replace-with-a-separate-random-secret"
export SMS_MFA_SESSION_SECRET="replace-with-at-least-32-random-bytes"
export STORE_DRIVER=memory
export SMS_PROVIDER=twilio
export TWILIO_ACCOUNT_SID="AC..."
export TWILIO_API_KEY_SID="SK..."
export TWILIO_API_KEY_SECRET="..."
export TWILIO_MESSAGING_SERVICE_SID="MG..."
go run ./cmd/server
```

Health check:

```bash
curl http://localhost:8080/health
```

Send OTP:

```bash
curl -X POST http://localhost:8080/v1/sms-otp/send \
  -H "Authorization: Bearer ${SMS_OTP_SERVICE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"subject_id":"user-123","phone_number":"+15555550100"}'
```

Verify OTP:

```bash
curl -X POST http://localhost:8080/v1/sms-otp/verify \
  -H "Authorization: Bearer ${SMS_OTP_SERVICE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"subject_id":"user-123","phone_number":"+15555550100","code":"123456"}'
```

The verify response returns an `mfa_token`. Your backend can validate it with:

```bash
curl -X POST http://localhost:8080/v1/sms-mfa/session/validate \
  -H "Authorization: Bearer ${SMS_OTP_SERVICE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"subject_id":"user-123","token":"<mfa_token>"}'
```

## Security Model

- Do not expose this service directly to browsers or mobile apps.
- Keep `SMS_OTP_SERVICE_API_TOKEN`, `SMS_OTP_SECRET`, provider credentials, and session secrets outside the repository.
- The caller must authorize the subject and phone number before calling `/send`.
- The service must not be the source of truth for users, roles, or authorized phone numbers.
- Challenge storage keeps a phone HMAC fingerprint, not the raw phone number.
- Logs must not include OTP codes, session tokens, full phone numbers, or secrets.
- SMS OTP is not phishing-resistant; regulated deployments should also provide a phishing-resistant factor for high-risk flows.

## Development

```bash
go test ./...
go build ./cmd/server
git diff --check
```

## License

Apache-2.0. See [LICENSE](LICENSE).

---

# Servicio de Factor OTP SMS

[English](#sms-otp-factor-service) | Español

Microservicio OTP SMS agnostico de proveedor para flujos de segundo factor.

El servicio es dueño de la creacion del challenge OTP, envio SMS, verificacion OTP y validacion de tokens MFA de corta duracion. Esta disenado como servicio privado de backend: tu backend confiable decide quien es el sujeto y que numero telefonico esta autorizado, y luego llama a este servicio por una API autenticada de servicio a servicio.

## Capacidades

- DDD tactico en el bounded context `smsotp`.
- Arquitectura hexagonal: dominio/aplicacion adentro; HTTP/store/proveedores SMS afuera.
- Challenges OTP guardados como HMAC con nonce, expiracion, intentos y cooldown.
- Proveedores SMS intercambiables: Twilio Messaging Service o Amazon Simple Notification Service.
- Plantilla configurable para el mensaje SMS.
- API interna estable y aliases BIAN-aligned para lenguaje bancario.
- Sin UI, sin acoplamiento a identity provider, sin politica propia de titularidad de telefono.

## Documentacion

- [Instalacion](docs/es/installation.md) | [Installation](docs/installation.md)
- [Configuracion](docs/es/configuration.md) | [Configuration](docs/configuration.md)
- [Scripts IaaS](docs/es/iaas-installation.md) | [IaaS installation scripts](docs/iaas-installation.md)
- [Integracion API](docs/es/api-integration.md) | [API integration](docs/api-integration.md)
- [Catalogo de errores](docs/es/error-catalog.md) | [Provider-agnostic error catalog](docs/error-catalog.md)
- [Arquitectura](docs/es/architecture.md) | [Architecture](docs/architecture.md)
- [Perfil de seguridad NIST](docs/es/nist-security-profile.md) | [NIST security profile](docs/nist-security-profile.md)
- [Contrato OpenAPI](docs/openapi.yaml)

## Inicio Rapido

```bash
cp .env.example .env
export SMS_OTP_SERVICE_API_TOKEN="replace-with-a-long-random-token"
export SMS_OTP_SECRET="replace-with-at-least-32-random-bytes"
export SMS_PHONE_HASH_SECRET="replace-with-a-separate-random-secret"
export SMS_MFA_SESSION_SECRET="replace-with-at-least-32-random-bytes"
export STORE_DRIVER=memory
export SMS_PROVIDER=twilio
export TWILIO_ACCOUNT_SID="AC..."
export TWILIO_API_KEY_SID="SK..."
export TWILIO_API_KEY_SECRET="..."
export TWILIO_MESSAGING_SERVICE_SID="MG..."
go run ./cmd/server
```

Health check:

```bash
curl http://localhost:8080/health
```

Enviar OTP:

```bash
curl -X POST http://localhost:8080/v1/sms-otp/send \
  -H "Authorization: Bearer ${SMS_OTP_SERVICE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"subject_id":"user-123","phone_number":"+15555550100"}'
```

Verificar OTP:

```bash
curl -X POST http://localhost:8080/v1/sms-otp/verify \
  -H "Authorization: Bearer ${SMS_OTP_SERVICE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"subject_id":"user-123","phone_number":"+15555550100","code":"123456"}'
```

La verificacion devuelve un `mfa_token`. Tu backend puede validarlo con:

```bash
curl -X POST http://localhost:8080/v1/sms-mfa/session/validate \
  -H "Authorization: Bearer ${SMS_OTP_SERVICE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"subject_id":"user-123","token":"<mfa_token>"}'
```

## Modelo de Seguridad

- No expongas este servicio directamente a navegadores ni apps moviles.
- Manten `SMS_OTP_SERVICE_API_TOKEN`, `SMS_OTP_SECRET`, credenciales de proveedor y secretos de sesion fuera del repositorio.
- El caller debe autorizar el sujeto y el telefono antes de llamar `/send`.
- El servicio no es fuente de verdad de usuarios, roles ni telefonos autorizados.
- El storage de challenges guarda un fingerprint HMAC del telefono, no el telefono en claro.
- Los logs no deben incluir codigos OTP, tokens de sesion, telefonos completos ni secretos.
- SMS OTP no es phishing-resistant; despliegues regulados deben ofrecer tambien un factor phishing-resistant para flujos de alto riesgo.

## Desarrollo

```bash
go test ./...
go build ./cmd/server
git diff --check
```

## Licencia

Apache-2.0. Ver [LICENSE](LICENSE).
