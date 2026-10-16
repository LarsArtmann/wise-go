# Wise Error Codes Reference

Every non-2xx Wise response carries Wise's own granular, machine-readable
error codes in the structured envelope (`errors[].code` / `errors[].message`).
The SDK exposes them three ways, so consumers never parse an error message
again:

1. **`APIError.Details`** — the full envelope as
   `[]WiseErrorDetail{Code WiseErrorCode, Message string}`, in wire order.
2. **`errors.Is(err, wise.ErrX)`** — the curated wave-1 sentinels below match
   by value through any `fmt.Errorf("%w")` wrap chain.
3. **`APIError.HasCode(WiseErrorCode)`** — any code, curated or not
   (`WiseErrorCode` is an open enum: unknown wire values pass through).

The first granular code also lands in `ErrorContext()` as `wise_code` for log
pipelines.

## Granular codes

Codes Wise documents across the API surface (census against the vendored
OpenAPI snapshot, 2026-10-16). HTTP-status context describes where Wise
typically emits the code; the status — not the body code — always drives the
SDK's typed error and its retry family.

| Code                      | Constant                             | Typical status      | Curated sentinel           | Retry?  | Guidance                                                                                                                                      |
| ------------------------- | ------------------------------------ | ------------------- | -------------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `not.valid`               | `WiseErrorCodeNotValid`              | 400, 422            | `ErrNotValid`              | No      | A value Wise rejects as not valid (often an amount). Fix the request.                                                                         |
| `invalid_value`           | `WiseErrorCodeInvalidValue`          | 400, 422            | `ErrInvalidValue`          | No      | A field value is out of range or malformed. Fix the request.                                                                                  |
| `parameter_missing`       | `WiseErrorCodeParameterMissing`      | 400, 422            | `ErrParameterMissing`      | No      | A required parameter was absent. Fix the request.                                                                                             |
| `parameter_invalid`       | `WiseErrorCodeParameterInvalid`      | 400, 422            | `ErrParameterInvalid`      | No      | A parameter's value failed validation. Fix the request.                                                                                       |
| `missing_value`           | — (open enum)                        | 400                 | —                          | No      | A required value was absent. Fix the request.                                                                                                 |
| `invalid_request`         | `WiseErrorCodeInvalidRequest`        | 400                 | `ErrInvalidRequest`        | No      | Generic malformed request. Fix the request.                                                                                                   |
| `unauthorized`            | `WiseErrorCodeUnauthorized`          | 401                 | `ErrUnauthorized`          | No      | Missing or invalid credentials. Fix the API key.                                                                                              |
| `forbidden`               | `WiseErrorCodeForbidden`             | 403                 | `ErrForbidden`             | No      | Credentials lack permission for the operation.                                                                                                |
| `resource_not_found`      | `WiseErrorCodeResourceNotFound`      | 404                 | `ErrResourceNotFound`      | No      | The addressed resource does not exist (or is not visible to this token).                                                                      |
| `operation_not_supported` | `WiseErrorCodeOperationNotSupported` | 405/422             | `ErrOperationNotSupported` | No      | The API surface does not support the operation for this resource or corridor.                                                                 |
| `rate_limit_exceeded`     | `WiseErrorCodeRateLimitExceeded`     | 429                 | `ErrRateLimitExceeded`     | Yes     | Wise throttled the request. `RateLimitError.RetryAfter` (from the `Retry-After` header) is honored automatically by the SDK's retry executor. |
| `system_error`            | `WiseErrorCodeSystemError`           | 5xx (sometimes 4xx) | `ErrSystemError`           | Usually | Wise-side failure. Inside a 5xx the error is a retryable `*ServerError`; inside a 4xx it is not — check `ErrorFamily()`.                      |
| `invalid_file_extension`  | — (open enum)                        | 400                 | —                          | No      | An upload (e.g. statement of funds) has an unsupported file extension.                                                                        |
| `unexpected.error`        | — (open enum)                        | 5xx/4xx             | —                          | Usually | Wise's generic failure code. Treat like `system_error`: retry only when the typed error says retryable.                                       |

Unknown codes: Wise adds codes over time. They flow through `Details` and
`HasCode` unmodified — switch on the constants for known cases and handle the
default explicitly.

## Typed errors by HTTP status

The status line drives the type; see the README's error table for the full
mapping (400 → `*BadRequestError`, 401/403 → `*AuthError`, 403 + 2FA headers
→ `*SCAChallengeError`, 404 → `*NotFoundError`, 429 → `*RateLimitError`,
5xx → `*ServerError`, everything else → `*APIError`). Only `RateLimitError`
and `ServerError` are retryable.

## Funding errors are a separate surface

Rejected funding attempts (`POST .../transfers/{id}/payments`) carry their own
code vocabulary in `FundingResponse` — typed as `FundingErrorCode`
(`balance.insufficient-funds`, `payment.option-unavailable`, ...), a separate
open enum on the parsed funding result, not part of the HTTP error envelope.

## Matching recipes

```go
// Curated code, wrap-proof:
if errors.Is(err, wise.ErrResourceNotFound) {
    fmt.Println("gone; refresh before retrying")
}

// Any code (open enum):
if apiErr, ok := errors.AsType[*wise.APIError](err); ok {
    if apiErr.HasCode(wise.WiseErrorCodeNotValid) {
        fmt.Println("do not retry; fix the request")
    }

    for _, detail := range apiErr.Details {
        fmt.Printf("%s: %s\n", detail.Code, detail.Message)
    }
}

// The type still carries the fields:
if rl, ok := errors.AsType[*wise.RateLimitError](err); ok {
    fmt.Printf("retry after %s\n", rl.RetryAfter)
}
```
