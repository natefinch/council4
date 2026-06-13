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

// This file is the issue #418 enforcement gate. It proves, by static analysis,
// that the compiler (cardgen/oracle/compiler) performs NO semantic
// interpretation of Oracle source text or tokens, and that cardgen lowering
// interprets Oracle wording only in the explicitly classified, individually
// justified ways recorded in loweringTextUseAllowlist.
//
// The architecture contract is: the parser owns Oracle vocabulary, spelling, and
// grammar and emits source-spanned typed syntax; the compiler mechanically maps
// that syntax; lowering mechanically consumes compiler semantics. Source text
// may survive into the compiler/lowering layers only as genuine literal values,
// rendering/diagnostic strings, or exact source-span/consumption accounting —
// never to derive meaning.
//
// The analyzer flags a function as performing Oracle-text interpretation when it
// applies a string-inspection operation to Oracle wording. "Oracle wording" is
// tracked syntactically as any value that flows from a `.Text` or `.Event`
// field (the rendered Oracle text carried by parser and compiler nodes and by
// shared.Token), through text transforms (strings.ToLower and friends), slices,
// concatenation, or shared.NormalizedWords. A "string-inspection operation" is a
// predicate/search/split call from the strings package, an equality comparison
// against a string literal, a switch on Oracle wording with string-literal
// cases, a regexp call, or shared.NormalizedWords. Pure assignment of Oracle
// text into a rendered Text field, passing it to a diagnostic constructor, or
// writing it to a strings.Builder are NOT interpretation and are not flagged.

// inspectionStringsFuncs are the strings-package functions that read meaning out
// of their string arguments. strings.Builder methods and pure transforms
// (ToLower, TrimSpace, ...) are intentionally excluded: they reshape or render
// text without deriving semantics from it.
var inspectionStringsFuncs = map[string]bool{
	"Contains": true, "ContainsAny": true, "ContainsRune": true,
	"EqualFold": true, "HasPrefix": true, "HasSuffix": true,
	"Index": true, "IndexAny": true, "IndexByte": true, "IndexRune": true,
	"LastIndex": true, "LastIndexAny": true, "Count": true,
	"Split": true, "SplitN": true, "SplitAfter": true, "SplitAfterN": true,
	"Fields": true, "FieldsFunc": true,
	"Cut": true, "CutPrefix": true, "CutSuffix": true,
	"TrimPrefix": true, "TrimSuffix": true,
}

// textTransformFuncs are strings-package calls that produce a transformed copy
// of their input. They propagate Oracle-text taint (their result is still Oracle
// wording) but are not themselves interpretation.
var textTransformFuncs = map[string]bool{
	"ToLower": true, "ToUpper": true, "ToTitle": true, "Title": true,
	"TrimSpace": true, "Trim": true, "TrimLeft": true, "TrimRight": true,
	"TrimPrefix": true, "TrimSuffix": true, "Replace": true, "ReplaceAll": true,
	"Map": true, "Clone": true,
}

// textInterpretationSite is one flagged Oracle-text interpretation occurrence.
type textInterpretationSite struct {
	File string
	Func string
	Line int
	Kind string
	Expr string
}

// allowedTextUse records one function that is permitted to read Oracle wording,
// with the category and justification that makes it allowed. Every flagged site
// in cardgen lowering must map to one of these, and every entry must match at
// least one real flagged site (no stale exemptions).
type allowedTextUse struct {
	File          string
	Func          string
	Category      string
	Justification string
}

