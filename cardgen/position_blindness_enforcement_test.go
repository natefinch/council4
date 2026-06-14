package cardgen

import (
	"cmp"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// This file is the issue #428 enforcement gate. It proves, by static analysis,
// that the compiler (cardgen/oracle/compiler) performs NO positional reasoning
// over Oracle source-span byte offsets: it never derives node identity,
// containment, or ordering from raw positions.
//
// The architecture contract is: the parser owns all positional reasoning and
// emits it as typed relationships (stable NodeIDs for identity and dense
// source-order ranks for ordering and containment); the compiler consumes those
// typed relationships mechanically. Source spans survive into the compiler only
// as pass-through values for diagnostics (unsupportedDiagnostic) and for
// lowering's retained-text rendering and source-consumption accounting — never
// to compute meaning.
//
// The analyzer flags two position-reasoning shapes in compiler non-test code:
//   - any read of a span boundary's byte Offset (a `.Offset` field selector),
//     which is how every offset comparison and span-containment test was spelled;
//   - any equality/inequality comparison whose operand is a `*Span` field
//     selector, which is how span values were used as node identity.
//
// Passing a span to a diagnostic, returning or constructing a span, copying it
// into a field, and comparing typed source-order ranks (the `.Order` fields) are
// NOT positional reasoning and are not flagged.

// spanReasoningSite is one flagged span-offset/identity reasoning occurrence.
type spanReasoningSite struct {
	File string
	Func string
	Line int
	Kind string
	Expr string
}

// TestCompilerIsPositionBlind proves the compiler package performs no positional
// reasoning over source-span byte offsets. The allowlist is empty: the compiler
// must consume the parser's typed identity (NodeID) and source-order (Order)
// relationships and may only pass spans through to diagnostics and lowering.
func TestCompilerIsPositionBlind(t *testing.T) {
	t.Parallel()
	sites := analyzeSpanReasoning(t, filepath.Join("oracle", "compiler"))
	if len(sites) != 0 {
		for _, s := range sites {
			t.Errorf("compiler must be position-blind, but %s:%d in %s reasons by source position: %s (%s)\n"+
				"  consume the parser's typed NodeID/Order relationships instead of byte offsets or span identity",
				s.File, s.Line, s.Func, s.Expr, s.Kind)
		}
	}
}

// TestPositionEnforcementDetectsViolations is a meta-test: it confirms the
// analyzer fires on representative reintroduced violations (an offset comparison,
// a span-containment offset test, and span-equality identity) and does not fire
// on allowed shapes (passing a span to a diagnostic, returning/copying a span,
// and comparing typed source-order ranks).
func TestPositionEnforcementDetectsViolations(t *testing.T) {
	t.Parallel()
	violating := `package sample

func follows(reference, verb spanLike) bool {
	return reference.Start.Offset >= verb.End.Offset
}

func contains(outer, inner spanLike) bool {
	return outer.Start.Offset <= inner.Start.Offset && outer.End.Offset >= inner.End.Offset
}

func sameNode(a, b nodeLike) bool {
	return a.Span == b.Span
}
`
	sites := analyzeSpanReasoningSource(t, "violating.go", violating)
	wantFns := map[string]bool{"follows": false, "contains": false, "sameNode": false}
	for _, s := range sites {
		if _, ok := wantFns[s.Func]; ok {
			wantFns[s.Func] = true
		}
	}
	for fn, found := range wantFns {
		if !found {
			t.Errorf("position analyzer failed to flag a known violation in %s", fn)
		}
	}

	allowedShapes := `package sample

func diagnose(span spanLike) diagnostic {
	return unsupportedDiagnostic(span, "unsupported")
}

func passthrough(node nodeLike) spanLike {
	return node.Span
}

func ordered(a, b nodeLike) bool {
	return a.Order.Start < b.Order.Start && a.Order.Contains(b.Order)
}

func identity(a, b nodeLike) bool {
	return a.NodeID == b.NodeID
}
`
	if got := analyzeSpanReasoningSource(t, "allowed.go", allowedShapes); len(got) != 0 {
		for _, s := range got {
			t.Errorf("position analyzer wrongly flagged an allowed shape: %s in %s: %s", s.Kind, s.Func, s.Expr)
		}
	}
}

// analyzeSpanReasoning parses every non-test .go file in dir (relative to the
// cardgen package directory) and returns the span-offset/identity reasoning
// sites the analyzer finds.
func analyzeSpanReasoning(t *testing.T, dir string) []spanReasoningSite {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	var sites []spanReasoningSite
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		rel := name
		if dir != "." {
			rel = filepath.ToSlash(filepath.Join(dir, name))
		}
		sites = append(sites, analyzeSpanReasoningSource(t, rel, string(src))...)
	}
	slices.SortFunc(sites, func(a, b spanReasoningSite) int {
		return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line))
	})
	return sites
}

// analyzeSpanReasoningSource parses one Go source file and returns its
// span-offset/identity reasoning sites.
func analyzeSpanReasoningSource(t *testing.T, rel, src string) []spanReasoningSite {
	t.Helper()
	fset := token.NewFileSet()
	file, err := goparser.ParseFile(fset, rel, src, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", rel, err)
	}
	var sites []spanReasoningSite
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		name := funcName(fn)
		ast.Inspect(fn, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.SelectorExpr:
				if node.Sel.Name == "Offset" {
					sites = append(sites, spanReasoningSite{rel, name, fset.Position(node.Pos()).Line, "offset", exprString(node)})
				}
			case *ast.BinaryExpr:
				if node.Op == token.EQL || node.Op == token.NEQ {
					if isSpanSelector(node.X) || isSpanSelector(node.Y) {
						sites = append(sites, spanReasoningSite{rel, name, fset.Position(node.Pos()).Line, "span-identity", exprString(node)})
					}
				}
			default:
			}
			return true
		})
	}
	return sites
}

// isSpanSelector reports whether expr is a field selector whose name is or ends
// with "Span" (Span, VerbSpan, SubjectSpan, ...): the span-valued node fields the
// compiler must never compare for identity.
func isSpanSelector(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	return ok && strings.HasSuffix(sel.Sel.Name, "Span")
}
