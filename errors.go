package wise

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

// --- Structured error types ---

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
	return "wise.api_error"
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
	RetryAfter time.Duration
}

func (e *RateLimitError) ErrorFamily() errorfamily.Family {
	return errorfamily.Transient
}

func (e *RateLimitError) IsRetryable() bool {
	return true
}

func (e *RateLimitError) ErrorContext() map[string]string {
	return map[string]string{
		"status_code": strconv.Itoa(e.StatusCode),
		"retry_after": e.RetryAfter.String(),
	}
}

// AuthError is returned when the Wise API returns HTTP 401 or 403.
type AuthError struct {
	APIError
}

func (e *AuthError) ErrorCode() string {
	return "wise.auth"
}

// NotFoundError is returned when the Wise API returns HTTP 404.
type NotFoundError struct {
	APIError
}

func (e *NotFoundError) ErrorCode() string {
	return "wise.not_found"
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
	return "wise.server"
}

func newAPIError(statusCode int, body string) error {
	errResp := parseErrorResponse(body)

	msg := body
	if errResp != nil && len(errResp.Errors) > 0 {
		msgs := make([]string, len(errResp.Errors))
		for i, e := range errResp.Errors {
			msgs[i] = fmt.Sprintf("%s: %s", e.Code, e.Message)
		}

		msg = joinStrings(msgs, "; ")
	}

	base := APIError{
		StatusCode: statusCode,
		Message:    msg,
		Body:       body,
	}

	switch {
	case statusCode == http.StatusTooManyRequests:
		return &RateLimitError{
			APIError:   base,
			RetryAfter: time.Second,
		}
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return &AuthError{APIError: base}
	case statusCode == http.StatusNotFound:
		return &NotFoundError{APIError: base}
	case statusCode >= 500:
		return &ServerError{APIError: base}
	default:
		return &base
	}
}

func parseErrorResponse(body string) *ErrorResponse {
	var errResp ErrorResponse
	if jsonUnmarshal([]byte(body), &errResp) == nil && len(errResp.Errors) > 0 {
		return &errResp
	}

	return nil
}

func joinStrings(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(ss[0])
	for _, s := range ss[1:] {
		result.WriteString(sep + s)
	}

	return result.String()
}
