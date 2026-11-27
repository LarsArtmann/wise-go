package wise

import (
	"context"

	"github.com/larsartmann/wise-go/internal/raw"
	"github.com/samber/oops"
)

// CurrencyInfo describes one currency Wise allows for transfers — reference
// data for populating currency pickers.
type CurrencyInfo struct {
	Code             Currency
	Symbol           string
	Name             string
	CountryKeywords  []string
	SupportsDecimals bool
}

// ListCurrencies returns every currency Wise allows for transfers
// (GET /v1/currencies). The endpoint is public and needs no authentication.
func (c *Client) ListCurrencies(ctx context.Context) ([]CurrencyInfo, error) {
	var currencies []raw.CurrencyInfo

	if err := c.get(ctx, "/v1/currencies", &currencies); err != nil {
		return nil, oops.Wrapf(err, "list currencies")
	}

	result := make([]CurrencyInfo, 0, len(currencies))
	for _, currency := range currencies {
		result = append(result, CurrencyInfo{
			Code:             Currency(currency.Code),
			Symbol:           currency.Symbol,
			Name:             currency.Name,
			CountryKeywords:  currency.CountryKeywords,
			SupportsDecimals: currency.SupportsDecimals,
		})
	}

	return result, nil
}
