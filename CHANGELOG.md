# Changelog

## [0.4.0] - 2026-07-16

- Endurece el servicio con perfil de seguridad NIST para despliegues regulados.
- Evita persistir telefonos en claro; los challenges guardan `phone_hash` con HMAC y secreto separado.
- Agrega validacion estricta de telefono E.164, codigo OTP numerico, JSON sin campos desconocidos y limite de cuerpo.
- Exige secretos fuertes y configuracion de proveedor al iniciar el servicio.
- Agrega headers de seguridad en respuestas JSON.
- Documenta controles NIST, riesgos de SMS como factor restringido y baseline bancaria.

## [0.3.0] - 2026-07-16

- Renombra y prepara el proyecto como `sms-otp-factor-service` para publicacion reusable.
- Agrega documentacion publica de instalacion, configuracion, integracion API, seguridad y arquitectura ISO/IEC/IEEE 42010 con C4 Model.
- Agrega `.env.example`, `SECURITY.md` y licencia Apache-2.0.
- Cambia ejemplos y pruebas a telefonos ficticios y elimina nombres/dominios internos del contrato publico.
- Agrega `OTP_MESSAGE_TEMPLATE` para configurar el texto SMS sin acoplar marca al codigo.
- Endurece el script de deploy para exigir `GCP_PROJECT_ID` explicito y service account configurable.

## [0.2.0] - 2026-07-16

- Agrega Amazon Simple Notification Service como proveedor SMS seleccionable con `SMS_PROVIDER=amazon_sns`.
- Agrega contrato HTTP BIAN-aligned para `Customer Access Entitlement` manteniendo compatibilidad con `/v1/admin/*`.
- Documenta OpenAPI del contrato interno y aliases BIAN-aligned.

## [0.1.0] - 2026-07-16

- Crea microservicio independiente `sms-otp-factor-service` para factor OTP SMS.
- Implementa arquitectura hexagonal con dominio `smsotp`, casos de uso de aplicacion, puertos, adaptadores HTTP, Firestore y Twilio.
- Agrega endpoints internos para enviar OTP, verificar OTP y validar token SMS MFA.
- Agrega deploy productivo versionado a Cloud Run sin Cloud Build.
