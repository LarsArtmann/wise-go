package wise

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/larsartmann/wise-go/internal/raw"
)

// activitiesPageSizeMax is the largest page size the endpoint documents
// (size: minimum 1, maximum 100; Wise's own default is 10).
const activitiesPageSizeMax = 100

// ActivityType is Wise's kind of activity. The documented values are listed
// as constants; Wise may add more, and unrecognised values are preserved.
type ActivityType string

// Documented activity kinds (Wise OpenAPI 2027Q1, ActivityType).
const (
	ActivityTypeAcquiringPayment       ActivityType = "ACQUIRING_PAYMENT"
	ActivityTypeAutoConversion         ActivityType = "AUTO_CONVERSION"
	ActivityTypeBalanceAdjustment      ActivityType = "BALANCE_ADJUSTMENT"
	ActivityTypeBalanceAssetFee        ActivityType = "BALANCE_ASSET_FEE"
	ActivityTypeBalanceCashback        ActivityType = "BALANCE_CASHBACK"
	ActivityTypeBalanceDeposit         ActivityType = "BALANCE_DEPOSIT"
	ActivityTypeBalanceHoldFee         ActivityType = "BALANCE_HOLD_FEE"
	ActivityTypeBalanceInterest        ActivityType = "BALANCE_INTEREST"
	ActivityTypeBankDetailsOrder       ActivityType = "BANK_DETAILS_ORDER"
	ActivityTypeBatchTransfer          ActivityType = "BATCH_TRANSFER"
	ActivityTypeCardCashback           ActivityType = "CARD_CASHBACK"
	ActivityTypeCardCheck              ActivityType = "CARD_CHECK"
	ActivityTypeCardOrder              ActivityType = "CARD_ORDER"
	ActivityTypeCardPayment            ActivityType = "CARD_PAYMENT"
	ActivityTypeCashWithdrawal         ActivityType = "CASH_WITHDRAWAL"
	ActivityTypeClaimableSendOrder     ActivityType = "CLAIMABLE_SEND_ORDER"
	ActivityTypeDirectDebitTransaction ActivityType = "DIRECT_DEBIT_TRANSACTION"
	ActivityTypeExcessRefund           ActivityType = "EXCESS_REFUND"
	ActivityTypeFeeRefund              ActivityType = "FEE_REFUND"
	ActivityTypeIncorporationOrder     ActivityType = "INCORPORATION_ORDER"
	ActivityTypeInterbalance           ActivityType = "INTERBALANCE"
	ActivityTypePaymentRequest         ActivityType = "PAYMENT_REQUEST"
	ActivityTypePrefundingTransfer     ActivityType = "PREFUNDING_TRANSFER"
	ActivityTypeReward                 ActivityType = "REWARD"
	ActivityTypeScheduledSendOrder     ActivityType = "SCHEDULED_SEND_ORDER"
	ActivityTypeTransfer               ActivityType = "TRANSFER"
)

// ActivityStatus is the lifecycle state of an activity. Wise may add values;
// unrecognised ones are preserved.
type ActivityStatus string

// Documented activity statuses (Wise OpenAPI 2027Q1, ActivityStatus).
const (
	ActivityStatusRequiresAttention ActivityStatus = "REQUIRES_ATTENTION"
	ActivityStatusInProgress        ActivityStatus = "IN_PROGRESS"
	ActivityStatusUpcoming          ActivityStatus = "UPCOMING"
	ActivityStatusCompleted         ActivityStatus = "COMPLETED"
	ActivityStatusCancelled         ActivityStatus = "CANCELLED"
)

// ActivityResourceType names the family of the underlying entity an
// activity snapshots. It is also the endpoint's monetaryResourceType filter.
type ActivityResourceType string

