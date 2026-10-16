package wise

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// --- Structured error types ---

// Error code constants returned by ErrorCode() on each error type.
const (
	errorCodeAPI       = "wise.api_error"
	errorCodeRateLimit = "wise.rate_limit"
	errorCodeAuth      = "wise.auth"
	errorCodeNotFound  = "wise.not_found"
	errorCodeServer    = "wise.server"
	errorCodeSCA       = "wise.sca_challenge"
)

// Error codes carried by errorfamily-classified errors (validation
// rejections and corruption wraps). This block is the catalog: every code
// the SDK emits is named here exactly once, call sites reference the
// constant, and tests pin the wire value.
const (
	errorCodeAccountInvalidRequest             = "wise.account.invalid_request"
	errorCodeAccountParseCreationTime          = "wise.account.parse_creation_time"
	errorCodeAccountParseModificationTime      = "wise.account.parse_modification_time"
	errorCodeAccountDetailsInvalidRequest      = "wise.account_details.invalid_request"
	errorCodeAccountDetailsParseIdentification = "wise.account_details.parse_identification"
	errorCodeBalanceInvalidRequest             = "wise.balance.invalid_request"
	errorCodeBalanceParseCreationTime          = "wise.balance.parse_creation_time"
	errorCodeBalanceParseCurrency              = "wise.balance.parse_currency"
	errorCodeBalanceParseType                  = "wise.balance.parse_type"
	errorCodeDeliveryEstimateParseDate         = "wise.delivery_estimate.parse_date"
	errorCodeFundingParseStatus                = "wise.funding.parse_status"
	errorCodeFundingParseType                  = "wise.funding.parse_type"
	errorCodeMoneyParse                        = "wise.money.parse"
	errorCodeOTTChannelUnsupported             = "wise.ott.channel_unsupported"
	errorCodeOTTDecode                         = "wise.ott.decode"
	errorCodeOTTInvalidRequest                 = "wise.ott.invalid_request"
	errorCodeOTTUncleared                      = "wise.ott.uncleared"
	errorCodeProfileInvalidRequest             = "wise.profile.invalid_request"
	errorCodeProfileParseCreatedAt             = "wise.profile.parse_created_at"
	errorCodeProfileParseType                  = "wise.profile.parse_type"
	errorCodeQuoteInvalidRequest               = "wise.quote.invalid_request"
	errorCodeQuoteParseCreated                 = "wise.quote.parse_created"
	errorCodeQuoteParseExpiration              = "wise.quote.parse_expiration"
	errorCodeQuoteParsePaymentOptionDelivery   = "wise.quote.parse_payment_option_delivery"
	errorCodeQuoteParsePaymentOptionSource     = "wise.quote.parse_payment_option_source"
	errorCodeQuoteParsePaymentOptionTarget     = "wise.quote.parse_payment_option_target"
	errorCodeQuoteParseSourceCurrency          = "wise.quote.parse_source_currency"
	errorCodeQuoteParseTargetCurrency          = "wise.quote.parse_target_currency"
	errorCodeRatesEmptyResponse                = "wise.rates.empty_response"
	errorCodeRatesInvalidRequest               = "wise.rates.invalid_request"
	errorCodeRatesParseSource                  = "wise.rates.parse_source"
	errorCodeRatesParseTarget                  = "wise.rates.parse_target"
	errorCodeRatesParseTime                    = "wise.rates.parse_time"
	errorCodeRecipientInvalidRequest           = "wise.recipient.invalid_request"
	errorCodeRecipientMap                      = "wise.recipient.map"
	errorCodeRecipientParseCurrency            = "wise.recipient.parse_currency"
	errorCodeResponseDecode                    = "wise.response.decode"
	errorCodeTotalFundsInvalidRequest          = "wise.total_funds.invalid_request"
	errorCodeTransactionParseDate              = "wise.transaction.parse_date"
	errorCodeTransactionsInvalidRequest        = "wise.transactions.invalid_request"
	errorCodeTransferInvalidRequest            = "wise.transfer.invalid_request"
	errorCodeTransferMap                       = "wise.transfer.map"
	errorCodeTransferParseCreated              = "wise.transfer.parse_created"
	errorCodeTransferParseCurrency             = "wise.transfer.parse_currency"
	errorCodeTransferParsePaymentCreated       = "wise.transfer.parse_payment_created"
	errorCodeTransferParsePaymentCurrency      = "wise.transfer.parse_payment_currency"
	errorCodeTransferParsePaymentUpdated       = "wise.transfer.parse_payment_updated"
	errorCodeUserInvalidRequest                = "wise.user.invalid_request"
	errorCodeUserParseDateOfBirth              = "wise.user.parse_date_of_birth"
	errorCodeWebhookDecode                     = "wise.webhook.decode"
	errorCodeWebhookInvalidRequest             = "wise.webhook.invalid_request"
)

