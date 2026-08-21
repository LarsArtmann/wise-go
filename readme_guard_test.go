package wise_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestREADMECodeFencesReferenceRealSymbols is a drift guard over the README:
// every client.Method or wise.Symbol mentioned inside a Go code fence must
// exist in the package's exported surface. The README documents the API by
// example; this test makes stale examples a test failure instead of a lie
// readers discover at compile time.
func TestREADMECodeFencesReferenceRealSymbols(t *testing.T) {
	t.Parallel()

	exported := exportedAPINames(t)
	referenced := symbolsReferencedByREADME(t)

	// Guard against vacuous passes: if the fence or symbol extraction breaks
	// (fence style change, regex drift), the loop below would trivially pass.
	if len(referenced) < 10 {
		t.Fatalf(
			"extracted only %d API references from README code fences; "+
				"the extraction itself must be broken",
			len(referenced),
		)
	}

	for _, mustExist := range []string{"New", "GetStatement", "Currency"} {
		if !referenced[mustExist] {
			t.Fatalf("expected README code fences to reference %s; extraction is broken", mustExist)
		}
	}

	var stale []string

	for name := range referenced {
		if !exported[name] {
			stale = append(stale, name)
		}
	}

	sort.Strings(stale)

	for _, name := range stale {
		t.Errorf(
			"README references %s, which does not exist in the package's exported API;\n"+
				"the README example has drifted from the code — fix the example or restore the symbol",
			name,
		)
	}
}

// exportedAPINames parses the non-test sources of the package and collects
// every exported top-level declaration plus every exported method on Client.
// Parsing beats maintaining a hand-written list: the list itself would drift.
func exportedAPINames(t *testing.T) map[string]bool {
	t.Helper()

	sources, err := nonTestGoFiles(".")
	if err != nil {
		t.Fatalf("list package sources: %v", err)
	}

	fset := token.NewFileSet()

	exported := make(map[string]bool)

	for _, source := range sources {
		file, parseErr := parser.ParseFile(fset, source, nil, parser.SkipObjectResolution)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", source, parseErr)
		}

		collectExportedNames(file, exported)
	}

	return exported
}

func nonTestGoFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	var names []string

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		names = append(names, filepath.Join(dir, name))
	}

	return names, nil
}

func collectExportedNames(file *ast.File, exported map[string]bool) {
	for _, decl := range file.Decls {
		switch node := decl.(type) {
		case *ast.FuncDecl:
			collectExportedFunc(node, exported)
		case *ast.GenDecl:
			for _, spec := range node.Specs {
				collectExportedSpec(spec, exported)
			}
		}
	}
}

func collectExportedFunc(node *ast.FuncDecl, exported map[string]bool) {
	if !node.Name.IsExported() {
		return
	}

	if node.Recv == nil || receiverIsClient(node.Recv) {
		exported[node.Name.Name] = true
	}
}

func collectExportedSpec(spec ast.Spec, exported map[string]bool) {
	switch spec := spec.(type) {
	case *ast.TypeSpec:
		exported[spec.Name.Name] = spec.Name.IsExported()
	case *ast.ValueSpec:
		for _, name := range spec.Names {
			exported[name.Name] = name.IsExported()
		}
	}
}

func receiverIsClient(receiver *ast.FieldList) bool {
	if receiver == nil || len(receiver.List) != 1 {
		return false
	}

	star, ok := receiver.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}

	identifier, ok := star.X.(*ast.Ident)

	return ok && identifier.Name == "Client"
}

var (
	goFencePattern = regexp.MustCompile("(?s)```go\n(.*?)```")
	apiRefPattern  = regexp.MustCompile(`\b(?:client|wise)\.([A-Z][A-Za-z0-9]*)`)
)

func symbolsReferencedByREADME(t *testing.T) map[string]bool {
	t.Helper()

	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}

	referenced := make(map[string]bool)

	for _, fence := range goFencePattern.FindAllStringSubmatch(string(readme), -1) {
		for _, match := range apiRefPattern.FindAllStringSubmatch(fence[1], -1) {
			referenced[match[1]] = true
		}
	}

	return referenced
}