// loweringTextUseAllowlist is the complete, justified classification of every
// place cardgen lowering reads Oracle wording. There are exactly two categories:
// diagnostics (which never change whether a card is supported or how it behaves)
// and rendering (which only decides whether to emit the retained source-text
// comment). Nothing here derives game meaning from Oracle wording.
var loweringTextUseAllowlist = []allowedTextUse{
	{
		File:     "lower.go",
		Func:     "triggerPatternCapabilityDetail",
		Category: "diagnostic",
		Justification: "Selects a specific unsupported-reason message by scanning the rendered " +
			"trigger event/condition wording. Typed trigger-pattern lowering has already failed " +
			"closed before this runs, so the card is unsupported regardless and no supported/" +
			"unsupported outcome or generated behavior depends on the message. Diagnostic-only.",
	},
	{
		File:     "static_declaration.go",
		Func:     "lowerStaticDeclarations",
		Category: "diagnostic",
		Justification: "On the already-failed (!ok) static-declaration path, refines the unsupported " +
			"detail message by checking the rendered ability wording. The declaration is " +
			"unrepresentable regardless; the wording check only sharpens the message. Diagnostic-only.",
	},
	{
		File:     "static_declaration.go",
		Func:     "lowerStaticDeclarationBlocker",
		Category: "diagnostic",
		Justification: "Chooses the unsupported-static-declaration detail message for an ability the " +
			"compiler already flagged as blocked. The rendered-wording check only affects the message " +
			"text, never support or behavior. Diagnostic-only.",
	},
	{
		File:     "render.go",
		Func:     "(Renderer).renderStaticAbility",
		Category: "rendering",
		Justification: "Emits the retained Oracle source-text comment only when the rendered Text is " +
			"non-empty. An emptiness check on a rendering field, not Oracle-wording interpretation.",
	},
	{
		File:          "render.go",
		Func:          "(Renderer).renderActivatedAbility",
		Category:      "rendering",
		Justification: "Empty-checks the rendered Text before emitting the retained source-text comment.",
	},
	{
		File:          "render.go",
		Func:          "(Renderer).renderManaAbility",
		Category:      "rendering",
		Justification: "Empty-checks the rendered Text before emitting the retained source-text comment.",
	},
	{
		File:          "render.go",
		Func:          "(Renderer).renderTriggeredAbility",
		Category:      "rendering",
		Justification: "Empty-checks the rendered Text before emitting the retained source-text comment.",
	},
	{
		File:          "render.go",
		Func:          "(Renderer).renderLoyaltyAbility",
		Category:      "rendering",
		Justification: "Empty-checks the rendered Text before emitting the retained source-text comment.",
	},
	{
		File:          "render.go",
		Func:          "(Renderer).renderControllerControlsCondition",
		Category:      "rendering",
		Justification: "Empty-checks the rendered condition Text before emitting the retained source-text comment.",
	},
	{
		File:          "render.go",
		Func:          "(Renderer).renderMode",
		Category:      "rendering",
		Justification: "Empty-checks the rendered mode Text before emitting the retained source-text comment.",
	},
	{
		File:          "render.go",
		Func:          "(Renderer).renderAdditionalCosts",
		Category:      "rendering",
		Justification: "Empty-checks the rendered additional-cost Text before emitting the retained source-text comment.",
	},
	{
		File:          "render.go",
		Func:          "renderAdditional",
		Category:      "rendering",
		Justification: "Empty-checks the rendered additional-cost Text before emitting the retained source-text comment.",
	},
}

// TestCompilerIsTextBlind proves the compiler package performs no semantic
// interpretation of Oracle source text or tokens. The allowlist is empty: the
// compiler must consume typed parser syntax mechanically and may only copy
// Oracle text into rendering/diagnostic Text fields or write it to a
// strings.Builder, neither of which the analyzer flags.
func TestCompilerIsTextBlind(t *testing.T) {
	t.Parallel()
	sites := analyzeTextInterpretation(t, filepath.Join("oracle", "compiler"))
	if len(sites) != 0 {
		for _, s := range sites {
			t.Errorf("compiler must be text-blind, but %s:%d in %s interprets Oracle text: %s (%s)",
				s.File, s.Line, s.Func, s.Expr, s.Kind)
		}
	}
}

// TestLoweringTextInterpretationIsAllowlisted proves cardgen lowering interprets
// Oracle wording only in the explicitly justified ways recorded above. A new,
// unclassified interpretation site fails the test; so does a stale allowlist
// entry that no longer matches any site, keeping the allowlist tight.
func TestLoweringTextInterpretationIsAllowlisted(t *testing.T) {
	t.Parallel()
	sites := analyzeTextInterpretation(t, ".")

	allowed := map[string]allowedTextUse{}
	for _, a := range loweringTextUseAllowlist {
		allowed[a.File+"::"+a.Func] = a
	}

	matched := map[string]bool{}
	for _, s := range sites {
		key := s.File + "::" + s.Func
		entry, ok := allowed[key]
		if !ok {
			t.Errorf("unallowlisted Oracle-text interpretation in lowering: %s:%d %s interprets %q (%s)\n"+
				"  migrate it to typed parser/compiler syntax, or add a justified entry to loweringTextUseAllowlist",
				s.File, s.Line, s.Func, s.Expr, s.Kind)
			continue
		}
		matched[key] = true
		if strings.TrimSpace(entry.Justification) == "" || strings.TrimSpace(entry.Category) == "" {
			t.Errorf("allowlist entry %s is missing a category or justification", key)
		}
	}
	for key := range allowed {
		if !matched[key] {
			t.Errorf("stale allowlist entry %q matches no Oracle-text interpretation site; remove it", key)
		}
	}
}

