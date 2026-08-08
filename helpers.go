package wise

import (
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/larsartmann/wise-go/internal/raw"
)

const defaultRetryAfter = time.Second

// toMoney converts a raw BalanceAmount to a validated Money value.
func toMoney(a raw.BalanceAmount) (Money, error) {
	currency, err := NewCurrency(a.Currency)
	if err != nil {
		return Money{}, fmt.Errorf("currency %q: %w", a.Currency, err)
	}

	return Money{Cents: a.Cents(), Currency: currency}, nil
}

// parseEnum maps a raw string to a typed enum value via a lookup table.
// Eliminates the duplicated switch-parser pattern across parseProfileType,
// parseBalanceType, and future typed enums.
func parseEnum[T ~string](table map[string]T, raw string, kind string) (T, error) {
	if v, ok := table[raw]; ok {
		return v, nil
	}

	var zero T

	//nolint:err113 // generic enum parser needs a dynamic error message
	return zero, fmt.Errorf("unknown %s %q", kind, raw)
}

func jsonDecode(resp *http.Response, target any) error {
	return json.UnmarshalRead(resp.Body, target)
}

func readBody(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)

	return string(body), err
}

// parseRetryAfter parses an HTTP Retry-After header value into a duration.
// Accepts both delta-seconds ("120") and HTTP-date (RFC1123) forms.
// Falls back to defaultRetryAfter when missing or unparseable.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return defaultRetryAfter
	}

	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}

	if t, err := time.Parse(time.RFC1123, value); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}

	return defaultRetryAfter
}

// parseWiseDate parses a Wise statement date string of the form "2006-01-02 15:04:05".
// Wise does not transmit a timezone; time.Parse interprets these values as UTC.
// Callers comparing the resulting time.Time to a local-time value must convert
// explicitly to avoid silent off-by-one-day errors at boundaries.
func parseWiseDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func parseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
