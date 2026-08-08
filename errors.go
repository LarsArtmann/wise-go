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
)

// APIError is the base error returned for non-2xx Wise API responses.
type APIError struct {
	StatusCode int
	Message    string
	Body       string
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

func newAPIError(statusCode int, body string, retryAfter time.Duration, rateLimitedBy string) error {
	errResp := parseErrorResponse(body)

	msg := body

	if errResp != nil && len(errResp.Errors) > 0 {
		msgs := make([]string, 0, len(errResp.Errors))
		for _, e := range errResp.Errors {
			msgs = append(msgs, fmt.Sprintf("%s: %s", e.Code, e.Message))
		}

		msg = strings.Join(msgs, "; ")
	}

	base := APIError{
		StatusCode: statusCode,
		Message:    msg,
		Body:       body,
	}

	switch {
	case statusCode == http.StatusTooManyRequests:
		return &RateLimitError{
			APIError:      base,
			RetryAfter:    retryAfter,
			RateLimitedBy: rateLimitedBy,
		}
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

func parseErrorResponse(body string) *raw.ErrorResponse {
	var errResp raw.ErrorResponse
	if json.Unmarshal([]byte(body), &errResp) == nil && len(errResp.Errors) > 0 {
		return &errResp
	}

	return nil
}
