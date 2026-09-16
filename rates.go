package wise

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/url"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// GetExchangeRate returns the current or historical exchange rate between two
// currencies. If at is the zero time, the current rate is returned; otherwise
// the rate at the requested time is returned (UTC is recommended).
func (c *Client) GetExchangeRate(
	ctx context.Context,
	source, target Currency,
	atTime time.Time,
) (*ExchangeRate, error) {
	if source == "" {
		return nil, errorfamily.NewRejection(
			"wise.rates.invalid_request",
			"source currency is required",
		)
	}

	if target == "" {
		return nil, errorfamily.NewRejection(
			"wise.rates.invalid_request",
			"target currency is required",
		)
	}

	if source == target {
		return nil, errorfamily.NewRejection(
			"wise.rates.invalid_request",
			"source and target currency must be different",
		)
	}

	query := func() string {
		v := url.Values{}
		v.Set("source", string(source))
		v.Set("target", string(target))

		if !atTime.IsZero() {
			v.Set("time", formatWiseTimestamp(atTime))
		}

		return v.Encode()
	}

	// The spec contract is an array of rate objects; capture the raw body so
	// the legacy single-object shape some Wise surfaces served stays
	// decodable without a second request (same tolerance doctrine as
	// parseWiseTimestamp).
	var payload jsontext.Value

	err := c.getWithQuery(ctx, "/v1/rates", query, &payload)
	if err != nil {
		return nil, fmt.Errorf("get exchange rate %s-%s: %w", source, target, err)
	}

	var rates []raw.ExchangeRate
	if unmarshalErr := json.Unmarshal(payload, &rates); unmarshalErr != nil {
		var single raw.ExchangeRate
		if singleErr := json.Unmarshal(payload, &single); singleErr != nil {
			return nil, fmt.Errorf("decode exchange rate %s-%s: %w", source, target, unmarshalErr)
		}

		rates = []raw.ExchangeRate{single}
	}

	if len(rates) == 0 {
		return nil, errorfamily.WrapCorruption(
			errors.New("empty rate list"),
			"wise.rates.empty_response",
			fmt.Sprintf("no rates for %s-%s", source, target),
		)
	}

	chosen := rates[0]
	for _, rate := range rates {
		if rate.Source == string(source) && rate.Target == string(target) {
			chosen = rate

			break
		}
	}

	result, mapErr := mapExchangeRate(chosen)
	if mapErr != nil {
		return nil, fmt.Errorf("map exchange rate %s-%s: %w", source, target, mapErr)
	}

	return result, nil
}

func mapExchangeRate(r raw.ExchangeRate) (*ExchangeRate, error) {
	atTime, err := parseWiseTimestamp(r.Time)
	if err != nil {
		return nil, errorfamily.WrapCorruption(
			err,
			"wise.rates.parse_time",
			fmt.Sprintf("parse time %q", r.Time),
		)
	}

	source, err := NewCurrency(r.Source)
	if err != nil {
		return nil, errorfamily.WrapCorruption(
			err,
			"wise.rates.parse_source",
			fmt.Sprintf("source currency %q", r.Source),
		)
	}

	target, err := NewCurrency(r.Target)
	if err != nil {
		return nil, errorfamily.WrapCorruption(
			err,
			"wise.rates.parse_target",
			fmt.Sprintf("target currency %q", r.Target),
		)
	}

	return &ExchangeRate{
		Source: source,
		Target: target,
		Rate:   r.Rate,
		Time:   atTime,
	}, nil
}
