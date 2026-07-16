# NIST Security Profile For Regulated SMS OTP Deployments

Español: [Perfil de Seguridad NIST](es/nist-security-profile.md)

This profile maps the service to NIST-aligned controls for regulated environments. It is not a certification, audit report, ATO, FedRAMP authorization, PCI assessment, or legal opinion.

## NIST References

- NIST SP 800-63B: Digital Identity Guidelines, Authentication and Lifecycle Management.
- NIST SP 800-53 Rev. 5: Security and Privacy Controls for Information Systems and Organizations.
- NIST SP 800-57 Part 1 Rev. 5: Recommendation for Key Management.

## Assurance Position

SMS OTP is treated as an out-of-band factor for step-up authentication, not as a phishing-resistant authenticator. Regulated deployments should provide a phishing-resistant alternative such as passkeys/WebAuthn or hardware-backed cryptographic authentication for higher assurance flows.

The service is designed to support a NIST SP 800-63B AAL2-style deployment only when it is combined with:

- A separate first factor controlled by the trusted backend.
- Authenticated protected channels for every client-to-backend and backend-to-service request.
- Tailored NIST SP 800-53 Moderate or stronger controls around the runtime, network, logs, keys, personnel, and operations.
- A documented compensating-control decision for using SMS as a restricted out-of-band channel.

## Implemented Technical Controls

| Area | Control intent | Implementation |
| --- | --- | --- |
| Service-to-service authentication | Prevent unauthenticated OTP operations. | All `/v1/*` endpoints require `Authorization: Bearer <SMS_OTP_SERVICE_API_TOKEN>`. |
| Secret strength | Reduce weak-key risk. | Runtime refuses to start unless service token, OTP HMAC secret, phone hash secret, and MFA session secret are at least 32 characters. |
| Key separation | Avoid one secret protecting unrelated data. | Separate secrets for OTP HMAC, phone fingerprint HMAC, and MFA session signing. HMAC inputs include context labels. |
| PII minimization | Avoid storing phone numbers in the challenge store. | Challenges persist `phone_hash`, never `phone_number`. |
| OTP confidentiality | Avoid storing OTP codes in clear text. | Challenges persist HMAC of subject, phone fingerprint, purpose, code, and nonce. |
| Replay reduction | Prevent repeated use of a valid OTP. | Successful verification deletes the challenge. Expired challenges are deleted on verification. |
| Brute-force reduction | Limit guessing. | Code shape validation, request rate limits, and maximum failed attempts. |
| Input validation | Reduce injection and parser abuse. | Strict JSON decoder rejects unknown fields, multiple JSON documents, malformed JSON, non-E.164 phone numbers, and malformed OTP codes. |
| Response protection | Reduce browser and intermediary leakage. | JSON responses include `Cache-Control: no-store`, CSP, `nosniff`, frame denial, and no-referrer headers. |
| Provider abstraction | Reduce vendor concentration risk. | Twilio and Amazon SNS are adapter implementations behind the SMS gateway port. |

## Required Deployment Controls

The following controls are outside the codebase and must be implemented by the deploying organization.

| Control family | Required deployment control |
| --- | --- |
| AC - Access Control | Only the trusted backend may reach `/v1/*`; restrict ingress with private networking, mTLS-capable gateways, identity-aware proxies, or equivalent controls. |
| AU - Audit and Accountability | Log request metadata and decision outcomes in the trusted backend without OTP codes, MFA tokens, full phone numbers, or secrets. |
| CM - Configuration Management | Manage environment variables and provider selection through approved change control. |
| IA - Identification and Authentication | Bind phone numbers through an authenticated account recovery or security settings process before this service is called. |
| IR - Incident Response | Maintain procedures for token/secret rotation, provider compromise, SIM-swap indicators, and OTP abuse. |
| RA - Risk Assessment | Document the risk acceptance for SMS as a restricted out-of-band factor. |
| SC - System and Communications Protection | Use TLS 1.2+ at every hop and prefer private service-to-service connectivity. |
| SI - System and Information Integrity | Monitor abnormal OTP request volume, delivery failure spikes, and repeated verification failures. |
| SR - Supply Chain Risk Management | Pin dependencies, review provider credentials, and scan container images in CI/CD. |

## Banking Deployment Baseline

For regulated financial institutions, use this service only behind a dedicated identity or access-control backend.

Minimum baseline:

- Production store encryption at rest using platform-managed or customer-managed keys.
- Secret Manager, AWS Secrets Manager, Vault, or equivalent for all secrets.
- Separate secrets per environment and per tenant where multi-tenant isolation is required.
- Dedicated service account with least privilege to the challenge store.
- No public unauthenticated ingress to `/v1/*`; public `/health` only if it exposes no sensitive data.
- SIEM integration for abuse signals and configuration changes.
- Documented retention policy for challenge records and provider delivery logs.
- Phishing-resistant MFA option for privileged or high-risk operations.

## Non-Goals

- The service does not verify user identity.
- The service does not decide whether a phone number belongs to a user.
- The service does not satisfy AAL3 by itself.
- The service does not make SMS phishing-resistant.
- The service does not replace organizational controls required by regulators.
