# Perfil de Seguridad NIST Para Despliegues SMS OTP Regulados

English: [NIST security profile](../nist-security-profile.md)

Este perfil mapea el servicio a controles alineados con NIST para entornos regulados. No es certificacion, informe de auditoria, ATO, autorizacion FedRAMP, evaluacion PCI ni opinion legal.

## Referencias NIST

- NIST SP 800-63B: Digital Identity Guidelines, Authentication and Lifecycle Management.
- NIST SP 800-53 Rev. 5: Security and Privacy Controls for Information Systems and Organizations.
- NIST SP 800-57 Part 1 Rev. 5: Recommendation for Key Management.

## Posicion de Aseguramiento

SMS OTP se trata como factor out-of-band para step-up authentication, no como autenticador phishing-resistant. Despliegues regulados deben ofrecer una alternativa phishing-resistant como passkeys/WebAuthn o autenticacion criptografica respaldada por hardware para flujos de mayor aseguramiento.

El servicio puede apoyar un despliegue estilo NIST SP 800-63B AAL2 solo cuando se combina con:

- Un primer factor separado controlado por el backend confiable.
- Canales protegidos autenticados en cada salto cliente-backend y backend-servicio.
- Controles NIST SP 800-53 Moderate o superiores alrededor de runtime, red, logs, llaves, personal y operaciones.
- Decision documentada de controles compensatorios por usar SMS como canal out-of-band restringido.

## Controles Tecnicos Implementados

| Area | Intencion | Implementacion |
| --- | --- | --- |
| Autenticacion servicio-a-servicio | Prevenir operaciones OTP no autenticadas. | Todos los endpoints `/v1/*` requieren bearer token. |
| Fortaleza de secretos | Reducir riesgo de llaves debiles. | El runtime no inicia si faltan service token, secretos HMAC o secreto de sesion de al menos 32 caracteres. |
| Separacion de llaves | Evitar que un secreto proteja datos no relacionados. | Secretos separados para OTP HMAC, fingerprint de telefono y firma de sesion MFA. |
| Minimizacion PII | Evitar telefonos en claro en el store. | Challenges persisten `phone_hash`, no `phone_number`. |
| Confidencialidad OTP | Evitar codigos OTP en claro. | Challenges persisten HMAC de sujeto, fingerprint telefonico, purpose, codigo y nonce. |
| Reduccion de replay | Evitar reuso de OTP valido. | Verificacion exitosa borra el challenge. |
| Reduccion brute-force | Limitar guessing. | Validacion de forma, rate limits y maximo de intentos. |
| Validacion de input | Reducir abuso del parser. | JSON estricto, E.164, OTP numerico y limite de cuerpo. |
| Proteccion de respuesta | Reducir leakage por navegador/intermediarios. | `Cache-Control: no-store`, CSP, `nosniff`, frame denial y no-referrer. |
| Abstraccion proveedor | Reducir concentracion de proveedor. | Twilio y Amazon SNS son adapters detras de un puerto. |

## Controles Requeridos de Despliegue

| Familia | Control requerido |
| --- | --- |
| AC - Access Control | Solo el backend confiable debe alcanzar `/v1/*`; restringe ingress con red privada, gateways, IAP, mTLS o controles equivalentes. |
| AU - Audit and Accountability | Registra metadata y decisiones en el backend sin OTP, tokens, telefonos completos ni secretos. |
| CM - Configuration Management | Gestiona variables y proveedores por change control aprobado. |
| IA - Identification and Authentication | Vincula telefonos mediante un proceso autenticado de recuperacion o security settings antes de llamar el servicio. |
| IR - Incident Response | Mantén procedimientos para rotacion, compromiso de proveedor, SIM-swap y abuso OTP. |
| RA - Risk Assessment | Documenta aceptacion de riesgo por SMS como factor out-of-band restringido. |
| SC - System and Communications Protection | Usa TLS 1.2+ en todos los saltos y preferentemente conectividad privada. |
| SI - System and Information Integrity | Monitorea volumen anomalo, fallos de entrega y fallos repetidos de verificacion. |
| SR - Supply Chain Risk Management | Pin de dependencias, revision de credenciales y escaneo de imagenes en CI/CD. |

## Baseline Bancaria

- Cifrado en reposo del store productivo con llaves administradas por plataforma o por cliente.
- Secret Manager, AWS Secrets Manager, Vault o equivalente para todos los secretos.
- Secretos separados por ambiente y tenant si hay multi-tenant isolation.
- Service account/role con minimo privilegio hacia el store.
- Sin ingress publico no autenticado a `/v1/*`.
- Integracion SIEM para senales de abuso y cambios de configuracion.
- Politica documentada de retencion para challenges y logs de entrega de proveedor.
- MFA phishing-resistant para operaciones privilegiadas o de alto riesgo.

## No Objetivos

- El servicio no verifica identidad de usuario.
- El servicio no decide si un telefono pertenece a un usuario.
- El servicio no satisface AAL3 por si solo.
- El servicio no vuelve phishing-resistant al SMS.
- El servicio no reemplaza controles organizacionales exigidos por reguladores.