// Documented resource types (Wise OpenAPI 2027Q1, ActivityResourceType).
const (
	ActivityResourceTypeAccrualCharge          ActivityResourceType = "ACCRUAL_CHARGE"
	ActivityResourceTypeAcquiringPayment       ActivityResourceType = "ACQUIRING_PAYMENT"
	ActivityResourceTypeAssetsWithdrawal       ActivityResourceType = "ASSETS_WITHDRAWAL"
	ActivityResourceTypeBalanceCashback        ActivityResourceType = "BALANCE_CASHBACK"
	ActivityResourceTypeBalanceInterest        ActivityResourceType = "BALANCE_INTEREST"
	ActivityResourceTypeBalanceTransaction     ActivityResourceType = "BALANCE_TRANSACTION"
	ActivityResourceTypeBankDetailsOrder       ActivityResourceType = "BANK_DETAILS_ORDER"
	ActivityResourceTypeBatchTransfer          ActivityResourceType = "BATCH_TRANSFER"
	ActivityResourceTypeCardCashback           ActivityResourceType = "CARD_CASHBACK"
	ActivityResourceTypeCardOrder              ActivityResourceType = "CARD_ORDER"
	ActivityResourceTypeCardTransaction        ActivityResourceType = "CARD_TRANSACTION"
	ActivityResourceTypeDirectDebitInstruction ActivityResourceType = "DIRECT_DEBIT_INSTRUCTION"
	ActivityResourceTypeDirectDebitTransaction ActivityResourceType = "DIRECT_DEBIT_TRANSACTION"
	ActivityResourceTypeFeeRefund              ActivityResourceType = "FEE_REFUND"
	ActivityResourceTypeIncidentRefund         ActivityResourceType = "INCIDENT_REFUND"
	ActivityResourceTypeIncorporationOrder     ActivityResourceType = "INCORPORATION_ORDER"
	ActivityResourceTypeOperationalTransaction ActivityResourceType = "OPERATIONAL_TRANSACTION"
	ActivityResourceTypePaymentRequest         ActivityResourceType = "PAYMENT_REQUEST"
	ActivityResourceTypeReward                 ActivityResourceType = "REWARD"
	ActivityResourceTypeRewardsRedemption      ActivityResourceType = "REWARDS_REDEMPTION"
	ActivityResourceTypeSendOrder              ActivityResourceType = "SEND_ORDER"
	ActivityResourceTypeSendOrderExecution     ActivityResourceType = "SEND_ORDER_EXECUTION"
	ActivityResourceTypeTransfer               ActivityResourceType = "TRANSFER"
)

// Activity is one entry of a profile's activity feed: a snapshot of a
// performed action (a transfer, a card payment, a deposit, a conversion)
// with display-oriented amounts.
//
// The feed is NOT the statement ledger: it is profile-scoped rather than
// balance-scoped, its amounts are formatted text, and it does not itemize
// fees. Where the two overlap they can differ (live comparison 2026-11-26 on
// one profile: about half of the statement's debit rows had no activity of
// the same amount, mostly transfers whose statement amount includes a fee).
// Read it as real-time visibility, never as the book of record.
type Activity struct {
	// ID is Wise's opaque activity id (stable for one activity).
	ID   string
	Type ActivityType
	// ResourceType and ResourceID identify the underlying entity (for
	// example CARD_TRANSACTION 343049).
	ResourceType ActivityResourceType
	ResourceID   string
	// Title may carry Wise's display markup (<strong>, <positive>).
	Title       string
	Description string
	// PrimaryAmount and SecondaryAmount are the activity's two money legs.
	// Wise puts the main amount first and a second leg (the converted
	// amount, or the source amount of a conversion) second; SecondaryAmount
	// is absent for single-leg activities.
	PrimaryAmount   ActivityAmount
	SecondaryAmount ActivityAmount
	Status          ActivityStatus
	CreatedOn       time.Time
	UpdatedOn       time.Time
}

// ActivityAmount is one money leg of an activity. Wise formats the feed for
// display, so the wire value is text with units, an optional sign, and
// optional <positive> markup. The SDK parses one plain amount into Money
// and keeps the raw text for everything else.
type ActivityAmount struct {
	// Text is Wise's display text exactly as sent ("" when none was sent).
	Text string
	// Present reports that Text carried one plain amount; Amount is
	// meaningful only then. A currency pair such as "USD, EUR" or a locale
	// format other than the canonical one leaves Present false.
	Present bool
	// Credit reports that Wise marked the leg as money IN (a leading "+" or
	// the <positive> markup). A present leg without Credit is money OUT or,
	// for a two-legged conversion, the source side.
	Credit bool
	// Amount is the leg's absolute value.
	Amount Money
}

// ListActivitiesRequest parameters for listing a profile's activity feed.
//
// Since and Until bound the window on the activity's creation time (both
// inclusive) and are omitted when zero. ResourceType (the endpoint's
// monetaryResourceType) and Status are optional filters. Cursor is the
// NextCursor of a previous page. Size is the page size (1 to 100); zero
// leaves Wise's default of 10.
type ListActivitiesRequest struct {
	ProfileID    ProfileID
	Since        time.Time
	Until        time.Time
	ResourceType ActivityResourceType
	Status       ActivityStatus
	Cursor       string
	Size         int
}

// ActivityPage is one page of a profile's activity feed.
type ActivityPage struct {
	// Activities are ordered newest first (Wise API order).
	Activities []Activity
	// NextCursor is the cursor of the following page; empty on the last page.
	NextCursor string
}

// ListActivities returns one page of a profile's activity feed
// (GET /v1/profiles/{profileId}/activities): snapshots of performed actions
// such as transfers, card payments, deposits, and conversions, newest first.
//
// The endpoint is not SCA-protected and serves a personal API token (live
// 2026-11-26 on five profiles: 200 for each, including four whose balance
// statements answer 403). It is the surface behind Wise's
// wise_list_activities MCP tool. See [Activity] for what the feed is and is
// not.
func (c *Client) ListActivities(ctx context.Context, req ListActivitiesRequest) (*ActivityPage, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}

	var page raw.ActivityPage

	path := fmt.Sprintf("/v1/profiles/%d/activities", req.ProfileID.Get())

	if err := c.getWithQuery(ctx, path, req.query, &page); err != nil {
		return nil, fmt.Errorf("list activities for profile %d: %w", req.ProfileID.Get(), err)
	}

	result := &ActivityPage{
		Activities: make([]Activity, 0, len(page.Activities)),
		NextCursor: "",
	}

	for _, activity := range page.Activities {
		mapped, err := mapActivity(activity)
		if err != nil {
			return nil, err
		}

		result.Activities = append(result.Activities, mapped)
	}

	if page.Cursor != nil {
		result.NextCursor = *page.Cursor
	}

	return result, nil
}

