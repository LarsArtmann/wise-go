package wise

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// quarterlyPathPattern matches a quarter-qualified endpoint path, e.g.
// /2026Q4/one-time-token/status or /2027Q1/balances/account-details/1. Bare
// quarter mentions (prose like "verified live on 2026Q4", the spec's unversioned
// template URLs) do not match — only a quarter followed by a path segment
// counts, because only those are endpoint contracts.
var quarterlyPathPattern = regexp.MustCompile(`/(20[0-9]{2}Q[1-4])/([A-Za-z{}.-]+)`)

// TestQuarterlyAPIVersionsConsistent pins every quarter-qualified endpoint path
// literal in the module — production code, test mocks, and the README — to the
// API-version constant that builds the real request. The quarterly surfaces
// roll over independently (one-time-token and webhook subscriptions on 2026Q4,
// account details on 2027Q1); without this pin a missed mock path or README
// example keeps serving the previous quarter while the client calls the new
// one. Consumers such as bank-sync pin their own mock paths and rely on the
// rollover ritual; this test is the SDK-side half of that contract.
func TestQuarterlyAPIVersionsConsistent(t *testing.T) {
	t.Parallel()

	wellFormedQuarter := regexp.MustCompile(`^20[0-9]{2}Q[1-4]$`)

	for constant, value := range map[string]string{
		"ottAPIVersion":                  ottAPIVersion,
		"webhookSubscriptionsAPIVersion": webhookSubscriptionsAPIVersion,
		"accountDetailsAPIVersion":       accountDetailsAPIVersion,
	} {
		if !wellFormedQuarter.MatchString(value) {
			t.Fatalf("%s = %q is not a well-formed Wise quarter", constant, value)
		}
	}

	// The path segment directly after the quarter identifies the surface.
	expected := map[string]string{
		"one-time-token": ottAPIVersion,
		"profiles":       webhookSubscriptionsAPIVersion, // subscription CRUD
		"balances":       accountDetailsAPIVersion,       // account-details GET
	}

	sources := []string{"README.md"}

	goFiles, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob root Go files: %v", err)
	}

	rawFiles, err := filepath.Glob(filepath.Join("internal", "raw", "*.go"))
	if err != nil {
		t.Fatalf("glob internal/raw Go files: %v", err)
	}

	sources = append(append(sources, goFiles...), rawFiles...)

	foundPerSurface := map[string]int{}

	for _, path := range sources {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		for _, match := range quarterlyPathPattern.FindAllStringSubmatch(string(content), -1) {
			found, segment := match[1], match[2]
			want, known := expected[segment]

			if !known {
				t.Errorf("%s: quarterly path /%s/%s is not wired to an API-version constant; "+
					"extend expected in TestQuarterlyAPIVersionsConsistent", path, found, segment)

				continue
			}

			foundPerSurface[segment]++

			if found != want {
				t.Errorf("%s: path quarter %s disagrees with the constant serving %q (%s)",
					path, found, segment, want)
			}
		}
	}

	// Guard against a vacuous pass: the scan must still find the live mocks.
	for segment, floor := range map[string]int{
		"one-time-token": 5,
		"profiles":       5,
		"balances":       3,
	} {
		if foundPerSurface[segment] < floor {
			t.Fatalf("found only %d %q quarterly paths (floor %d); the scan itself must be broken",
				foundPerSurface[segment], segment, floor)
		}
	}
}