// TestEnforcementDetectsViolations is a meta-test: it confirms the analyzer
// actually fires on representative reintroduced violations (a token-spelling
// comparison, a strings.Contains over rendered wording, and shared.NormalizedWords),
// and does not fire on allowed shapes (copying Text into a field, writing it to a
// builder, comparing a typed enum field).
func TestEnforcementDetectsViolations(t *testing.T) {
	t.Parallel()
	violating := `package sample

import (
	"strings"

	"x/shared"
)

func equalWord(token shared.Token, word string) bool {
	return token.Kind == shared.Word && strings.EqualFold(token.Text, word)
}

func detect(ability struct{ Text string }) bool {
	lowered := strings.ToLower(ability.Text)
	return strings.Contains(lowered, "if able")
}

func words(tokens []shared.Token) []string {
	return shared.NormalizedWords(tokens)
}
`
	sites := analyzeSource(t, "violating.go", violating)
	wantFns := map[string]bool{"equalWord": false, "detect": false, "words": false}
	for _, s := range sites {
		if _, ok := wantFns[s.Func]; ok {
			wantFns[s.Func] = true
		}
	}
	for fn, found := range wantFns {
		if !found {
			t.Errorf("analyzer failed to flag a known violation in %s", fn)
		}
	}

	allowedShapes := `package sample

import (
	"strings"

	"x/shared"
)

type compiled struct {
	Text  string
	Event int
}

func render(c compiled, b *strings.Builder) {
	_, _ = b.WriteString(c.Text)
}

func copyText(c compiled) compiled {
	return compiled{Text: c.Text}
}

func typedEnum(c compiled) bool {
	return c.Event == 3
}
`
	if got := analyzeSource(t, "allowed.go", allowedShapes); len(got) != 0 {
		for _, s := range got {
			t.Errorf("analyzer wrongly flagged an allowed shape: %s in %s: %s", s.Kind, s.Func, s.Expr)
		}
	}
}

// analyzeTextInterpretation parses every non-test .go file in dir (relative to
// the cardgen package directory) and returns the Oracle-text interpretation
// sites the analyzer finds.
func analyzeTextInterpretation(t *testing.T, dir string) []textInterpretationSite {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	var sites []textInterpretationSite
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
		sites = append(sites, analyzeSource(t, rel, string(src))...)
	}
	slices.SortFunc(sites, func(a, b textInterpretationSite) int {
		return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line))
	})
	return sites
}

// analyzeSource parses one Go source file and returns its Oracle-text
// interpretation sites.
func analyzeSource(t *testing.T, rel, src string) []textInterpretationSite {
	t.Helper()
	fset := token.NewFileSet()
	file, err := goparser.ParseFile(fset, rel, src, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", rel, err)
	}
	var sites []textInterpretationSite
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		name := funcName(fn)
		tainted := oracleTaintedLocals(fn)
		ast.Inspect(fn, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
					if pkg, ok := sel.X.(*ast.Ident); ok {
						switch {
						case pkg.Name == "strings" && inspectionStringsFuncs[sel.Sel.Name] && anyOracleArg(node.Args, tainted):
							sites = append(sites, textInterpretationSite{rel, name, fset.Position(node.Pos()).Line, "strings." + sel.Sel.Name, exprString(node)})
						case pkg.Name == "regexp":
							sites = append(sites, textInterpretationSite{rel, name, fset.Position(node.Pos()).Line, "regexp." + sel.Sel.Name, exprString(node)})
						case pkg.Name == "shared" && sel.Sel.Name == "NormalizedWords":
							sites = append(sites, textInterpretationSite{rel, name, fset.Position(node.Pos()).Line, "shared.NormalizedWords", exprString(node)})
						default:
						}
					}
				}
			case *ast.BinaryExpr:
				if node.Op == token.EQL || node.Op == token.NEQ {
					ox := oracleTextValued(node.X, tainted)
					oy := oracleTextValued(node.Y, tainted)
					if (ox || oy) && (isStringLit(node.X) || isStringLit(node.Y)) {
						sites = append(sites, textInterpretationSite{rel, name, fset.Position(node.Pos()).Line, "cmp", exprString(node)})
					}
				}
			case *ast.SwitchStmt:
				if node.Tag != nil && oracleTextValued(node.Tag, tainted) && switchHasStringLit(node) {
					sites = append(sites, textInterpretationSite{rel, name, fset.Position(node.Pos()).Line, "switch", exprString(node.Tag)})
				}
			default:
			}
			return true
		})
	}
	return sites
}

