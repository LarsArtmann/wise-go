package wise

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

const defaultRetryAfter = time.Second

// fetchByID validates a non-zero ID and performs a GET for a single resource.
// It centralises the boilerplate shared by GetProfile, GetTransfer, and future
// get-by-id endpoints so the public methods stay short and resource-specific.
func fetchByID[T any](
	ctx context.Context,
	client *Client,
	id int64,
	resource string,
	path string,
	target *T,
) error {
	if id == 0 {
		return errorfamily.NewRejection(
			"wise."+resource+".invalid_request",
			resource+"ID is required",
		)
	}

	if err := client.get(ctx, path, target); err != nil {
		return fmt.Errorf("get %s %d: %w", resource, id, err)
	}

	return nil
}

// toMoney converts a raw BalanceAmount to a validated Money value.
// A malformed amount is permanent response corruption, never a retryable
// failure, so the error is classified as Corruption.
func toMoney(a raw.BalanceAmount) (Money, error) {
	currency, err := NewCurrency(a.Currency)
	if err != nil {
		return Money{}, errorfamily.WrapCorruption(err, "wise.money.parse", fmt.Sprintf("currency %q", a.Currency))
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

const cErrTransferIDRequired = "transferID is required"

func errTransferIDRequired() error {
	return errorfamily.NewRejection(
		"wise.transfer.invalid_request",
		cErrTransferIDRequired,
	)
}

// parseWiseTimestamp parses a Wise API timestamp. Wise is inconsistent about
// separators and zone designators across endpoints: some emit full RFC3339
// ("2020-05-27T10:27:22Z"), others omit the zone ("2020-05-27T10:27:22") or
// use a space separator ("2020-05-27 10:27:22"). Zoneless values are
// interpreted as UTC. Callers comparing the resulting time.Time to a
// local-time value must convert explicitly to avoid silent off-by-one-day
// errors at boundaries.
func parseWiseTimestamp(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		// Wise's delivery-estimate timestamps carry milliseconds and a
		// numeric zone ("2018-01-10T12:15:00.000+0000"), which RFC3339
		// without fractional-second tolerance rejects.
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000Z0700",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	var errs []error

	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}

		errs = append(errs, err)
	}

	return time.Time{}, errors.Join(errs...)
}

// formatWiseTimestamp renders a time.Time as a Wise query-parameter value.
// Wise rejects zone offsets other than "Z" with HTTP 422
// ("wrong.date.format"), so outgoing timestamps MUST be normalized to UTC
// regardless of the caller's zone — mirror of parseWiseTimestamp, which
// tolerates Wise's loose input formats.
func formatWiseTimestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}