func (r ListActivitiesRequest) validate() error {
	if err := requireID(r.ProfileID, errorCodeActivityInvalidRequest, "profileID"); err != nil {
		return err
	}

	if r.Size < 0 || r.Size > activitiesPageSizeMax {
		return errorfamily.NewRejection(
			errorCodeActivityInvalidRequest,
			fmt.Sprintf("size must be between 1 and %d (or 0 for Wise's default)", activitiesPageSizeMax),
		)
	}

	return nil
}

// query renders the request as the endpoint's query string; unset filters
// are omitted.
func (r ListActivitiesRequest) query() string {
	values := url.Values{}

	if !r.Since.IsZero() {
		values.Set("since", formatWiseTimestamp(r.Since))
	}

	if !r.Until.IsZero() {
		values.Set("until", formatWiseTimestamp(r.Until))
	}

	if r.ResourceType != "" {
		values.Set("monetaryResourceType", string(r.ResourceType))
	}

	if r.Status != "" {
		values.Set("status", string(r.Status))
	}

	if r.Cursor != "" {
		values.Set("nextCursor", r.Cursor)
	}

	if r.Size > 0 {
		values.Set("size", strconv.Itoa(r.Size))
	}

	return values.Encode()
}

func mapActivity(activity raw.Activity) (Activity, error) {
	createdOn, err := parseWiseTimestamp(activity.CreatedOn)
	if err != nil {
		return Activity{}, errorfamily.WrapCorruption(
			err,
			errorCodeActivityMap,
			fmt.Sprintf("parse createdOn %q of activity %q", activity.CreatedOn, activity.ID),
		)
	}

	updatedOn, err := parseWiseTimestamp(activity.UpdatedOn)
	if err != nil {
		return Activity{}, errorfamily.WrapCorruption(
			err,
			errorCodeActivityMap,
			fmt.Sprintf("parse updatedOn %q of activity %q", activity.UpdatedOn, activity.ID),
		)
	}

	return Activity{
		ID:              activity.ID,
		Type:            ActivityType(activity.Type),
		ResourceType:    ActivityResourceType(activity.Resource.Type),
		ResourceID:      activity.Resource.ID,
		Title:           activity.Title,
		Description:     activity.Description,
		PrimaryAmount:   parseActivityAmount(activity.PrimaryAmount),
		SecondaryAmount: parseActivityAmount(activity.SecondaryAmount),
		Status:          ActivityStatus(activity.Status),
		CreatedOn:       createdOn,
		UpdatedOn:       updatedOn,
	}, nil
}

// activityAmountPattern reads Wise's canonical formatted amount: optional
// <positive> markup, an optional sign, an amount with optional comma
// grouping and at most two decimals, a space, and the ISO 4217 code. It
// deliberately matches nothing else: a locale format such as
// "1.234,56 EUR" must leave the amount unparsed rather than be read as a
// thousandth of its value.
var activityAmountPattern = regexp.MustCompile(
	`^(?:<positive>)?\s*([+-])?\s*([0-9]+(?:,[0-9]{3})*(?:\.[0-9]{1,2})?) ([A-Z]{3})(?:</positive>)?$`,
)

// activityAmountWholeDigits bounds the whole part so the cent value always
// fits an int64.
const activityAmountWholeDigits = 15

// parseActivityAmount reads one formatted activity amount. Anything that is
// not one canonical amount (an empty string, a currency pair, another locale
// format, an unrepresentable value) yields Present false with the text kept.
func parseActivityAmount(text string) ActivityAmount {
	result := ActivityAmount{Text: text, Present: false, Credit: false, Amount: Money{Cents: 0, Currency: ""}}

	match := activityAmountPattern.FindStringSubmatch(strings.TrimSpace(text))
	if match == nil {
		return result
	}

	whole, fraction, _ := strings.Cut(strings.ReplaceAll(match[2], ",", ""), ".")

	if len(whole) > activityAmountWholeDigits {
		return result
	}

	wholeValue, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return result
	}

	if len(fraction) == 1 {
		fraction += "0"
	}

	if fraction == "" {
		fraction = "00"
	}

	fractionValue, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return result
	}

	currency, err := NewCurrency(match[3])
	if err != nil {
		return result
	}

	result.Present = true
	result.Credit = match[1] == "+" || strings.Contains(text, "<positive>")
	result.Amount = Money{Cents: wholeValue*100 + fractionValue, Currency: currency}

	return result
}
