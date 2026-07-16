# Security Policy

## Supported Versions

The `main` branch receives security fixes.

## Reporting A Vulnerability

Do not open a public issue for a suspected vulnerability.

Report privately through GitHub Security Advisories for this repository, or contact the maintainers through the organization security channel.

Include:

- Affected endpoint or component.
- Reproduction steps.
- Expected and observed behavior.
- Any relevant logs with secrets, OTP codes, phone numbers, and tokens redacted.

## Security Expectations

- Deploy this service behind a trusted backend.
- Do not expose `/v1/*` directly to browsers or mobile apps.
- Keep service tokens, separated HMAC secrets, and provider credentials in a secret manager.
- Use fake phone numbers in tests and documentation.
- Do not log OTP codes, MFA tokens, full phone numbers, provider credentials, or request authorization headers.
- Treat SMS OTP as a restricted out-of-band factor and provide phishing-resistant MFA for privileged or high-risk regulated workflows.
