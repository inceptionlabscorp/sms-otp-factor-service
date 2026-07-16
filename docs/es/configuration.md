# Configuracion

English: [Configuration](../configuration.md)

La configuracion se basa en variables de entorno. Los secretos deben inyectarse por el runtime o un gestor de secretos; nunca deben commitearse.

## Variables Core

| Variable | Requerida | Descripcion |
| --- | --- | --- |
| `PORT` | No | Puerto HTTP. Default: `8080`. |
| `SMS_OTP_SERVICE_API_TOKEN` | Si | Bearer token requerido por todos los endpoints `/v1/*`. |
| `SMS_OTP_SECRET` | Si | Secreto HMAC para hashear challenges OTP. Usa al menos 32 bytes aleatorios. |
| `SMS_PHONE_HASH_SECRET` | Si | Secreto HMAC separado para fingerprint de telefonos en storage. Usa al menos 32 bytes aleatorios. |
| `SMS_MFA_SESSION_SECRET` | Si | Secreto HMAC para firmar tokens de sesion MFA. Usa al menos 32 bytes aleatorios. |
| `OTP_MESSAGE_TEMPLATE` | No | Plantilla del SMS. Soporta `{{CODE}}` y `{{MINUTES}}`. |

Plantilla default:

```text
Your verification code is {{CODE}}. It expires in {{MINUTES}} minutes.
```

## Store

| Variable | Requerida | Descripcion |
| --- | --- | --- |
| `STORE_DRIVER` | No | `firestore`, `dynamodb` o `memory`. Default: `firestore`. |
| `GCP_PROJECT_ID` | Para Firestore | Proyecto Google Cloud usado por el adapter Firestore REST. |
| `GOOGLE_CLOUD_PROJECT` | Para Firestore | Variable alternativa cuando `GCP_PROJECT_ID` no existe. |
| `FIRESTORE_COLLECTION` | No | Coleccion de challenges. Default: `sms_otp_challenges`. |
| `AWS_REGION` | Para DynamoDB | Region AWS usada por el adapter DynamoDB. |
| `DYNAMODB_TABLE` | Para DynamoDB | Tabla DynamoDB. Default en scripts: `sms-otp-challenges`. |

Usa `memory` solo para desarrollo local y pruebas. Usa `firestore` en GCP y `dynamodb` en AWS.

## Seleccion de Proveedor SMS

| Variable | Requerida | Descripcion |
| --- | --- | --- |
| `SMS_PROVIDER` | No | `twilio` o `amazon_sns`. Default: `twilio`. |

## Twilio

Requeridas cuando `SMS_PROVIDER=twilio`.

| Variable | Descripcion |
| --- | --- |
| `TWILIO_ACCOUNT_SID` | Account SID de Twilio, normalmente empieza con `AC`. |
| `TWILIO_API_KEY_SID` | API key SID de Twilio, normalmente empieza con `SK`. |
| `TWILIO_API_KEY_SECRET` | Secreto de la API key. |
| `TWILIO_MESSAGING_SERVICE_SID` | Messaging Service SID, normalmente empieza con `MG`. |

El servicio usa HTTP raw contra Twilio y no requiere SDK de Twilio.

## Amazon SNS

Requeridas cuando `SMS_PROVIDER=amazon_sns`.

| Variable | Requerida | Descripcion |
| --- | --- | --- |
| `AWS_REGION` | Si | Region AWS para SNS, por ejemplo `us-east-1`. |
| `AWS_ACCESS_KEY_ID` | Solo fuera de roles | Access key con `sns:Publish`. |
| `AWS_SECRET_ACCESS_KEY` | Solo fuera de roles | Secret access key. |
| `AWS_SESSION_TOKEN` | No | Session token temporal. |
| `AWS_SNS_SMS_TYPE` | No | `Transactional` o `Promotional`. Default: `Transactional`. |
| `AWS_SNS_SENDER_ID` | No | Sender ID donde pais/carrier lo soporten. |

El adapter firma requests con AWS SigV4. En runtimes AWS puede usar role credentials por container metadata, por lo que access keys estaticas no son necesarias.

## Nombres Sugeridos de Secretos

| Variable de entorno | Nombre sugerido |
| --- | --- |
| `SMS_OTP_SERVICE_API_TOKEN` | `sms-otp-service-api-token` |
| `SMS_OTP_SECRET` | `sms-otp-secret` |
| `SMS_PHONE_HASH_SECRET` | `sms-phone-hash-secret` |
| `SMS_MFA_SESSION_SECRET` | `sms-mfa-session-secret` |
| `TWILIO_ACCOUNT_SID` | `twilio-account-sid` |
| `TWILIO_API_KEY_SID` | `twilio-api-key-sid` |
| `TWILIO_API_KEY_SECRET` | `twilio-api-key-secret` |
| `TWILIO_MESSAGING_SERVICE_SID` | `twilio-messaging-service-sid` |
| `AWS_ACCESS_KEY_ID` | `aws-sns-access-key-id` |
| `AWS_SECRET_ACCESS_KEY` | `aws-sns-secret-access-key` |
| `AWS_SESSION_TOKEN` | `aws-sns-session-token` |
