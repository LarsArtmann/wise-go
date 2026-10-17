package wise_test

// Live smoke tests against the Wise sandbox (api.wise-sandbox.com). They run
// ONLY when WISE_SANDBOX_API_KEY is set; without the variable every test in
// this file skips, so local `go test ./...` and CI stay network-free.
// See the README "Sandbox verification" section for how to run them.

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/wise-go"
)

// newSandboxClient returns a client pointed at the Wise sandbox, or skips
// the calling test when no sandbox API key is configured.
func newSandboxClient(t *testing.T) *wise.Client {
	t.Helper()

	apiKey := os.Getenv("WISE_SANDBOX_API_KEY")
	if apiKey == "" {
		t.Skip("WISE_SANDBOX_API_KEY not set — skipping sandbox live tests (see README: Sandbox verification)")
	}

	return wise.New(apiKey, wise.WithSandbox())
}

// TestSandboxLiveGetMe verifies the API key and the sandbox baseline: token
// resolution, TLS, request signing, and response decoding.
func TestSandboxLiveGetMe(t *testing.T) {
	client := newSandboxClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tokenOwner, err := client.GetMe(ctx)
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}

	if tokenOwner.Email == "" {
		t.Errorf("GetMe returned an empty email — unexpected for a valid sandbox key")
	}

	t.Logf("sandbox identity: id=%d name=%q email=%q", tokenOwner.ID.Get(), tokenOwner.Name, tokenOwner.Email)
}

// TestSandboxLiveListTransfers verifies the transfers read end to end: the
// endpoint is not SCA-protected, so a bare personal token must get a list
// (possibly empty) rather than an auth failure.
func TestSandboxLiveListTransfers(t *testing.T) {
	client := newSandboxClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	transfers, err := client.ListTransfers(ctx, wise.ListTransfersRequest{})
	if err != nil {
		t.Fatalf("ListTransfers: %v", err)
	}

	// A fresh sandbox profile legitimately has zero transfers; the assertion
	// is that the call succeeds and auto-pagination terminates.
	t.Logf("sandbox transfers: %d", len(transfers))
}

// TestSandboxLiveGetTransfer fetches one transfer by ID. It needs a transfer
// the key can actually read, so it stays silent unless WISE_TEST_TRANSFER_ID
// is set — a bare key with no data would otherwise fail misleadingly.
func TestSandboxLiveGetTransfer(t *testing.T) {
	client := newSandboxClient(t)

	id := os.Getenv("WISE_TEST_TRANSFER_ID")
	if id == "" {
		t.Skip("WISE_TEST_TRANSFER_ID not set — skipping GetTransfer smoke (set it to a transfer ID the key owns)")
	}

	transferID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		t.Fatalf("WISE_TEST_TRANSFER_ID %q is not an int64: %v", id, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	transfer, err := client.GetTransfer(ctx, wise.NewTransferID(transferID))
	if err != nil {
		t.Fatalf("GetTransfer(%d): %v", transferID, err)
	}

	if transfer.Source.Currency == "" || transfer.Target.Currency == "" {
		t.Errorf("GetTransfer(%d) returned empty Money currencies: %+v", transferID, transfer)
	}

	t.Logf("transfer %d: status=%s source=%s target=%s",
		transferID, transfer.Status, transfer.Source, transfer.Target)
}

// firstSandboxProfile returns the key's first profile (sandbox accounts ship
// a personal and a business one) as the stand-in profile for read paths.
func firstSandboxProfile(t *testing.T, client *wise.Client) wise.Profile {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	profiles, err := client.ListProfiles(ctx)
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}
	if len(profiles) == 0 {
		t.Fatal("ListProfiles returned no profiles — unexpected for a valid sandbox key")
	}

	return profiles[0]
}