func funcName(fn *ast.FuncDecl) string {
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		return "(" + exprString(fn.Recv.List[0].Type) + ")." + fn.Name.Name
	}
	return fn.Name.Name
}

// oracleTextValued reports whether expr evaluates to Oracle source wording: a
// .Text/.Event field, a tainted local, a slice/index of such, a string transform
// of such, shared.NormalizedWords of such, or a concatenation involving such. It
// deliberately does not treat arbitrary function-call results or struct literals
// as Oracle text, so taint stays on the wording rather than leaking onto domain
// values such as produced diagnostics.
func oracleTextValued(expr ast.Expr, tainted map[string]bool) bool {
	switch x := expr.(type) {
	case *ast.ParenExpr:
		return oracleTextValued(x.X, tainted)
	case *ast.Ident:
		return tainted[x.Name]
	case *ast.SelectorExpr:
		return x.Sel.Name == "Text" || x.Sel.Name == "Event"
	case *ast.IndexExpr:
		return oracleTextValued(x.X, tainted)
	case *ast.SliceExpr:
		return oracleTextValued(x.X, tainted)
	case *ast.BinaryExpr:
		if x.Op == token.ADD {
			return oracleTextValued(x.X, tainted) || oracleTextValued(x.Y, tainted)
		}
	case *ast.CallExpr:
		sel, ok := x.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return false
		}
		transform := pkg.Name == "strings" && textTransformFuncs[sel.Sel.Name]
		normalized := pkg.Name == "shared" && sel.Sel.Name == "NormalizedWords"
		if transform || normalized {
			for _, a := range x.Args {
				if oracleTextValued(a, tainted) {
					return true
				}
			}
		}
	default:
	}
	return false
}

// oracleTaintedLocals computes, to a fixpoint, the set of local identifiers in fn
// that hold Oracle wording.
func oracleTaintedLocals(fn *ast.FuncDecl) map[string]bool {
	tainted := map[string]bool{}
	for changed := true; changed; {
		changed = false
		ast.Inspect(fn, func(n ast.Node) bool {
			switch s := n.(type) {
			case *ast.AssignStmt:
				if len(s.Lhs) == len(s.Rhs) {
					for i := range s.Rhs {
						if id, ok := s.Lhs[i].(*ast.Ident); ok && id.Name != "_" && !tainted[id.Name] && oracleTextValued(s.Rhs[i], tainted) {
							tainted[id.Name] = true
							changed = true
						}
					}
				}
			case *ast.RangeStmt:
				if id, ok := s.Value.(*ast.Ident); ok && id.Name != "_" && !tainted[id.Name] && oracleTextValued(s.X, tainted) {
					tainted[id.Name] = true
					changed = true
				}
			case *ast.ValueSpec:
				if len(s.Names) == len(s.Values) {
					for i := range s.Values {
						if s.Names[i].Name != "_" && !tainted[s.Names[i].Name] && oracleTextValued(s.Values[i], tainted) {
							tainted[s.Names[i].Name] = true
							changed = true
						}
					}
				}
			default:
			}
			return true
		})
	}
	return tainted
}

func anyOracleArg(args []ast.Expr, tainted map[string]bool) bool {
	for _, a := range args {
		if oracleTextValued(a, tainted) {
			return true
		}
	}
	return false
}

func isStringLit(n ast.Node) bool {
	lit, ok := n.(*ast.BasicLit)
	return ok && lit.Kind == token.STRING
}

func switchHasStringLit(sw *ast.SwitchStmt) bool {
	for _, stmt := range sw.Body.List {
		cc, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		for _, e := range cc.List {
			if isStringLit(e) {
				return true
			}
		}
	}
	return false
}

func exprString(n ast.Node) string {
	switch x := n.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return exprString(x.X) + "." + x.Sel.Name
	case *ast.StarExpr:
		return "*" + exprString(x.X)
	case *ast.CallExpr:
		args := make([]string, len(x.Args))
		for i, a := range x.Args {
			args[i] = exprString(a)
		}
		return exprString(x.Fun) + "(" + strings.Join(args, ", ") + ")"
	case *ast.BasicLit:
		return x.Value
	case *ast.BinaryExpr:
		return exprString(x.X) + " " + x.Op.String() + " " + exprString(x.Y)
	case *ast.IndexExpr:
		return exprString(x.X) + "[" + exprString(x.Index) + "]"
	case *ast.ParenExpr:
		return "(" + exprString(x.X) + ")"
	default:
		return "<expr>"
	}
}
