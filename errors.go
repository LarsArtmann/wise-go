package wise

import (
	"encoding/json/v2"
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

// APIError is the base error returned for non-2xx Wise API responses.
type APIError struct {
	StatusCode int
	Message    string
	Body       string
	// Headers carries the response headers when available. Wise puts
	// diagnostics that never appear in the body here (SCA challenge tokens,
	// rate-limit owner), so dropping them hides the actual verdict.
	Headers http.Header
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
// token works on non-SCA endpoints. Clear it by approving the challenge (Wise
// app / web) and retrying with the one-time token in [HeaderTwoFAApproval].
type SCAChallengeError struct {
	APIError
}

func (e *SCAChallengeError) ErrorCode() string {
	return errorCodeSCA
}

func (e *SCAChallengeError) Error() string {
	return fmt.Sprintf(
		"wise: sca challenge (%d): strong customer authentication required "+
			"(x-2fa-approval-result=%q, x-2fa-approval=%q); approve the challenge "+
			"in the Wise app, then retry with the one-time token in the x-2fa-approval header",
		e.StatusCode,
		e.Headers.Get(HeaderTwoFAApprovalResult),
		e.Headers.Get(HeaderTwoFAApproval),
	)
}

// TwoFAApprovalToken returns the one-time token (OTT) this challenge was
// issued with, for the retry request's x-2fa-approval header.
func (e *SCAChallengeError) TwoFAApprovalToken() string {
	return e.Headers.Get(HeaderTwoFAApproval)
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

	if errResp != nil && len(errResp.Errors) > 0 {
		msgs := make([]string, 0, len(errResp.Errors))
		for _, e := range errResp.Errors {
			msgs = append(msgs, fmt.Sprintf("%s: %s", e.Code, e.Message))
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
