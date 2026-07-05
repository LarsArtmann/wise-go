package wise

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
)

const defaultRetryAfter = time.Second

func jsonDecode(resp *http.Response, target any) error {
	return json.NewDecoder(resp.Body).Decode(target)
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
