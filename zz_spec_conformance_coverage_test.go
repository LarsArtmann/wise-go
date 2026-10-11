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

	verdict := evaluateConformanceCoverage()

	switch {
	case !verdict.suiteRan:
		fmt.Fprintln(
			os.Stdout,
			"spec conformance coverage: suite did not run (a -run filter likely excluded it); floors not asserted",
		)
	case verdict.violation != "":
		fmt.Fprintf(os.Stderr, "spec conformance coverage guard: %s\n", verdict.violation)

		code = 1
	default:
		templates, exemptPaths, exchanges, exemptAccountsLists := conformanceCoverageSnapshot()
		fmt.Fprintf(
			os.Stdout,
			"spec conformance: %d exchanges validated across %d distinct spec operations; %d exempt statement variants; %d exempt legacy recipient lists\n",
			exchanges,
			len(templates),
			len(exemptPaths),
			exemptAccountsLists,
		)
	}

	os.Exit(code)
}

// conformanceCoverageVerdict is the outcome of the post-run floor check.
type conformanceCoverageVerdict struct {
	// violation is the floor-violation message, empty when conforming.
	violation string
	// suiteRan reports whether the floors were assertable at all: without
	// the full Ginkgo suite (a -run filter excluded it) nothing meaningful
	// is recorded, even though standalone validator tests may record a few
	// exchanges.
	suiteRan bool
}

func evaluateConformanceCoverage() conformanceCoverageVerdict {
	if !conformanceSuiteCompleted {
		return conformanceCoverageVerdict{}
	}

	templates, exemptPaths, exchanges, _ := conformanceCoverageSnapshot()

	if _, err := loadConformanceSpec(); err != nil {
		return conformanceCoverageVerdict{
			violation: fmt.Sprintf("load spec snapshot %s: %v", specSnapshotPath, err),
			suiteRan:  true,
		}
	}

	if doc := conformanceLoaded.doc; doc.OpenAPI != "3.1.0" || len(doc.Paths.Map()) < 150 {
		return conformanceCoverageVerdict{
			violation: fmt.Sprintf(
				"spec snapshot %s looks wrong (openapi=%q paths=%d); the conformance results are not trustworthy",
				specSnapshotPath, doc.OpenAPI, len(doc.Paths.Map()),
			),
			suiteRan: true,
		}
	}

	if exchanges < minValidatedExchanges {
		return conformanceCoverageVerdict{
			violation: fmt.Sprintf(
				"spec conformance validated only %d exchanges (floor %d); the recorder hook must be broken",
				exchanges, minValidatedExchanges,
			),
			suiteRan: true,
		}
	}

	if len(templates) < minConformingTemplates {
		return conformanceCoverageVerdict{
			violation: fmt.Sprintf(
				"spec conformance covered only %d distinct spec operations (floor %d): %v; "+
					"the recorder hook or the suite must be broken",
				len(templates), minConformingTemplates, templates,
			),
			suiteRan: true,
		}
	}

	if len(exemptPaths) < minExemptStatementVariants {
		return conformanceCoverageVerdict{
			violation: fmt.Sprintf(
				"expected at least %d exempt statement format variants (csv/pdf/xlsx/xml), got %d: %v",
				minExemptStatementVariants, len(exemptPaths), exemptPaths,
			),
			suiteRan: true,
		}
	}

	return conformanceCoverageVerdict{suiteRan: true}
}