// Sentinel errors name distinct response-shape failures so consumers can
// match them with errors.Is. The SDK returns them only wrapped (with a
// Corruption classification and a code from the catalog above), never bare,
// and they are declared with the error interface type on purpose: they are
// matched by value, never by type.
var (
	// ErrEmptyRateList reports a rates payload that decoded to zero entries.
	ErrEmptyRateList error = errors.New("empty rate list")

	// ErrAccountIdentificationCardinality marks a wire accountIdentification
	// that carries zero or several oneOf members; the wrap's detail message
	// holds the count.
	ErrAccountIdentificationCardinality error = errors.New(
		"accountIdentification member count is not exactly one",
	)

	// ErrOTTAlternativesShape is the sentinel for alternatives payloads that
	// are neither the documented object form nor a bare string array.
	ErrOTTAlternativesShape error = errors.New("alternatives are neither objects nor strings")

	// ErrTransferRefundAmountMissing marks a transfers#refund payload whose
	// resource block lacks the refund amount/currency pair.
	ErrTransferRefundAmountMissing error = errors.New(
		"refund amount or currency missing from resource block",
	)
)

// Curated wave-1 code sentinels: the Wise wire codes with a consumer
// branching story, matchable through any wrap chain via errors.Is on an
// *APIError (APIError.Is compares the structured Details codes, never
// message text). Declared with the error interface type like every sentinel
// in this file: they are matched by value, never by type. Codes without a
// sentinel (open-enum passthrough) stay reachable via [APIError.HasCode].
var (
	// ErrNotValid reports Wise rejecting a value as not valid (wire code
	// "not.valid"), typically a malformed amount or field value.
	ErrNotValid error = errors.New("wise: not valid")

	// ErrInvalidValue reports Wise rejecting a field value (wire code
	// "invalid_value").
	ErrInvalidValue error = errors.New("wise: invalid value")

	// ErrParameterMissing reports a required request parameter Wise did not
	// receive (wire code "parameter_missing").
	ErrParameterMissing error = errors.New("wise: parameter missing")

	// ErrParameterInvalid reports a request parameter whose value Wise
	// rejected (wire code "parameter_invalid").
	ErrParameterInvalid error = errors.New("wise: parameter invalid")

	// ErrInvalidRequest reports a generic malformed request (wire code
	// "invalid_request").
	ErrInvalidRequest error = errors.New("wise: invalid request")

	// ErrUnauthorized reports missing or invalid credentials (wire code
	// "unauthorized").
	ErrUnauthorized error = errors.New("wise: unauthorized")

	// ErrForbidden reports credentials without permission for the operation
	// (wire code "forbidden").
	ErrForbidden error = errors.New("wise: forbidden")

	// ErrResourceNotFound reports that the addressed resource does not exist
	// (wire code "resource_not_found").
	ErrResourceNotFound error = errors.New("wise: resource not found")

	// ErrOperationNotSupported reports that the API surface does not support
	// the requested operation (wire code "operation_not_supported").
	ErrOperationNotSupported error = errors.New("wise: operation not supported")

	// ErrRateLimitExceeded reports that Wise throttled the request (wire
	// code "rate_limit_exceeded"); back off and retry.
	ErrRateLimitExceeded error = errors.New("wise: rate limit exceeded")

	// ErrSystemError reports a Wise-side system failure (wire code
	// "system_error"), typically transient.
	ErrSystemError error = errors.New("wise: system error")
)

