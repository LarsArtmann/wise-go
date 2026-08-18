package wise

import (
	"context"
	"fmt"

	id "github.com/larsartmann/go-branded-id"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// balanceTypesQuery lists every balance type the SDK can map (keep in sync
// with the parseBalanceType table). The v4 balances endpoint rejects requests
// without a types filter with "query.types: NotNull", so ListBalances always
// sends the full set and keeps visibility/investment filtering client-side.
const balanceTypesQuery = "STANDARD,SAVINGS"

// ListBalances returns all balances for a profile.
//
// Only visible, non-investment balances are returned. Wise exposes no
// per-balance endpoint, so there is no way to fetch a hidden or invested
// balance individually through this SDK.
func (c *Client) ListBalances(ctx context.Context, profileID ProfileID) ([]Balance, error) {
	path := fmt.Sprintf("/v4/profiles/%d/balances?types=%s", profileID.Get(), balanceTypesQuery)

	var balances []raw.Balance

	err := c.get(ctx, path, &balances)
	if err != nil {
		return nil, fmt.Errorf("list balances for profile %d: %w", profileID.Get(), err)
	}

	results := make([]Balance, 0, len(balances))
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
) (*Balance, error) {
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

func mapBalance(b raw.Balance) (Balance, error) {
	createdAt, err := parseWiseTimestamp(b.CreationTime)
	if err != nil {
		return Balance{}, errorfamily.WrapCorruption(
			err,
			"wise.balance.parse_creation_time",
			fmt.Sprintf("parse creation_time %q", b.CreationTime),
		)
	}

	balanceType, err := parseBalanceType(b.Type)
	if err != nil {
		return Balance{}, errorfamily.WrapCorruption(
			err,
			"wise.balance.parse_type",
			fmt.Sprintf("parse type %q", b.Type),
		)
	}

	currency, err := NewCurrency(b.Currency)
	if err != nil {
		return Balance{}, errorfamily.WrapCorruption(
			err,
			"wise.balance.parse_currency",
			fmt.Sprintf("currency %q", b.Currency),
		)
	}

	amount, err := toMoney(b.Amount)
	if err != nil {
		return Balance{}, fmt.Errorf("amount: %w", err)
	}

	reserved, err := toMoney(b.ReservedAmount)
	if err != nil {
		return Balance{}, fmt.Errorf("reserved amount: %w", err)
	}

	return Balance{
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
