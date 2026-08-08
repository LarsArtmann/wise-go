package wise

import (
	"context"
	"fmt"

	id "github.com/larsartmann/go-branded-id"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// ListBalances returns all balances for a profile.
//
// Only visible, non-investment balances are returned. Wise exposes no
// per-balance endpoint, so there is no way to fetch a hidden or invested
// balance individually through this SDK.
func (c *Client) ListBalances(ctx context.Context, profileID ProfileID) ([]BalanceResult, error) {
	path := fmt.Sprintf("/v4/profiles/%d/balances", profileID.Get())

	var balances []raw.Balance

	err := c.get(ctx, path, &balances)
	if err != nil {
		return nil, fmt.Errorf("list balances for profile %d: %w", profileID.Get(), err)
	}

	results := make([]BalanceResult, 0, len(balances))
	for _, b := range balances {
		if !b.Visible || InvestmentState(b.InvestmentState) != InvestmentStateNotInvested {
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
func (c *Client) GetBalance(
	ctx context.Context,
	profileID ProfileID,
	balanceID BalanceID,
) (*BalanceResult, error) {
	balances, err := c.ListBalances(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("get balance %d for profile %d: %w", balanceID.Get(), profileID.Get(), err)
	}

	for _, balance := range balances {
		if balance.ID == balanceID {
			return &balance, nil
		}
	}

	return nil, errorfamily.NewRejection(
		"wise.balance.not_found",
		fmt.Sprintf("balance %d not found for profile %d", balanceID.Get(), profileID.Get()),
	)
}

func mapBalance(b raw.Balance) (BalanceResult, error) {
	createdAt, err := parseRFC3339(b.CreationTime)
	if err != nil {
		return BalanceResult{}, fmt.Errorf("parse creation_time %q: %w", b.CreationTime, err)
	}

	balanceType, err := parseBalanceType(b.Type)
	if err != nil {
		return BalanceResult{}, fmt.Errorf("parse type %q: %w", b.Type, err)
	}

	currency, err := NewCurrency(b.Currency)
	if err != nil {
		return BalanceResult{}, fmt.Errorf("currency %q: %w", b.Currency, err)
	}

	amount, err := toMoney(b.Amount)
	if err != nil {
		return BalanceResult{}, fmt.Errorf("amount: %w", err)
	}

	reserved, err := toMoney(b.ReservedAmount)
	if err != nil {
		return BalanceResult{}, fmt.Errorf("reserved amount: %w", err)
	}

	return BalanceResult{
		ID:        id.NewID[BalanceBrand](b.ID),
		Currency:  currency,
		Type:      balanceType,
		Name:      b.Name,
		Amount:    amount,
		Reserved:  reserved,
		Visible:   b.Visible,
		CreatedAt: createdAt,
	}, nil
}

func parseBalanceType(s string) (BalanceType, error) {
	return parseEnum(map[string]BalanceType{
		"STANDARD": BalanceTypeStandard,
		"SAVINGS":  BalanceTypeSavings,
	}, s, "balance type")
}