// codeSentinels maps the curated code sentinels to the Wise wire code each
// stands for; [APIError.Is] consults it. Keyed by sentinel identity (each is
// a distinct errors.New value), so lookup is pointer comparison and a
// look-alike error never matches by accident.
//
//nolint:gochecknoglobals // sentinel-code lookup table built from the errors.New sentinels above
var codeSentinels = map[error]WiseErrorCode{
	ErrNotValid:              WiseErrorCodeNotValid,
	ErrInvalidValue:          WiseErrorCodeInvalidValue,
	ErrParameterMissing:      WiseErrorCodeParameterMissing,
	ErrParameterInvalid:      WiseErrorCodeParameterInvalid,
	ErrInvalidRequest:        WiseErrorCodeInvalidRequest,
	ErrUnauthorized:          WiseErrorCodeUnauthorized,
	ErrForbidden:             WiseErrorCodeForbidden,
	ErrResourceNotFound:      WiseErrorCodeResourceNotFound,
	ErrOperationNotSupported: WiseErrorCodeOperationNotSupported,
	ErrRateLimitExceeded:     WiseErrorCodeRateLimitExceeded,
	ErrSystemError:           WiseErrorCodeSystemError,
}

// WiseErrorCode is Wise's own granular, machine-readable error code from the
// errors[].code field of a non-2xx response body (for example
// resource_not_found on a 404 or rate_limit_exceeded on a 429). It is an open
// enum: Wise documents a growing set of codes and may add new ones at any
// time, so values received from the wire pass through unmodified; switch on
// the constants for known cases and handle the default explicitly. Read the
// codes structurally from [APIError.Details], match the curated ones with
// errors.Is against the Err* code sentinels, or branch on any code (curated
// or not) with [APIError.HasCode].
type WiseErrorCode string

// Granular Wise error codes from the documented API surface (census against
// the vendored OpenAPI snapshot, 2026-10-16). This list is deliberately NOT
// exhaustive: it names the codes with a consumer branching story. Unknown
// wire values are valid WiseErrorCodes.
const (
	WiseErrorCodeNotValid              WiseErrorCode = "not.valid"
	WiseErrorCodeInvalidValue          WiseErrorCode = "invalid_value"
	WiseErrorCodeParameterMissing      WiseErrorCode = "parameter_missing"
	WiseErrorCodeParameterInvalid      WiseErrorCode = "parameter_invalid"
	WiseErrorCodeMissingValue          WiseErrorCode = "missing_value"
	WiseErrorCodeInvalidRequest        WiseErrorCode = "invalid_request"
	WiseErrorCodeUnauthorized          WiseErrorCode = "unauthorized"
	WiseErrorCodeForbidden             WiseErrorCode = "forbidden"
	WiseErrorCodeResourceNotFound      WiseErrorCode = "resource_not_found"
	WiseErrorCodeOperationNotSupported WiseErrorCode = "operation_not_supported"
	WiseErrorCodeRateLimitExceeded     WiseErrorCode = "rate_limit_exceeded"
	WiseErrorCodeSystemError           WiseErrorCode = "system_error"
	WiseErrorCodeInvalidFileExtension  WiseErrorCode = "invalid_file_extension"
	WiseErrorCodeUnexpectedError       WiseErrorCode = "unexpected.error"
)

// Wise's Strong Customer Authentication (SCA) response headers. Wise answers
// SCA-protected endpoints (balance statements for UK/EEA profiles among them)
// with HTTP 403 and an EMPTY body: the verdict lives in
// x-2fa-approval-result and the one-time token (OTT) to clear the challenge
// in x-2fa-approval. Retrying the same request with that token in the
// x-2fa-approval header completes the flow once the user approved the
// challenge (e.g. in the Wise app). Values are in canonical MIME form (what
// net/http's Header.Get/Set use as map keys; on the wire HTTP header names
// are case-insensitive).
const (
	HeaderTwoFAApproval       = "X-2fa-Approval"
	HeaderTwoFAApprovalResult = "X-2fa-Approval-Result"
)

