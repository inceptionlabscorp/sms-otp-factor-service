# Changelog

## [0.5.6] - 2026-07-17

- EN: Adds a provider-agnostic API error catalog and standardizes HTTP error responses.
- ES: Agrega un catalogo de errores API agnostico a proveedor y estandariza respuestas HTTP de error.
- EN: Documents stable error codes in English, Spanish and OpenAPI.
- ES: Documenta codigos de error estables en ingles, espanol y OpenAPI.

## [0.5.5] - 2026-07-17

- EN: Removes internal agent instructions from the public repository and ignores local agent metadata.
- ES: Elimina instrucciones internas de agentes del repositorio publico e ignora metadata local de agentes.

## [0.5.4] - 2026-07-16

- EN: Fixes shell default handling for `OTP_MESSAGE_TEMPLATE` so `{{CODE}}` and `{{MINUTES}}` are preserved in deployed environments.
- ES: Corrige el manejo del default shell de `OTP_MESSAGE_TEMPLATE` para preservar `{{CODE}}` y `{{MINUTES}}` en ambientes desplegados.
- EN: Adds test coverage for the default OTP message body.
- ES: Agrega cobertura de prueba para el cuerpo de mensaje OTP por defecto.

## [0.5.3] - 2026-07-16

- EN: Adds AWS Lambda free-tier installer using Lambda Web Adapter, Function URL, ECR, DynamoDB and IAM.
- ES: Agrega instalador AWS Lambda compatible con free tier usando Lambda Web Adapter, Function URL, ECR, DynamoDB e IAM.
- EN: Documents Lambda as the recommended AWS free-tier runtime.
- ES: Documenta Lambda como runtime AWS recomendado para free tier.
- EN: Grants both Lambda Function URL permissions required by AWS for new public Function URLs.
- ES: Concede ambos permisos requeridos por AWS para Function URLs publicas nuevas.

## [0.5.2] - 2026-07-16

- EN: Adds AWS App Runner availability preflight before provisioning resources.
- ES: Agrega preflight de disponibilidad de AWS App Runner antes de provisionar recursos.
- EN: Documents the App Runner account/region subscription requirement.
- ES: Documenta el requisito de habilitacion/suscripcion de App Runner por cuenta/region.

## [0.5.1] - 2026-07-16

- EN: Makes README bilingual and adds Spanish mirrors for the public documentation under `docs/es/`.
- ES: Convierte el README en bilingue y agrega espejos en espanol para la documentacion publica bajo `docs/es/`.
- EN: Adds the project rule that public README and documentation must always remain in English and Spanish.
- ES: Agrega la regla de proyecto para mantener README y documentacion publica siempre en ingles y espanol.

## [0.5.0] - 2026-07-16

- Agrega instaladores IaaS reales para GCP y AWS.
- Agrega adapter DynamoDB productivo para despliegues AWS.
- Agrega resolucion de credenciales AWS por role/container metadata para evitar access keys estaticas en App Runner.
- Agrega scripts `install-gcp.sh` y `install-aws.sh` para provisionar runtime, store, secretos, IAM y container deployment.
- Documenta instalacion GCP/AWS en `docs/iaas-installation.md`.

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
