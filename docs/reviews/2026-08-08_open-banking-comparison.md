# wise-go SDK vs. Wise Open Banking API — Comparison Analysis

**Date:** 2026-08-08  
**Author:** Research session  
**Sources:** wise-go source code analysis + [Wise Open Banking docs](https://docs.wise.com/guides/developer/open-banking) + [Production OIDC config](https://wise.com/openbanking/.well-known/openid-configuration)

---

## Executive Summary

wise-go and the Wise Open Banking API are **fundamentally different APIs for different audiences**. wise-go wraps Wise's proprietary Public API for account owners. The Open Banking API is a standards-compliant (OB UK 3.1.11) interface for Financially Regulated Third Party Providers (TPPs) accessing Wise user data on behalf of end users with explicit consent.

They are **not interchangeable** and serve different use cases.

---

## 1. Audience & Purpose

|                   | wise-go SDK                                         | Open Banking API                                                                            |
| ----------------- | --------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| **Audience**      | Any developer / individual with a Wise account      | Financially Regulated TPPs (AISPs, PISPs, CBPIIs)                                           |
| **Prerequisites** | Wise account + API key (from web UI)                | OB Directory membership or eIDAS certificate                                                |
| **Standard**      | Wise proprietary API                                | Open Banking UK standard v3.1.11 (OBIE)                                                     |
| **Purpose**       | Direct programmatic access to your own Wise account | Account information access, payment initiation, funds confirmation — on behalf of end users |
| **Consent model** | None — token grants access to own account           | User-driven consent flow (SCA required, browser-based authorization)                        |

---

## 2. Authentication

|                         | wise-go SDK                           | Open Banking API                                                             |
| ----------------------- | ------------------------------------- | ---------------------------------------------------------------------------- |
| **Method**              | Bearer token (static API key)         | mTLS (RFC 8705) + OAuth2 Hybrid Flow                                         |
| **Credentials**         | API key string from Wise UI           | OBWAC/OBSeal/eIDAS certificates + `client_id`                                |
| **Consent**             | None                                  | Mandatory SCA browser flow (user authorizes each consent)                    |
| **Token lifetime**      | Static, no expiry                     | access_token ~12h, refresh_token ~20 years                                   |
| **Token endpoint**      | N/A                                   | `POST /open-banking/auth/token`                                              |
| **Authorization**       | N/A                                   | `GET /openbanking/authorize` (browser redirect, Hybrid Flow `code id_token`) |
| **Client registration** | N/A                                   | Dynamic Client Registration (`POST /open-banking/v3.3/register`)             |
| **Cert CN matching**    | N/A                                   | Client certificate CN must match `clientId`                                  |
| **Token auth method**   | Header: `Authorization: Bearer {key}` | `tls_client_auth` (mTLS on token endpoint)                                   |

### wise-go Authentication

```go
// Simple bearer token, set on every request
func (c *Client) setAuth(req *http.Request) {
    if c.apiKey != "" {
        req.Header.Set("Authorization", "Bearer "+c.apiKey)
    }
}
```

`Authenticate(ctx)` validates the key by calling `ListProfiles` internally. No OAuth, no token refresh, no session management.

### Open Banking Authentication Flow

1. **Create access token** (client_credentials grant) → `POST /open-banking/auth/token`
2. **Create consent** → `POST /open-banking/v3.1.11/aisp/account-access-consents` (specify permissions)
3. **User authorization** (Hybrid Flow) → `GET /openbanking/authorize` with signed JWT Request Object
   - User logs in → 2FA → selects profile → reviews & authorizes → redirect with `code` + `id_token`
4. **Exchange code for token** (authorization_code grant) → `POST /open-banking/auth/token`
5. **Refresh as needed** (refresh_token grant)

### OIDC Configuration (from production `.well-known`)

| Field                                         | Value                                                                                             |
| --------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| `issuer`                                      | `https://openbanking.transferwise.com`                                                            |
| `authorization_endpoint`                      | `https://transferwise.com/openbanking/authorize`                                                  |
| `token_endpoint`                              | `https://openbanking.transferwise.com/open-banking/auth/token`                                    |
| `registration_endpoint`                       | `https://openbanking.transferwise.com/open-banking/v3.3/register`                                 |
| `grant_types_supported`                       | `authorization_code`, `client_credentials`, `refresh_token`                                       |
| `response_types_supported`                    | `code id_token` (Hybrid Flow only)                                                                |
| `scopes_supported`                            | `accounts`, `payments`, `fundsconfirmations`, `openid`, `openbanking`, `cop`, `name-verification` |
| `token_endpoint_auth_methods_supported`       | `tls_client_auth`                                                                                 |
| `id_token_signing_alg_values_supported`       | `PS256`                                                                                           |
| `request_object_signing_alg_values_supported` | `HS256`, `PS256`, `ES256`                                                                         |
| `subject_types_supported`                     | `pairwise`                                                                                        |
| `acr_values_supported`                        | `urn:openbanking:psd2:sca`                                                                        |

---

## 3. Base URLs

| Environment                 | wise-go SDK                             | Open Banking API                                                                                                     |
| --------------------------- | --------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| **Production**              | `https://api.wise.com`                  | `https://openbanking.transferwise.com`                                                                               |
| **Sandbox**                 | `https://api.sandbox.transferwise.tech` | `https://openbanking.wise-sandbox.com`                                                                               |
| **Authorization (browser)** | N/A                                     | `https://transferwise.com/openbanking/authorize` (prod) / `https://wise-sandbox.com/openbanking/authorize` (sandbox) |

---

## 4. Endpoint Coverage

### wise-go SDK (3 read-only endpoints)

| Method | Endpoint                                                                 | SDK Function                               | Source               |
| ------ | ------------------------------------------------------------------------ | ------------------------------------------ | -------------------- |
| `GET`  | `/v2/profiles`                                                           | `ListProfiles` / `Authenticate` / `Health` | `profiles.go:15`     |
| `GET`  | `/v4/profiles/{profileId}/balances`                                      | `ListBalances` / `GetBalance`              | `balances.go:18`     |
| `GET`  | `/v1/profiles/{profileId}/balance-statements/{balanceId}/statement.json` | `ListTransactions`                         | `transactions.go:24` |

### Open Banking API (~17 endpoints, read + write)

#### AISP (Account Information Service Provider)

| Method   | Endpoint                                                         | Operation            |
| -------- | ---------------------------------------------------------------- | -------------------- |
| `POST`   | `/open-banking/v3.1.11/aisp/account-access-consents`             | Create consent       |
| `GET`    | `/open-banking/v3.1.11/aisp/account-access-consents/{consentId}` | Get consent          |
| `DELETE` | `/open-banking/v3.1.11/aisp/account-access-consents/{consentId}` | Delete consent       |
| `GET`    | `/open-banking/v3.1.11/aisp/accounts`                            | List all accounts    |
| `GET`    | `/open-banking/v3.1.11/aisp/accounts/{AccountId}`                | Get specific account |
| `GET`    | `/open-banking/v3.1.11/aisp/accounts/{AccountId}/balances`       | Get account balances |
| `GET`    | `/open-banking/v3.1.11/aisp/accounts/{AccountId}/transactions`   | Get transactions     |
| `GET`    | `/open-banking/v3.1.11/aisp/accounts/{AccountId}/direct-debits`  | Get direct debits    |

#### PISP (Payment Initiation Service Provider)

| Method | Endpoint                                                    | Operation                            |
| ------ | ----------------------------------------------------------- | ------------------------------------ |
| `POST` | `/open-banking/v3.1.11/pisp/domestic-payment-consents`      | Create domestic payment consent      |
| `POST` | `/open-banking/v3.1.11/pisp/domestic-payments`              | Create domestic payment              |
| `GET`  | `/open-banking/v3.1.11/pisp/domestic-payments/{id}`         | Get domestic payment                 |
| `POST` | `/open-banking/v3.1.11/pisp/international-payment-consents` | Create international payment consent |
| `POST` | `/open-banking/v3.1.11/pisp/international-payments`         | Create international payment         |
| `GET`  | `/open-banking/v3.1.11/pisp/international-payments/{id}`    | Get international payment            |

#### OAuth/OIDC

| Method | Endpoint                      | Purpose                          |
| ------ | ----------------------------- | -------------------------------- |
| `POST` | `/open-banking/auth/token`    | Token endpoint (all grant types) |
| `GET`  | `/openbanking/authorize`      | Authorization endpoint (browser) |
| `POST` | `/open-banking/v3.3/register` | Dynamic Client Registration      |

#### CBPII (Card-Based Payment Instrument Issuer)

- Scope `fundsconfirmations` supported for confirming availability of funds.

### Coverage Comparison

| Resource               | wise-go                                                        | Open Banking                            |
| ---------------------- | -------------------------------------------------------------- | --------------------------------------- |
| Profiles/Accounts      | `GET /v2/profiles`                                             | `GET /aisp/accounts` + `/{id}`          |
| Balances               | `GET /v4/profiles/{id}/balances`                               | `GET /aisp/accounts/{id}/balances`      |
| Transactions           | `GET /v1/profiles/{id}/balance-statements/{id}/statement.json` | `GET /aisp/accounts/{id}/transactions`  |
| Direct Debits          | Not covered                                                    | `GET /aisp/accounts/{id}/direct-debits` |
| Consents               | N/A                                                            | Create/Get/Delete                       |
| Domestic Payments      | Not covered                                                    | Create + Get                            |
| International Payments | Not covered                                                    | Create + Get                            |
| Payment Consents       | Not covered                                                    | Domestic + international                |
| Funds Confirmation     | Not covered                                                    | CBPII scope                             |

---

## 5. Data Models

|                            | wise-go SDK                                                                  | Open Banking API                                                                         |
| -------------------------- | ---------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| **Standard**               | Wise proprietary JSON                                                        | Open Banking UK (OBReadAccount6, OBReadBalance1, OBReadTransaction6)                     |
| **Account type**           | `Profile` (Personal/Business) + `Balance` (per currency)                     | `Account` (per currency, with SortCode/IBAN scheme names, BIC `TRWIGB22`)                |
| **Money**                  | `Money{Cents int64, Currency}` — integer cents, no float exposed             | OB amount objects (string-based `Amount` + `Currency`)                                   |
| **Transaction fees**       | Embedded in transaction (`Fees`, `Total` fields)                             | Separate DEBIT transactions referencing the original                                     |
| **Transaction window**     | Unlimited (date-range query)                                                 | Max 450 days per query; consent age restricts history (>90d consent → only 90d lookback) |
| **IDs**                    | Branded types (`ProfileID int64`, `BalanceID int64`, `TransactionID string`) | Opaque `AccountId` strings; transaction IDs differ between OB v3.1 and v3.1.11           |
| **Account identification** | `profileId` + `balanceId` (Wise internal)                                    | `UK.OBIE.SortCodeAccountNumber` or `UK.OBIE.IBAN` + `UK.OBIE.BICFI` (`TRWIGB22`)         |
| **Account subtype**        | `Standard` / `Savings`                                                       | `EMoney` (Wise accounts are e-money accounts)                                            |

### wise-go Type System (two-layer)

**Public types** (clean Go types):

| Type          | Key Fields                                                                                                                                                                                                                                           |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Profile`     | `ID ProfileID`, `Type ProfileType`, `Name string`, `Email string`, `CreatedAt time.Time`                                                                                                                                                             |
| `Balance`     | `ID BalanceID`, `Currency Currency`, `Type BalanceType`, `Name string`, `Amount Money`, `Reserved Money`, `Visible bool`, `CreatedAt time.Time`                                                                                                      |
| `Transaction` | `ID TransactionID`, `ProfileID`, `BalanceID`, `Amount Money`, `Fees Money`, `Total Money`, `RunningBalance Money`, `Exchange *TransactionExchange`, `Type TransactionType`, `Description`, `Reference`, `Category`, `MerchantName`, `Date time.Time` |
| `Money`       | `Cents int64`, `Currency Currency`                                                                                                                                                                                                                   |
| `Currency`    | `string` (ISO 4217, validated via `NewCurrency`)                                                                                                                                                                                                     |

**Raw types** (internal, mirror Wise JSON wire format) in `internal/raw/types.go`.

### Open Banking Data Models

**Account** (OBReadAccount6):

```json
{
	"AccountId": "504",
	"Currency": "GBP",
	"AccountType": "Personal",
	"AccountSubType": "EMoney",
	"Account": [
		{
			"SchemeName": "UK.OBIE.SortCodeAccountNumber",
			"Identification": "230xxx1000xxxx",
			"Name": "John Smith (GBP)"
		}
	],
	"Servicer": {
		"SchemeName": "UK.OBIE.BICFI",
		"Identification": "TRWIGB22"
	}
}
```

Some accounts have an empty `Account` array (no local bank details) — use `AccountId` for identification.

---

## 6. AISP Permissions

| Permission                | Description                  |
| ------------------------- | ---------------------------- |
| `ReadAccountsBasic`       | Basic account information    |
| `ReadAccountsDetail`      | Detailed account information |
| `ReadBalances`            | Account balances             |
| `ReadTransactionsBasic`   | Basic transaction data       |
| `ReadTransactionsCredits` | Credit transactions only     |
| `ReadTransactionsDebits`  | Debit transactions only      |
| `ReadTransactionsDetail`  | Detailed transaction data    |
| `ReadDirectDebits`        | Direct debit information     |

---

## 7. Payment Capabilities (PISP, Open Banking only)

| Feature              | Domestic Payments                               | International Payments                                      |
| -------------------- | ----------------------------------------------- | ----------------------------------------------------------- |
| **Currencies**       | Same-currency transfers                         | Cross-currency transfers                                    |
| **Amount modes**     | Fixed amount                                    | Fixed Source Amount OR Fixed Target Amount                  |
| **Exchange rate**    | N/A                                             | INDICATIVE at consent; real rate shown during authorization |
| **Cut-off time**     | 30 min after consent creation                   | —                                                           |
| **Account schemes**  | `UK.OBIE.SortCodeAccountNumber`, `UK.OBIE.IBAN` | Same                                                        |
| **Reference limits** | EUR: 35 chars, GBP: 18 chars                    | Same                                                        |

wise-go has **no payment capabilities** — read-only.

---

## 8. Architecture & Design

|                          | wise-go SDK                                                                               | Open Banking API (if a client were built)          |
| ------------------------ | ----------------------------------------------------------------------------------------- | -------------------------------------------------- |
| **Retry**                | `failsafe-go` (exponential backoff + jitter, 3 attempts, 100ms-5s)                        | Governed by OB UK standard (rate limit headers)    |
| **Retryable conditions** | HTTP 429, 5xx, network errors                                                             | OB UK standard error handling                      |
| **Errors**               | Typed (`RateLimitError`, `AuthError`, `NotFoundError`, `ServerError`) via go-error-family | OB standard error format (`OBErrorResponse1`)      |
| **Pagination**           | None (single response)                                                                    | OB UK standard (`page`, `page-size` params)        |
| **Sandbox**              | `WithSandbox()` option                                                                    | Separate sandbox base URL + sandbox certificates   |
| **Type safety**          | Branded IDs, typed enums, two-layer (raw -> public) types                                 | Would need to model OB UK schemas                  |
| **Package structure**    | Flat `package wise` (8 files, single package)                                             | Would likely need sub-packages for AISP/PISP/CBPII |
| **Dependencies**         | `failsafe-go`, `go-branded-id`, `go-error-family`                                         | Would need OAuth2/mTLS client + OB schema types    |
| **JSON**                 | `encoding/json/v2` (requires `GOEXPERIMENT=jsonv2`)                                       | Standard JSON per OB UK spec                       |

---

## 9. Contingency Mechanism

Wise offers a **Connected Applications API** as a secondary/backup providing the same functionality for AISPs and PISPs during major Open Banking API outages. Registration must be requested proactively via `openbanking@wise.com`.

---

## 10. Migration Notes (OB v3.1 to v3.1.11)

- Transaction IDs are **different** between v3.1 and v3.1.11 — reconcile via `supplementaryData` fields
- "Fee split" feature is **enabled by default** for all TPPs in v3.1.11 (was opt-in in v3.1)

---

## 11. Implications for bank-sync

| Question                                      | Answer                                                                                                                                                                                                            |
| --------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Should bank-sync switch to Open Banking?**  | **No.** bank-sync is a personal tool syncing your own accounts. The Wise Public API (via wise-go) is the right fit. Open Banking requires regulatory TPP status, certificate management, and a user consent flow. |
| **What would adopting Open Banking enable?**  | Payment initiation (send money programmatically), direct debit visibility, and multi-user consent-driven access — but with heavy regulatory overhead.                                                             |
| **Is the data the same?**                     | Largely yes for accounts/balances/transactions, but the JSON shapes are entirely different (OB UK vs. Wise proprietary). Fees are modeled differently (embedded vs. separate debit transactions).                 |
| **Could wise-go support Open Banking later?** | Possible but would be a **separate client** — different base URL, auth (mTLS + OAuth), data models (OB UK schemas), and consent lifecycle. Not an extension of the existing client.                               |
| **Transaction history limits**                | wise-go: unlimited date range. Open Banking: max 450 days per query, with consent-age restrictions (90-day lookback window for older consents). wise-go is better for deep historical sync.                       |

---

## 12. Key External References

| Resource                          | URL                                                                                                                                            |
| --------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Wise Open Banking guide           | https://docs.wise.com/guides/developer/open-banking                                                                                            |
| Wise Open Banking (markdown)      | https://docs.wise.com/guides/developer/open-banking.md                                                                                         |
| Wise Public API reference         | https://docs.wise.com/api-reference                                                                                                            |
| Production OIDC config            | https://wise.com/openbanking/.well-known/openid-configuration                                                                                  |
| OB UK Standard v3.1.11            | https://openbankinguk.github.io/read-write-api-site3/v3.1.11/profiles/read-write-data-api-profile.html                                         |
| OB Payment Message Formats        | https://openbankinguk.github.io/read-write-api-site3/v3.1.11/references/domestic-payment-message-formats.html                                  |
| OB Dynamic Client Registration    | https://openbankinguk.github.io/dcr-docs-pub/v3.3/dynamic-client-registration.html                                                             |
| OB Security Profile (Hybrid Flow) | https://openbanking.atlassian.net/wiki/spaces/DZ/pages/83919096/Open+Banking+Security+Profile+-+Implementer+s+Draft+v1.1.2                     |
| OB Directory (eIDAS certs)        | https://openbanking.atlassian.net/wiki/spaces/DZ/pages/1322979574/Open+Banking+Directory+Usage+-+eIDAS+release+Production+-+v1.9               |
| Wise Multi-Currency Account       | https://wise.com/multi-currency-account                                                                                                        |
| Wise Versioning Policy            | https://docs.wise.com/guides/developer/global-versioning                                                                                       |
| Q2 2026 Performance Report (PDF)  | https://docs.wise.com/assets/wise-open-banking-reporting-q2-2026.af3953c4ce1f26755d90add67bd0634e1539883b387400eccfe4c941d12fe704.6b9bfc58.pdf |
| Open Banking Directory            | https://www.openbanking.org.uk/providers/directory/                                                                                            |
| Wise Open Banking contact         | `openbanking@wise.com`                                                                                                                         |

---

## 13. Certificate Management (Open Banking only)

| Participant Type            | Certificate Update Method                                                                     |
| --------------------------- | --------------------------------------------------------------------------------------------- |
| OB Directory (OBWAC/OBSeal) | **Automatic** — added to JWKS endpoints, usable as soon as associated with Software Statement |
| OB Directory (eIDAS)        | Automatic if trust chain unchanged; **manual** if any cert in chain changed                   |
| Other participants          | **Manual only** — contact `openbanking@wise.com`                                              |

---

## Appendix A: wise-go SDK Quick Reference

**Module:** `github.com/larsartmann/wise-go` v0.5.0  
**Go:** 1.26.5 (requires `GOEXPERIMENT=jsonv2`)

### Client Methods

| Method             | Signature                                                                                           | Description                                         |
| ------------------ | --------------------------------------------------------------------------------------------------- | --------------------------------------------------- |
| `New`              | `New(apiKey string, opts ...Option) *Client`                                                        | Constructor with functional options                 |
| `Authenticate`     | `(c *Client) Authenticate(ctx) error`                                                               | Validates API key (calls `ListProfiles`)            |
| `Health`           | `(c *Client) Health(ctx) error`                                                                     | Health check (delegates to `Authenticate`)          |
| `ListProfiles`     | `(c *Client) ListProfiles(ctx) ([]Profile, error)`                                                  | List all profiles                                   |
| `ListBalances`     | `(c *Client) ListBalances(ctx, profileID ProfileID) ([]Balance, error)`                             | List visible, non-investment balances               |
| `GetBalance`       | `(c *Client) GetBalance(ctx, profileID, balanceID) (*Balance, error)`                               | Get specific balance (linear scan)                  |
| `ListTransactions` | `(c *Client) ListTransactions(ctx, req ListTransactionsRequest) (*ListTransactionsResponse, error)` | List transactions for a balance within a date range |

### Configuration Options

| Option                               | Effect                                              |
| ------------------------------------ | --------------------------------------------------- |
| `WithSandbox()`                      | Sandbox base URL                                    |
| `WithBaseURL(url)`                   | Custom base URL                                     |
| `WithTimeout(d)`                     | HTTP client timeout (default: 30s)                  |
| `WithRetry(max, minDelay, maxDelay)` | Retry policy (default: 3 retries, 100ms-5s backoff) |
| `WithHTTPClient(client)`             | Custom `Doer` for testing/middleware                |

### Typed Errors

| Type              | HTTP Status | Retryable    | ErrorFamily |
| ----------------- | ----------- | ------------ | ----------- |
| `APIError` (base) | any non-2xx | No (default) | `Rejection` |
| `RateLimitError`  | 429         | Yes          | `Transient` |
| `AuthError`       | 401, 403    | No           | `Rejection` |
| `NotFoundError`   | 404         | No           | `Rejection` |
| `ServerError`     | 5xx         | Yes          | `Transient` |

### Enums

| Enum              | Values                                                                              |
| ----------------- | ----------------------------------------------------------------------------------- |
| `ProfileType`     | `Personal`, `Business`                                                              |
| `BalanceType`     | `Standard`, `Savings`                                                               |
| `InvestmentState` | `NotInvested`, `Invested`                                                           |
| `TransactionType` | `Card`, `Credit`, `Debit`, `Exchange`, `Fee`, `Refund`, `Transfer`, `Payment`       |
| `DetailType`      | `CardPayment`, `CardRefund`, `Transfer`, `Payment`, `Conversion`, `Exchange`, `Fee` |

### Dependencies (3 production)

- `github.com/failsafe-go/failsafe-go v0.9.6` — retry with backoff
- `github.com/larsartmann/go-branded-id v0.5.1` — phantom-typed IDs
- `github.com/larsartmann/go-error-family v0.10.0` — error classification
