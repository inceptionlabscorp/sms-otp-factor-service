# API Integration

Español: [Integracion API](es/api-integration.md)

This service is intended for backend-to-backend use. A browser or mobile app should never call it directly.

## Integration Responsibilities

Your trusted backend must:

- Authenticate the user with your identity provider.
- Authorize the user and action.
- Resolve the authorized E.164 phone number from your own user profile or security settings.
- Call this service with `subject_id` and the authorized `phone_number`.
- Store the returned `mfa_token` only in a protected server/client session boundary that matches your security model.
- Call `/session/validate` before allowing protected actions that require SMS OTP.
- Offer a phishing-resistant alternative for privileged or high-risk flows in regulated deployments.

This service must:

- Create and store OTP challenges without storing the OTP in clear text.
- Store phone fingerprints instead of raw phone numbers.
- Send the OTP through the configured SMS provider.
- Verify submitted OTP codes.
- Issue and validate short-lived SMS MFA session tokens.

## Authentication

All `/v1/*` endpoints require:

```http
Authorization: Bearer <SMS_OTP_SERVICE_API_TOKEN>
```

`GET /health` is unauthenticated.

## Canonical Endpoints

### Send OTP

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

Response:

```json
{
  "status": "sent"
}
```

### Verify OTP

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

Response:

```json
{
  "mfa_token": "<signed-token>",
  "expires_in": 900,
  "method": "sms"
}
```

### Validate MFA Session

```http
POST /v1/sms-mfa/session/validate
Authorization: Bearer <SMS_OTP_SERVICE_API_TOKEN>
Content-Type: application/json

{
  "subject_id": "user-123",
  "token": "<signed-token>"
}
```

Response:

```json
{
  "valid": true
}
```

## BIAN-Aligned Aliases

The service also exposes internal BIAN-aligned aliases using Customer Access Entitlement language and BIAN-style action terms:

| Canonical endpoint | BIAN-aligned alias |
| --- | --- |
| `POST /v1/sms-otp/send` | `POST /v1/bian/customer-access-entitlement/sms-otp/initiate` |
| `POST /v1/sms-otp/verify` | `POST /v1/bian/customer-access-entitlement/sms-otp/execute` |
| `POST /v1/sms-mfa/session/validate` | `POST /v1/bian/customer-access-entitlement/sms-mfa-session/evaluate` |

Legacy `/v1/admin/*` routes remain available as backward-compatible aliases for existing private integrations.

These aliases are intentionally described as BIAN-aligned, not BIAN-certified.

## Error Handling

The service returns JSON errors:

```json
{
  "error": "invalid_code"
}
```

Common errors:

| Error | Meaning |
| --- | --- |
| `unauthorized` | Missing or invalid service bearer token. |
| `invalid_request` | Request body is malformed or missing required fields. |
| `cooldown_active` | A challenge was requested too soon. |
| `challenge_not_found` | No active challenge exists for the subject and purpose. |
| `challenge_expired` | The active challenge expired. |
| `invalid_code` | Code is wrong. |
| `too_many_attempts` | The challenge exceeded allowed verification attempts. |

## OpenAPI

The machine-readable contract is maintained in [openapi.yaml](openapi.yaml).
