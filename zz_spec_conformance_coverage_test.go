package wise_test

// Coverage guard for the spec conformance harness: asserts the recorder
// actually saw a substantial slice of the SDK surface, so a broken hook
// cannot produce a vacuous pass. Floors are empirically pinned, not an
// endpoint inventory; the inventory lives in the client code and FEATURES.md.
//
// The floors live ONLY in TestMain's post-run assertion, making them
// shuffle-proof by construction: `go test -shuffle` can reorder Test
// functions freely (the old zz_-last per-test guard raced mid-order runs and
// could see partial counts — fixed 2026-10-07). The assertion fires only
// when the full Ginkgo suite completed; a -run filter that excludes it
// leaves the floors unasserted by design.

import (
	"fmt"
	"os"
	"testing"
)

const (
	minConformingTemplates     = 25
	minExemptStatementVariants = 4
	minValidatedExchanges      = 60
)

// conformanceSuiteCompleted is set by the Ginkgo bootstrap once the full
// suite has run; see TestWiseClient.
var conformanceSuiteCompleted bool

func TestMain(m *testing.M) {
	code := m.Run()
	if code != 0 {
		os.Exit(code)
	}

	violation, suiteRan := coverageViolation()

	switch {
	case !suiteRan:
		fmt.Println("spec conformance coverage: suite did not run (a -run filter likely excluded it); floors not asserted")
	case violation != "":
		fmt.Fprintf(os.Stderr, "spec conformance coverage guard: %s\n", violation)
		code = 1
	default:
		templates, exemptPaths, exchanges, exemptAccountsLists := conformanceCoverageSnapshot()
		fmt.Printf(
			"spec conformance: %d exchanges validated across %d distinct spec operations; %d exempt statement variants; %d exempt legacy recipient lists\n",
			exchanges,
			len(templates),
			len(exemptPaths),
			exemptAccountsLists,
		)
	}

	os.Exit(code)
}

// coverageViolation evaluates the recorder floors once: it returns the
// floor-violation message ("" when conforming) and whether the full Ginkgo
// suite ran, without which the assertion is meaningless (a -run filter that
// excluded it still records a few exchanges via standalone validator tests).
func coverageViolation() (violation string, suiteRan bool) {
	if !conformanceSuiteCompleted {
		return "", false
	}

	templates, exemptPaths, exchanges, _ := conformanceCoverageSnapshot()

	if _, err := loadConformanceSpec(); err != nil {
		return fmt.Sprintf("load spec snapshot %s: %v", specSnapshotPath, err), true
	}

	if doc := conformanceLoaded.doc; doc.OpenAPI != "3.1.0" || len(doc.Paths.Map()) < 150 {
		return fmt.Sprintf(
			"spec snapshot %s looks wrong (openapi=%q paths=%d); the conformance results are not trustworthy",
			specSnapshotPath, doc.OpenAPI, len(doc.Paths.Map()),
		), true
	}

	if exchanges < minValidatedExchanges {
		return fmt.Sprintf(
			"spec conformance validated only %d exchanges (floor %d); the recorder hook must be broken",
			exchanges, minValidatedExchanges,
		), true
	}

	if len(templates) < minConformingTemplates {
		return fmt.Sprintf(
			"spec conformance covered only %d distinct spec operations (floor %d): %v; "+
				"the recorder hook or the suite must be broken",
			len(templates), minConformingTemplates, templates,
		), true
	}

	if len(exemptPaths) < minExemptStatementVariants {
		return fmt.Sprintf(
			"expected at least %d exempt statement format variants (csv/pdf/xlsx/xml), got %d: %v",
			minExemptStatementVariants, len(exemptPaths), exemptPaths,
		), true
	}

	return "", true
}