// WiseErrorDetail is one entry of Wise's structured error envelope: the
// granular wire code plus Wise's own human-readable message for that entry.
// A single response can carry several entries (the envelope is an array).
type WiseErrorDetail struct {
	Code    WiseErrorCode
	Message string
}

// APIError is the base error returned for non-2xx Wise API responses.
type APIError struct {
	StatusCode int
	Message    string
	Body       string
	// Headers carries the response headers when available. Wise puts
	// diagnostics that never appear in the body here (SCA challenge tokens,
	// rate-limit owner), so dropping them hides the actual verdict.
	Headers http.Header
	// Details carries Wise's structured error envelope (errors[].code /
	// errors[].message from the response body) so consumers can branch on
	// Wise's granular codes without parsing Message. Empty when the body was
	// not Wise's structured envelope (empty body, plain text, non-JSON).
	// Message stays the flattened join of the same entries for backward
	// compatibility; prefer Details for matching.
	Details []WiseErrorDetail
}

func (e *APIError) Error() string {
	return fmt.Sprintf("wise: api error (%d): %s", e.StatusCode, e.Message)
}

func (e *APIError) ErrorCode() string {
	return errorCodeAPI
}

func (e *APIError) ErrorFamily() errorfamily.Family {
	return errorfamily.Rejection
}

func (e *APIError) ErrorContext() map[string]string {
	return map[string]string{
		"status_code": strconv.Itoa(e.StatusCode),
	}
}

// HasCode reports whether any structured error detail carries the given Wise
// wire code. This is the matcher for codes without a curated sentinel (and a
// convenience for the curated ones); unknown codes work because WiseErrorCode
// is an open enum.
func (e *APIError) HasCode(code WiseErrorCode) bool {
	if e == nil {
		return false
	}

	for _, detail := range e.Details {
		if detail.Code == code {
			return true
		}
	}

	return false
}

// Is reports whether the error carries the Wise wire code represented by a
// curated code sentinel (see codeSentinels), so errors.Is matches through
// arbitrary fmt.Errorf("%w") wrap chains and through every typed subtype
// that embeds APIError. It compares structured Details codes only (never
// message text) and deliberately returns false for everything else: type
// extraction stays the job of errors.AsType.
func (e *APIError) Is(target error) bool {
	if e == nil {
		return false
	}

	want, ok := codeSentinels[target]
	if !ok {
		return false
	}

	return e.HasCode(want)
}

// RateLimitError is returned when the Wise API returns HTTP 429.
type RateLimitError struct {
	APIError

	RetryAfter    time.Duration
	RateLimitedBy string
}

func (e *RateLimitError) ErrorFamily() errorfamily.Family {
	return errorfamily.Transient
}

func (e *RateLimitError) IsRetryable() bool {
	return true
}

func (e *RateLimitError) ErrorContext() map[string]string {
	ctx := map[string]string{
		"status_code": strconv.Itoa(e.StatusCode),
		"retry_after": e.RetryAfter.String(),
	}
	if e.RateLimitedBy != "" {
		ctx["rate_limited_by"] = e.RateLimitedBy
	}

	return ctx
}

func (e *RateLimitError) ErrorCode() string {
	return errorCodeRateLimit
}

// AuthError is returned when the Wise API returns HTTP 401 or 403.
type AuthError struct {
	APIError
}

func (e *AuthError) ErrorCode() string {
	return errorCodeAuth
}

// SCAChallengeError is returned when the Wise API answers 403 with Wise's
// two-factor headers instead of a body: Strong Customer Authentication is
// required for the request. This is not a token permission problem — the same
// token works on non-SCA endpoints. Clear the challenge by verifying it over
// the one-time-token endpoints (see [Client.ClearSCAChallenge]) or by viewing
// a statement on wise.com, then retry with the one-time token in
// [HeaderTwoFAApproval]. The token value deliberately never appears in the
// error message — fetch it with [SCAChallengeError.TwoFAApprovalToken].
type SCAChallengeError struct {
	APIError
}

func (e *SCAChallengeError) ErrorCode() string {
	return errorCodeSCA
}

