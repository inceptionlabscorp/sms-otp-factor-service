# Catalogo de errores agnostico a proveedor

English: [Provider-agnostic error catalog](../error-catalog.md)

SMS OTP Factor Service devuelve codigos de error estables y agnosticos a proveedor en el campo JSON `error`.

La API nunca expone nombres de error especificos de Twilio, Amazon SNS, carrier, proveedor cloud, credenciales o SDK en respuestas publicas. Esos detalles pertenecen solo a logs confiables del servidor y telemetria operacional.

```json
{
  "error": "invalid_code"
}
```

## Contrato API

| HTTP status | Codigo de error | Significado | Accion del caller |
| --- | --- | --- | --- |
| 401 | `unauthorized` | Bearer token de servicio faltante, malformado o invalido. | No reintentar con el mismo token. Corregir credenciales de servicio. |
| 404 | `not_found` | La ruta no existe. | Corregir endpoint o metodo HTTP. |
| 400 | `invalid_request` | JSON malformado, campos desconocidos, telefono E.164 invalido, campos obligatorios faltantes o formato OTP invalido. | Corregir el request antes de reintentar. |
| 429 | `rate_limited` | El envio OTP excedio la politica del challenge activo. | Esperar antes de solicitar otro challenge. |
| 400 | `challenge_expired` | El challenge OTP activo expiro. | Iniciar un nuevo challenge de envio. |
| 400 | `invalid_code` | La verificacion OTP no puede aceptarse. Cubre intencionalmente challenge faltante, codigo incorrecto, telefono distinto e intentos agotados. | Solicitar un nuevo codigo o reintentar segun politica local sin revelar cual condicion ocurrio. |
| 503 | `service_not_configured` | Falta configuracion requerida, como credenciales SMS, secretos HMAC, store o secreto de firma de sesion MFA. | Corregir configuracion del despliegue. |
| 500 | `operation_failed` | La operacion fallo despues de validar el request, incluyendo fallas de entrega del proveedor o persistencia. | Tratar como falla operacional. Reintentar solo segun politica del backend confiable. |
| 503 | `timeout` | El handler HTTP excedio su timeout configurado. | Reintentar segun politica de idempotencia y rate limit. |

## Notas de seguridad

- `invalid_code` es amplio de forma intencional para evitar enumeracion de cuentas y challenges.
- `challenge_expired` es distinto para que el caller pueda pedir un nuevo OTP de forma segura.
- `operation_failed` no debe parsearse para comportamiento especifico del proveedor.
- Los logs no deben incluir OTP, telefonos completos, tokens MFA, bearer tokens, secretos de proveedor ni errores raw de credenciales.
- Los callers confiables deben mapear estos codigos a sus propios errores de dominio y no deben mostrar detalles internos de remediacion a usuarios finales.

## Compatibilidad

Solo el codigo `error` es estable. Texto humano, mensajes de log, clases de excepcion de proveedor y errores internos de dominio no forman parte del contrato publico de la API.
