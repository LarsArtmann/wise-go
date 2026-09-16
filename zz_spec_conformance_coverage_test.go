package wise_test

// Coverage guard for the spec conformance harness: runs after the whole
// Ginkgo suite (file name sorts last) and asserts the recorder actually saw
// a substantial slice of the SDK surface, so a broken hook cannot produce a
// vacuous pass. Floors are empirically pinned, not an endpoint inventory;
// the inventory lives in the client code and FEATURES.md.

import (
	"testing"
)

const (
	minConformingTemplates     = 25
	minExemptStatementVariants = 3
	minValidatedExchanges      = 60
)

func TestSpecConformanceCoverage(t *testing.T) {
	templates, exemptPaths, exchanges := conformanceCoverageSnapshot()

	if _, err := loadConformanceSpec(); err != nil {
		t.Fatalf("load spec snapshot %s: %v", specSnapshotPath, err)
	}

	if doc := conformanceLoaded.doc; doc.OpenAPI != "3.1.0" || len(doc.Paths.Map()) < 150 {
		t.Fatalf(
			"spec snapshot %s looks wrong (openapi=%q paths=%d); the conformance results are not trustworthy",
			specSnapshotPath, doc.OpenAPI, len(doc.Paths.Map()),
		)
	}

	if exchanges < minValidatedExchanges {
		t.Fatalf(
			"spec conformance validated only %d exchanges (floor %d); the recorder hook must be broken",
			exchanges, minValidatedExchanges,
		)
	}

	if len(templates) < minConformingTemplates {
		t.Fatalf(
			"spec conformance covered only %d distinct spec operations (floor %d): %v; "+
				"the recorder hook or the suite must be broken",
			len(templates), minConformingTemplates, templates,
		)
	}

	if len(exemptPaths) < minExemptStatementVariants {
		t.Fatalf(
			"expected at least %d exempt statement format variants (csv/pdf/xlsx), got %d: %v",
			minExemptStatementVariants, len(exemptPaths), exemptPaths,
		)
	}

	t.Logf(
		"spec conformance: %d exchanges validated across %d distinct spec operations; %d exempt statement variants",
		exchanges, len(templates), len(exemptPaths),
	)
}
