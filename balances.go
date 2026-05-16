package wise

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
)

// ListBalances returns all balances for a profile.
// Returns only visible, non-investment balances by default.
// Use WithHiddenBalances() option to include all balances.
func (c *Client) ListBalances(ctx context.Context, profileID int64) ([]BalanceResult, error) {
	path := fmt.Sprintf("/v4/profiles/%d/balances", profileID)

	var balances []Balance

	err := c.get(ctx, path, &balances)
	if err != nil {
		return nil, fmt.Errorf("list balances for profile %d: %w", profileID, err)
	}

	results := make([]BalanceResult, 0, len(balances))
	for _, b := range balances {
		if !b.Visible || b.InvestmentState != "NOT_INVESTED" {
			continue
		}

		result, mapErr := mapBalance(b)
		if mapErr != nil {
			return nil, fmt.Errorf("map balance %d: %w", b.ID, mapErr)
		}

		results = append(results, result)
	}

	return results, nil
}

// GetBalance returns a specific balance by ID within a profile.
func (c *Client) GetBalance(ctx context.Context, profileID int64, balanceID int64) (*BalanceResult, error) {
	balances, err := c.ListBalances(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("get balance %d for profile %d: %w", balanceID, profileID, err)
	}

	for _, balance := range balances {
		if balance.ID == balanceID {
			return &balance, nil
		}
	}

	return nil, errors.Newf("balance %d not found for profile %d", balanceID, profileID)
}

func mapBalance(b Balance) (BalanceResult, error) {
	createdAt, err := parseRFC3339(b.CreationTime)
	if err != nil {
		return BalanceResult{}, fmt.Errorf("parse creation_time %q: %w", b.CreationTime, err)
	}

	balanceType, err := parseBalanceType(b.Type)
	if err != nil {
		return BalanceResult{}, fmt.Errorf("parse type %q: %w", b.Type, err)
	}

	return BalanceResult{
		ID:               b.ID,
		Currency:         b.Currency,
		Type:             balanceType,
		Name:             b.Name,
		AmountCents:      b.Amount.Cents(),
		AmountCurrency:   b.Amount.Currency,
		ReservedCents:    b.ReservedAmount.Cents(),
		ReservedCurrency: b.ReservedAmount.Currency,
		Visible:          b.Visible,
		CreatedAt:        createdAt,
	}, nil
}

func parseBalanceType(s string) (BalanceType, error) {
	switch s {
	case "STANDARD":
		return BalanceTypeStandard, nil
	case "SAVINGS":
		return BalanceTypeSavings, nil
	default:
		return "", fmt.Errorf("unknown balance type %q", s)
	}
}