func (e *SCAChallengeError) ErrorContext() map[string]string {
	return map[string]string{
		"status_code":           strconv.Itoa(e.StatusCode),
		"approval_result":       e.Headers.Get(HeaderTwoFAApprovalResult),
		"approval_token_issued": strconv.FormatBool(e.TwoFAApprovalToken() != ""),
	}
}

func (e *SCAChallengeError) Error() string {
	return fmt.Sprintf(
		"wise: sca challenge (%d): strong customer authentication required "+
			"(x-2fa-approval-result=%q, one-time token issued: %t); clear the challenge "+
			"(verify it via the one-time-token endpoints or view a statement on wise.com), "+
			"then retry with the token from TwoFAApprovalToken() in the x-2fa-approval header",
		e.StatusCode,
		e.Headers.Get(HeaderTwoFAApprovalResult),
		e.TwoFAApprovalToken() != "",
	)
}

// TwoFAApprovalToken returns the one-time token (OTT) this challenge was
// issued with, for the retry request's x-2fa-approval header. The zero
// value means Wise issued a tokenless (PSD2 flavor) challenge — nothing to
// clear programmatically, but the request is still SCA-blocked.
func (e *SCAChallengeError) TwoFAApprovalToken() OneTimeToken {
	return OneTimeToken(e.Headers.Get(HeaderTwoFAApproval))
}

// NotFoundError is returned when the Wise API returns HTTP 404.
type NotFoundError struct {
	APIError
}

func (e *NotFoundError) ErrorCode() string {
	return errorCodeNotFound
}

// ServerError is returned when the Wise API returns HTTP 5xx.
type ServerError struct {
	APIError
}

func (e *ServerError) ErrorFamily() errorfamily.Family {
	return errorfamily.Transient
}

func (e *ServerError) IsRetryable() bool {
	return true
}

func (e *ServerError) ErrorCode() string {
	return errorCodeServer
}

func newAPIError(
	statusCode int,
	body string,
	headers http.Header,
	retryAfter time.Duration,
	rateLimitedBy string,
) error {
	errResp := parseErrorResponse(body)

	msg := body

	var details []WiseErrorDetail

	if errResp != nil && len(errResp.Errors) > 0 {
		msgs := make([]string, 0, len(errResp.Errors))
		details = make([]WiseErrorDetail, 0, len(errResp.Errors))

		for _, e := range errResp.Errors {
			msgs = append(msgs, fmt.Sprintf("%s: %s", e.Code, e.Message))
			details = append(details, WiseErrorDetail{Code: WiseErrorCode(e.Code), Message: e.Message})
		}

		msg = strings.Join(msgs, "; ")
	}

	if headers == nil {
		headers = http.Header{}
	}

	base := APIError{
		StatusCode: statusCode,
		Message:    msg,
		Body:       body,
		Headers:    headers,
		Details:    details,
	}

	switch {
	case statusCode == http.StatusTooManyRequests:
		return &RateLimitError{
			APIError:      base,
			RetryAfter:    retryAfter,
			RateLimitedBy: rateLimitedBy,
		}
	case isSCAChallenge(statusCode, headers):
		return &SCAChallengeError{APIError: base}
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return &AuthError{APIError: base}
	case statusCode == http.StatusNotFound:
		return &NotFoundError{APIError: base}
	case statusCode >= http.StatusInternalServerError:
		return &ServerError{APIError: base}
	default:
		return &base
	}
}

// isSCAChallenge detects Wise's header-only SCA rejection: a 403 whose body
// is empty (or unhelpful) while the two-factor headers carry the verdict.
// Wise does not consistently send both headers, so either one triggers.
func isSCAChallenge(statusCode int, headers http.Header) bool {
	if statusCode != http.StatusForbidden {
		return false
	}

	return headers.Get(HeaderTwoFAApprovalResult) != "" || headers.Get(HeaderTwoFAApproval) != ""
}

func parseErrorResponse(body string) *raw.ErrorResponse {
	var errResp raw.ErrorResponse
	if json.Unmarshal([]byte(body), &errResp) == nil && len(errResp.Errors) > 0 {
		return &errResp
	}

	return nil
}
