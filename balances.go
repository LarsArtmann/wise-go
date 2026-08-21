package wise

import (
	"context"
	"fmt"

	id "github.com/larsartmann/go-branded-id"
	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// Wise's wire values for balance types; the public BalanceType enum uses
// lowercase values, so both directions convert through these.
const (
	wireBalanceTypeStandard = "STANDARD"
	wireBalanceTypeSavings  = "SAVINGS"

	// balanceTypesQuery lists every balance type the SDK can map (keep in sync
	// with the parseBalanceType table). The v4 balances endpoint rejects
	// requests without a types filter with "query.types: NotNull", so
	// ListBalances always sends the full set and keeps visibility/investment
	// filtering client-side.
	balanceTypesQuery = wireBalanceTypeStandard + "," + wireBalanceTypeSavings
)

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

// GetBalance returns a specific balance by ID within a profile via the
// direct per-balance endpoint. Unlike ListBalances, it does not filter:
// hidden and invested balances are retrievable individually.
func (c *Client) GetBalance(
	ctx context.Context,
	profileID ProfileID,
	balanceID BalanceID,
) (*Balance, error) {
	if profileID.Get() == 0 {
		return nil, errorfamily.NewRejection(
			"wise.balance.invalid_request",
			"profileID is required",
		)
	}

	if balanceID.Get() == 0 {
		return nil, errorfamily.NewRejection(
			"wise.balance.invalid_request",
			"balanceID is required",
		)
	}

	path := fmt.Sprintf("/v4/profiles/%d/balances/%d", profileID.Get(), balanceID.Get())

	var balance raw.Balance

	if err := c.get(ctx, path, &balance); err != nil {
		return nil, fmt.Errorf("get balance %d for profile %d: %w", balanceID.Get(), profileID.Get(), err)
	}

	result, mapErr := mapBalance(balance)
	if mapErr != nil {
		return nil, fmt.Errorf("map balance %d: %w", balance.ID, mapErr)
	}

	return &result, nil
}

// CreateBalanceRequest opens a new balance for a profile. Name is required
// for SAVINGS balances.
type CreateBalanceRequest struct {
	ProfileID ProfileID
	Currency  Currency
	Type      BalanceType // BalanceTypeStandard or BalanceTypeSavings.
	Name      string
}

// CreateBalance opens a new balance for a profile
// (POST /v4/profiles/{profileId}/balances) and returns it.
func (c *Client) CreateBalance(ctx context.Context, req CreateBalanceRequest) (*Balance, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	body := map[string]any{
		wireKeyCurrency: string(req.Currency),
		wireKeyType:     balanceTypeWire(req.Type),
	}

	if req.Name != "" {
		body["name"] = req.Name
	}

	var balance raw.Balance

	if err := c.post(ctx, fmt.Sprintf("/v4/profiles/%d/balances", req.ProfileID.Get()), body, &balance); err != nil {
		return nil, fmt.Errorf("create %s balance for profile %d: %w", req.Currency, req.ProfileID.Get(), err)
	}

	result, mapErr := mapBalance(balance)
	if mapErr != nil {
		return nil, fmt.Errorf("map created balance %d: %w", balance.ID, mapErr)
	}

	return &result, nil
}

func (r CreateBalanceRequest) validate() error {
	if r.ProfileID.Get() == 0 {
		return errorfamily.NewRejection("wise.balance.invalid_request", "profileID is required")
	}

	if _, err := NewCurrency(string(r.Currency)); err != nil {
		return errorfamily.NewRejection("wise.balance.invalid_request", "currency is required")
	}

	if r.Type != BalanceTypeStandard && r.Type != BalanceTypeSavings {
		return errorfamily.NewRejection(
			"wise.balance.invalid_request",
			"type must be BalanceTypeStandard or BalanceTypeSavings",
		)
	}

	if r.Type == BalanceTypeSavings && r.Name == "" {
		return errorfamily.NewRejection("wise.balance.invalid_request", "name is required for savings balances")
	}

	return nil
}

// balanceTypeWire renders the public lowercase BalanceType as Wise's
// uppercase wire value.
func balanceTypeWire(balanceType BalanceType) string {
	if balanceType == BalanceTypeSavings {
		return wireBalanceTypeSavings
	}

	return wireBalanceTypeStandard
}

// TotalFunds is a profile-wide funds overview in a single currency.
type TotalFunds struct {
	Worth     Money // Cash plus the valuation of any invested portfolio.
	Available Money // Cash plus any approved overdraft limit.
}

// GetTotalFunds returns a profile-wide funds overview converted to the given
// currency (GET /v4/profiles/{profileId}/total-funds/{currency}).
func (c *Client) GetTotalFunds(
	ctx context.Context,
	profileID ProfileID,
	currency Currency,
) (*TotalFunds, error) {
	if profileID.Get() == 0 {
		return nil, errorfamily.NewRejection("wise.total_funds.invalid_request", "profileID is required")
	}

	if _, err := NewCurrency(string(currency)); err != nil {
		return nil, errorfamily.NewRejection("wise.total_funds.invalid_request", "currency is required")
	}

	path := fmt.Sprintf("/v4/profiles/%d/total-funds/%s", profileID.Get(), string(currency))

	var funds raw.TotalFunds

	if err := c.get(ctx, path, &funds); err != nil {
		return nil, fmt.Errorf("get total funds for profile %d in %s: %w", profileID.Get(), currency, err)
	}

	worth, err := toMoney(funds.TotalWorth)
	if err != nil {
		return nil, fmt.Errorf("total worth: %w", err)
	}

	available, err := toMoney(funds.TotalAvailable)
	if err != nil {
		return nil, fmt.Errorf("total available: %w", err)
	}

	return &TotalFunds{Worth: worth, Available: available}, nil
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
		wireBalanceTypeStandard: BalanceTypeStandard,
		wireBalanceTypeSavings:  BalanceTypeSavings,
	}, s, "balance type")
}
