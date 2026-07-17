# Provider-Agnostic Error Catalog

Español: [Catalogo de errores](es/error-catalog.md)

The SMS OTP Factor Service returns stable, provider-agnostic error codes in the JSON field `error`.

The API never exposes Twilio, Amazon SNS, carrier, cloud-provider, credential, or SDK-specific error names in public responses. Provider details belong only in trusted server-side logs and operational telemetry.

```json
{
  "error": "invalid_code"
}
```

## API Contract

| HTTP status | Error code | Meaning | Caller action |
| --- | --- | --- | --- |
| 401 | `unauthorized` | Missing, malformed, or invalid service bearer token. | Do not retry with the same token. Fix service credentials. |
| 404 | `not_found` | The route does not exist. | Fix the endpoint path or HTTP method. |
| 400 | `invalid_request` | Malformed JSON, unknown fields, invalid E.164 phone number, missing required fields, or invalid OTP format. | Fix the request before retrying. |
| 429 | `rate_limited` | OTP send request exceeded the active challenge policy. | Wait before requesting another challenge. |
| 400 | `challenge_expired` | The active OTP challenge expired. | Start a new send challenge. |
| 400 | `invalid_code` | The OTP verification cannot be accepted. This intentionally covers missing challenge, wrong code, phone mismatch, and exhausted attempts. | Ask for a new code or retry according to local policy without revealing which condition occurred. |
| 503 | `service_not_configured` | Required service configuration is missing, such as SMS provider credentials, HMAC secrets, store, or MFA session signing secret. | Fix deployment configuration. |
| 500 | `operation_failed` | The operation failed after request validation, including provider delivery failures or persistence failures. | Treat as operational failure. Retry only according to trusted backend policy. |
| 503 | `timeout` | The HTTP handler exceeded its configured timeout. | Retry according to idempotency and rate-limit policy. |

## Security Notes

- `invalid_code` is intentionally broad to avoid account and challenge enumeration.
- `challenge_expired` is distinct so callers can safely prompt for a new OTP.
- `operation_failed` must not be parsed for provider-specific behavior.
- Logs must not include OTP values, full phone numbers, MFA tokens, bearer tokens, provider secrets, or raw credential errors.
- Trusted callers should map these codes to their own domain errors and should not show internal remediation details to end users.

## Compatibility

Only the `error` code is stable. Human-readable text, log messages, provider exception classes, and internal domain errors are not part of the public API contract.