// TestSandboxLiveGetExchangeRate pins the /v1/rates wire shape against the
// live sandbox (assumption #4 of the 2026-10-17 verification pass): both the
// current and the historical lookup decode, and the rate timestamp — the
// no-millis numeric-zone layout that broke naive parsing in production — is
// non-zero.
func TestSandboxLiveGetExchangeRate(t *testing.T) {
	client := newSandboxClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	current, err := client.GetExchangeRate(ctx, wise.Currency("EUR"), wise.Currency("USD"), time.Time{})
	if err != nil {
		t.Fatalf("GetExchangeRate(EUR/USD, current): %v", err)
	}
	if current.Rate <= 0 {
		t.Errorf("current rate %v is not positive", current.Rate)
	}
	if current.Time.IsZero() {
		t.Error("current rate Time is zero — the timestamp did not decode")
	}

	historical, err := client.GetExchangeRate(
		ctx, wise.Currency("EUR"), wise.Currency("USD"), time.Now().UTC().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("GetExchangeRate(EUR/USD, -24h): %v", err)
	}
	if historical.Rate <= 0 {
		t.Errorf("historical rate %v is not positive", historical.Rate)
	}
	if historical.Time.IsZero() {
		t.Error("historical rate Time is zero — the timestamp did not decode")
	}

	t.Logf("rates: current=%f@%s historical=%f@%s",
		current.Rate, current.Time.Format(time.RFC3339),
		historical.Rate, historical.Time.Format(time.RFC3339))
}

