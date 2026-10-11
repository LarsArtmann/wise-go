package wise_test

// Live smoke tests against the Wise sandbox (api.wise-sandbox.com). They run
// ONLY when WISE_SANDBOX_API_KEY is set; without the variable every test in
// this file skips, so local `go test ./...` and CI stay network-free.
// See the README "Sandbox verification" section for how to run them.

import (
	"context"
	"os"
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
