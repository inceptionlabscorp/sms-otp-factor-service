# AGENTS.md

Metodologia obligatoria para agentes que trabajen en `sms-otp-factor-service`.

## Reglas base

- No alucines.
- Siempre que te pidan una implementacion, hazlo real; no dejes mocks ni funcionalidades dummy.
- Cuando pidan documentar arquitectura o desarrollo, usa ISO/IEC/IEEE 42010 con C4 Model.
- README y documentacion publica deben mantenerse siempre en ingles y espanol.

## Alcance

- Este repo contiene solo el microservicio independiente de factor OTP SMS.
- El servicio genera/verifica OTP, envia SMS via Twilio raw y firma/valida tokens SMS MFA.
- No contiene UI ni reglas de autorizacion de usuarios; esas viven en el backend integrador.
- No almacenar secretos, OTPs en claro, tokens ni telefonos completos en logs.

## Validacion

```bash
go test ./...
go build ./cmd/server
git diff --check
```

## Deploy

- Cloud Build esta prohibido.
- Deploy productivo solo con script versionado:

```bash
./scripts/deploy-prod.sh
```
