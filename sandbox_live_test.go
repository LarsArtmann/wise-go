package wise_test

// Live smoke tests against the Wise sandbox (api.wise-sandbox.com). They run
// ONLY when WISE_SANDBOX_API_KEY is set; without the variable every test in
// this file skips, so local `go test ./...` and CI stay network-free.
// See the README "Sandbox verification" section for how to run them.

import (
	"context"
	"os"
	"strconv"
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
