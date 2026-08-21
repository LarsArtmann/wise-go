package wise

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
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

	fset := token.NewFileSet()

	packages, err := parser.ParseDir(fset, ".", func(fileInfo fs.FileInfo) bool {
		return !strings.HasSuffix(fileInfo.Name(), "_test.go")
	}, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse package sources: %v", err)
	}

	exported := make(map[string]bool)

	for _, pkg := range packages {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				switch node := decl.(type) {
				case *ast.FuncDecl:
					if !node.Name.IsExported() {
						continue
					}

					if node.Recv == nil {
						exported[node.Name.Name] = true
						continue
					}

					if receiverIsClient(node.Recv) {
						exported[node.Name.Name] = true
					}
				case *ast.GenDecl:
					for _, spec := range node.Specs {
						switch spec := spec.(type) {
						case *ast.TypeSpec:
							exported[spec.Name.Name] = spec.Name.IsExported()
						case *ast.ValueSpec:
							for _, name := range spec.Names {
								exported[name.Name] = name.IsExported()
							}
						}
					}
				}
			}
		}
	}

	return exported
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
