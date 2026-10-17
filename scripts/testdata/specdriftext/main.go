// Command specdriftext extracts every (method, path template) the SDK's
// client helpers call from root *.go files, replacing the regex extractor in
// scripts/spec_drift.py with a real Go AST walk. It prints a JSON array of
// {method, template, file, line, enclosing} and is consumed file-mode via
// `go run` so it stays outside the module (scripts/testdata/ is invisible to
// go build/test/lint/coverage).
//
// Parity contract with the regex extractor it replaced (2026-10-17): same
// call shapes, same normalization inputs. Shapes: c.get/getWithQuery/
// getWithQueryHeaders/getRaw/post/postWithHeaders/put/patch/delete with the
// path as the first argument after ctx (literal, fmt.Sprintf head, or an
// ident assigned earlier in the same function); fetchByID call sites (path
// passed through); c.request with an http.Method literal or with a method
// parameter resolved against the enclosing function's repo-wide call sites
// (one entry per distinct method, UNRESOLVED when nothing binds).
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type endpoint struct {
	Method    string `json:"method"`
	Template  string `json:"template"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Enclosing string `json:"enclosing"`
}

var verbMethods = map[string]string{
	"get": "GET", "getWithQuery": "GET", "getWithQueryHeaders": "GET",
	"getRaw": "GET", "post": "POST", "postWithHeaders": "POST",
	"put": "PUT", "patch": "PATCH", "delete": "DELETE",
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: specdriftext <repo-root>")
		os.Exit(2)
	}

	fset := token.NewFileSet()
	parsed := map[string]*ast.File{}
	decls := map[string]*ast.FuncDecl{} // function name -> decl (repo-wide)

	files, err := filepath.Glob(filepath.Join(os.Args[1], "*.go"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "glob: %v\n", err)
		os.Exit(2)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", f, err)
			os.Exit(2)
		}
		parsed[f] = file
		ast.Inspect(file, func(n ast.Node) bool {
			if fd, ok := n.(*ast.FuncDecl); ok && fd.Name != nil {
				decls[fd.Name.Name] = fd
			}
			return true
		})
	}

	var out []endpoint
	// symbolic: (file, enclosing func name, param name, call line) -> endpoint pending method binding
	type pendingSym struct {
		endpoint  endpoint
		paramName string
	}
	var symbolic []pendingSym

	for file, tree := range parsed {
		base := filepath.Base(file)
		ast.Inspect(tree, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			callee := calleeName(call.Fun)
			enclosing, encDecl := enclosingFunc(tree, call)

			switch {
			case callee == "request" && isReceiver(call.Fun, "c"):
				if len(call.Args) < 3 {
					return true
				}
				ep := endpoint{File: base, Line: fset.Position(call.Lparen).Line, Enclosing: enclosing}
				ep.Template = resolvePath(call.Args[2], encDecl, fset)
				if ep.Template == "" {
					return true
				}
				switch m := call.Args[1].(type) {
				case *ast.SelectorExpr:
					if pkg, ok := m.X.(*ast.Ident); ok && pkg.Name == "http" && strings.HasPrefix(m.Sel.Name, "Method") {
						ep.Method = strings.ToUpper(strings.TrimPrefix(m.Sel.Name, "Method"))
						out = append(out, ep)
					}
				case *ast.Ident:
					ep.Method = ""
					symbolic = append(symbolic, pendingSym{endpoint: ep, paramName: m.Name})
				}
			case verbMethods[callee] != "" && (isReceiver(call.Fun, "c") || isReceiver(call.Fun, "client")):
				if len(call.Args) < 2 {
					return true
				}
				ep := endpoint{
					Method:    verbMethods[callee],
					File:      base,
					Line:      fset.Position(call.Lparen).Line,
					Enclosing: enclosing,
				}
				ep.Template = resolvePath(call.Args[1], encDecl, fset)
				if ep.Template != "" {
					out = append(out, ep)
				}
			case callee == "fetchByID":
				for _, arg := range call.Args {
					t := resolvePath(arg, encDecl, fset)
					if strings.HasPrefix(t, "/") {
						out = append(out, endpoint{
							Method: "GET", Template: t, File: base,
							Line:      fset.Position(call.Lparen).Line,
							Enclosing: enclosing,
						})
						break
					}
				}
			}
			return true
		})
	}

	// Bind symbolic c.request methods: find repo-wide call sites of the
	// enclosing function and read the http.Method literal passed in the
	// parameter's position.
	for _, sym := range symbolic {
		methods := map[string]bool{}
		if decl, ok := decls[sym.endpoint.Enclosing]; ok {
			idx := paramIndex(decl, sym.paramName)
			if idx >= 0 {
				for _, tree := range parsed {
					ast.Inspect(tree, func(n ast.Node) bool {
						call, ok := n.(*ast.CallExpr)
						if !ok || calleeName(call.Fun) != sym.endpoint.Enclosing || len(call.Args) <= idx {
							return true
						}
						if sel, ok := call.Args[idx].(*ast.SelectorExpr); ok {
							if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "http" && strings.HasPrefix(sel.Sel.Name, "Method") {
								methods[strings.ToUpper(strings.TrimPrefix(sel.Sel.Name, "Method"))] = true
							}
						}
						return true
					})
				}
			}
		}
		if len(methods) == 0 {
			sym.endpoint.Method = "UNRESOLVED"
			out = append(out, sym.endpoint)
			continue
		}
		for m := range methods {
			ep := sym.endpoint
			ep.Method = m
			out = append(out, ep)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Line < out[j].Line
	})
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "encode: %v\n", err)
		os.Exit(2)
	}
}

func calleeName(expr ast.Expr) string {
	switch fn := expr.(type) {
	case *ast.SelectorExpr:
		return fn.Sel.Name
	case *ast.Ident:
		return fn.Name
	}
	return ""
}

func isReceiver(fun ast.Expr, name string) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == name
}

func enclosingFunc(file *ast.File, call ast.Node) (string, *ast.FuncDecl) {
	var best *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		fd, ok := n.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			return true
		}
		if fd.Body.Pos() <= call.Pos() && call.End() <= fd.Body.End() {
			best = fd
		}
		return true
	})
	if best == nil {
		return "", nil
	}
	return best.Name.Name, best
}

func paramIndex(decl *ast.FuncDecl, name string) int {
	if decl.Type.Params == nil {
		return -1
	}
	idx := 0
	for _, field := range decl.Type.Params.List {
		for _, n := range field.Names {
			if n.Name == name {
				return idx
			}
			idx++
		}
		if len(field.Names) == 0 {
			idx++
		}
	}
	return -1
}

// resolvePath resolves a path expression: string literal, fmt.Sprintf head
// literal, or an identifier assigned earlier in the same function body.
func resolvePath(expr ast.Expr, decl *ast.FuncDecl, fset *token.FileSet) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			return strings.Trim(e.Value, "`\"")
		}
	case *ast.CallExpr:
		if calleeName(e.Fun) == "Sprintf" && len(e.Args) > 0 {
			return resolvePath(e.Args[0], decl, fset)
		}
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			return resolvePath(e.X, decl, fset)
		}
	case *ast.Ident:
		if decl == nil || decl.Body == nil {
			return ""
		}
		var found string
		ast.Inspect(decl.Body, func(n ast.Node) bool {
			if n != nil && n.End() > expr.Pos() {
				return n.Pos() <= expr.Pos()
			}
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for i, lhs := range assign.Lhs {
				id, ok := lhs.(*ast.Ident)
				if !ok || id.Name != e.Name || i >= len(assign.Rhs) {
					continue
				}
				if resolved := resolvePath(assign.Rhs[i], decl, fset); resolved != "" {
					found = resolved
				}
			}
			return true
		})
		return found
	case *ast.ParenExpr:
		return resolvePath(e.X, decl, fset)
	}
	return ""
}
