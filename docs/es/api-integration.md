# Integracion API

English: [API integration](../api-integration.md)

Este servicio esta pensado para uso backend-to-backend. Un navegador o app movil nunca debe llamarlo directamente.

## Responsabilidades de Integracion

Tu backend confiable debe:

- Autenticar al usuario con tu identity provider.
- Autorizar al usuario y la accion.
- Resolver el telefono E.164 autorizado desde tu perfil de usuario o configuracion de seguridad.
- Llamar este servicio con `subject_id` y `phone_number` autorizado.
- Guardar el `mfa_token` solo dentro de una frontera de sesion protegida.
- Llamar `/session/validate` antes de permitir acciones protegidas que requieren SMS OTP.
- Ofrecer un factor phishing-resistant alternativo para flujos privilegiados o de alto riesgo en despliegues regulados.

Este servicio debe:

- Crear y guardar challenges OTP sin almacenar OTP en claro.
- Guardar fingerprints de telefono en lugar de telefonos raw.
- Enviar OTP mediante el proveedor SMS configurado.
- Verificar codigos OTP.
- Emitir y validar tokens de sesion MFA SMS de corta duracion.

## Autenticacion

Todos los endpoints `/v1/*` requieren:

```http
Authorization: Bearer <SMS_OTP_SERVICE_API_TOKEN>
```

`GET /health` no requiere autenticacion.

## Endpoints Canonicos

### Enviar OTP

```http
POST /v1/sms-otp/send
Authorization: Bearer <SMS_OTP_SERVICE_API_TOKEN>
Content-Type: application/json

{
  "subject_id": "user-123",
  "phone_number": "+15555550100",
  "purpose": "sms_mfa"
}
```

Respuesta:

```json
{
  "status": "sent"
}
```

### Verificar OTP

```http
POST /v1/sms-otp/verify
Authorization: Bearer <SMS_OTP_SERVICE_API_TOKEN>
Content-Type: application/json

{
  "subject_id": "user-123",
  "phone_number": "+15555550100",
  "code": "123456",
  "purpose": "sms_mfa"
}
```

Respuesta:

```json
{
  "mfa_token": "<signed-token>",
  "expires_in": 900,
  "method": "sms"
}
```

### Validar Sesion MFA

```http
POST /v1/sms-mfa/session/validate
Authorization: Bearer <SMS_OTP_SERVICE_API_TOKEN>
Content-Type: application/json

{
  "subject_id": "user-123",
  "token": "<signed-token>"
}
```

Respuesta:

```json
{
  "valid": true
}
```

## Aliases BIAN-Aligned

| Endpoint canonico | Alias BIAN-aligned |
| --- | --- |
| `POST /v1/sms-otp/send` | `POST /v1/bian/customer-access-entitlement/sms-otp/initiate` |
| `POST /v1/sms-otp/verify` | `POST /v1/bian/customer-access-entitlement/sms-otp/execute` |
| `POST /v1/sms-mfa/session/validate` | `POST /v1/bian/customer-access-entitlement/sms-mfa-session/evaluate` |

Estos aliases son BIAN-aligned, no BIAN-certified. Las rutas legacy `/v1/admin/*` siguen disponibles como aliases backward-compatible para integraciones privadas existentes.

## Errores

El servicio devuelve errores JSON:

```json
{
  "error": "invalid_code"
}
```

Errores comunes:

| Error | Significado |
| --- | --- |
| `unauthorized` | Bearer token faltante o invalido. |
| `invalid_request` | Request invalido. |
| `cooldown_active` | Challenge solicitado demasiado pronto. |
| `challenge_not_found` | No existe challenge activo. |
| `challenge_expired` | El challenge activo expiro. |
| `invalid_code` | Codigo incorrecto. |
| `too_many_attempts` | Se excedieron los intentos permitidos. |

## OpenAPI

El contrato machine-readable esta en [openapi.yaml](../openapi.yaml).