// TestSandboxLiveFundTransferNotFound pins the funding-error contract
// (assumption #5, verified SDK-side 2026-10-17): funding a nonexistent
// transfer surfaces Wise's typed 404. The wire body arrives as
// application/json with the router-shaped envelope — the spec's text/plain
// note applies to domain funding errors only.
func TestSandboxLiveFundTransferNotFound(t *testing.T) {
	client := newSandboxClient(t)
	profile := firstSandboxProfile(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.FundTransfer(ctx, profile.ID, wise.NewTransferID(1))
	notFound, ok := errors.AsType[*wise.NotFoundError](err)
	if !ok {
		t.Fatalf("FundTransfer(nonexistent): expected *NotFoundError, got %v", err)
	}
	if notFound.StatusCode != http.StatusNotFound {
		t.Errorf("FundTransfer(nonexistent) status = %d, want 404", notFound.StatusCode)
	}
}

// raiseSandboxSCAChallenge triggers Wise's strong-customer-authentication
// flow by requesting a statement file for a real balance: sandbox statements
// are SCA-protected regardless of profile type (2026-10-17 lane evidence).
func raiseSandboxSCAChallenge(t *testing.T, client *wise.Client) *wise.SCAChallengeError {
	t.Helper()

	profile := firstSandboxProfile(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	balances, err := client.ListBalances(ctx, profile.ID)
	if err != nil {
		t.Fatalf("ListBalances: %v", err)
	}
	if len(balances) == 0 {
		t.Fatal("ListBalances returned no balances — expected the pre-funded sandbox balances")
	}

	_, err = client.GetStatement(ctx, wise.GetStatementRequest{
		ProfileID: profile.ID,
		BalanceID: balances[0].ID,
		Currency:  balances[0].Currency,
		From:      time.Now().UTC().Add(-30 * 24 * time.Hour),
		To:        time.Now().UTC(),
		Format:    wise.StatementFormatCSV,
	})
	challenge, ok := errors.AsType[*wise.SCAChallengeError](err)
	if !ok {
		t.Fatalf("GetStatement: expected *SCAChallengeError, got %v", err)
	}

	return challenge
}

// TestSandboxLiveSCAChallengeShape pins the SCA wire contract end to end:
// HTTP 403 with the verdict and one-time token in the x-2fa-approval
// headers, classified as *SCAChallengeError (never a plain *AuthError),
// with the OTT retrievable via TwoFAApprovalToken and never rendered into
// the error string.
func TestSandboxLiveSCAChallengeShape(t *testing.T) {
	client := newSandboxClient(t)
	challenge := raiseSandboxSCAChallenge(t, client)

	if challenge.StatusCode != http.StatusForbidden {
		t.Errorf("SCA challenge status = %d, want 403", challenge.StatusCode)
	}
	if verdict := challenge.Headers.Get(wise.HeaderTwoFAApprovalResult); verdict == "" {
		t.Error("x-2fa-approval-result header is empty — the SCA verdict is missing")
	}
	if challenge.TwoFAApprovalToken().IsZero() {
		t.Error("SCA challenge carried no one-time token — cannot clear the challenge")
	}

	var asError error = challenge
	if _, ok := errors.AsType[*wise.AuthError](asError); ok {
		t.Error("SCA challenge classified as *AuthError — must be *SCAChallengeError")
	}
	if message := challenge.Error(); message == "" || strings.Contains(message, string(challenge.TwoFAApprovalToken())) {
		t.Errorf("SCA challenge error string is empty or leaks the one-time token")
	}
}

// TestSandboxLiveGetOTTStatus walks the OTT lifecycle against the quarterly
// versioned surface (2026Q4 since the 2026-10-16 rollover): an SCA-raised
// token resolves to a status with the statement action, a bounded validity
// window, and one pending SIGNATURE challenge with no phone alternatives.
func TestSandboxLiveGetOTTStatus(t *testing.T) {
	client := newSandboxClient(t)
	challenge := raiseSandboxSCAChallenge(t, client)

	ott := challenge.TwoFAApprovalToken()
	if ott.IsZero() {
		t.Fatal("SCA challenge carried no one-time token — cannot query OTT status")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	status, err := client.GetOTTStatus(ctx, ott)
	if err != nil {
		t.Fatalf("GetOTTStatus: %v", err)
	}
	if status.ActionType != wise.OTTAction("BALANCE__GET_STATEMENT") {
		t.Errorf("OTT action = %q, want BALANCE__GET_STATEMENT", status.ActionType)
	}
	if status.Validity <= 0 || status.Validity > time.Hour {
		t.Errorf("OTT validity = %s, want within (0, 1h]", status.Validity)
	}

	pending := status.PendingRequired()
	if len(pending) == 0 {
		t.Fatal("OTT status reports no pending required challenges")
	}
	// SIGNATURE is not in the curated challenge-type constants (open enum —
	// sandbox presents it; SMS/WhatsApp/voice would be clearable via phone).
	if pending[0].Primary.Type != wise.OTTChallengeType("SIGNATURE") {
		t.Errorf("pending challenge type = %q, want SIGNATURE", pending[0].Primary.Type)
	}
	if len(pending[0].Alternatives) != 0 {
		t.Errorf("SIGNATURE challenge unexpectedly offers %d alternatives — the sandbox web login is the only clear path", len(pending[0].Alternatives))
	}
}

// TestSandboxLiveCreateBalanceIdempotent pins the balance-creation contract
// (assumption #2, upgraded to VERIFIED 2026-10-17 after the initially
// recorded 403 lock stopped reproducing): creation round-trips in sandbox,
// and re-sending the same X-idempotence-uuid returns the already-created
// balance instead of erroring or duplicating. The fixed idempotency key and
// currency keep repeated test runs non-mutating after the first — every
// later run replays the same balance.
func TestSandboxLiveCreateBalanceIdempotent(t *testing.T) {
	client := newSandboxClient(t)
	profile := firstSandboxProfile(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := wise.CreateBalanceRequest{
		ProfileID:      profile.ID,
		Currency:       wise.Currency("NOK"),
		Type:           wise.BalanceTypeStandard,
		Name:           "wise-go sandbox live test",
		IdempotencyKey: "6f2a9c34-8e17-4b5d-a0f2-93c1e7b8d4a6",
	}

	first, err := client.CreateBalance(ctx, req)
	if err != nil {
		t.Fatalf("CreateBalance(NOK): %v", err)
	}
	if first.ID.IsZero() || first.Currency != wise.Currency("NOK") || first.Type != wise.BalanceTypeStandard {
		t.Errorf("CreateBalance(NOK) returned %+v — unexpected shape for a fresh creation", first)
	}

	second, err := client.CreateBalance(ctx, req)
	if err != nil {
		t.Fatalf("CreateBalance(NOK, replay): %v", err)
	}
	if second.ID != first.ID {
		t.Errorf("idempotent replay returned balance %d, want the original %d", second.ID.Get(), first.ID.Get())
	}
}
