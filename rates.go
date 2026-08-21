package wise

import (
	"context"
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

	var rate raw.ExchangeRate

	err := c.getWithQuery(ctx, "/v1/rates", query, &rate)
	if err != nil {
		return nil, fmt.Errorf("get exchange rate %s-%s: %w", source, target, err)
	}

	result, mapErr := mapExchangeRate(rate)
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
