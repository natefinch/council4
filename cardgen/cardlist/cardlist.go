// Package cardlist scans a generated card letter-package directory and renders
// its cards.go, which registers each card as a cardset.Entry (a name paired with
// its constructor) for lazy loading.
//
// A card is a package-level var holding its builder as a value (var X = newX,
// where func newX() *game.CardDef builds it), so the CardDef is constructed only
// when invoked. Token definitions instead hold the built CardDef (var xToken =
// newXToken(), or a direct literal), because card builders reference them by
// pointer. scan uses that distinction — a card's exported var initializer is a
// bare builder identifier, a token's is a call or a composite literal — to list
// the cards and, for a token package, list its eager defs instead.
package cardlist

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// cardsetImportPath is the package providing cardset.Entry.
const cardsetImportPath = "github.com/natefinch/council4/mtg/cards/cardset"

// Entry is a card's canonical name paired with the package var that constructs
// it (a func() *game.CardDef value).
type Entry struct {
	Name    string // printed card name (front face), from CardFace.Name
	Builder string // exported var holding the builder, e.g. LightningBolt
}

// Generate produces the cards.go source for a letter package directory, choosing
// the registration form from the package's contents. A token package (its
// exported defs are eager — a var holding a built CardDef) lists them as a
// []*game.CardDef. A card package (its exported cards are vars holding a builder
// value) lists cardset.Entry constructors so the registry builds them lazily.
// generator names the tool in the header.
func Generate(dir, pkgName, generator string) ([]byte, error) {
	cards, eager, err := scan(dir)
	if err != nil {
		return nil, err
	}
	if len(eager) > 0 {
		return renderEager(pkgName, generator, eager)
	}
	return render(pkgName, generator, cards)
}

// scan parses a letter package and separates its exported CardDef registrations:
// cards (a var holding a builder value) and eager defs (a var holding a built
// CardDef, i.e. a token package). Card entries are sorted by builder.
func scan(dir string) (cards []Entry, eager []string, err error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	fset := token.NewFileSet()
	var parsed []*ast.File
	for _, file := range files {
		name := file.Name()
		if file.IsDir() || !strings.HasSuffix(name, ".go") ||
			name == "cards.go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, parseErr := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("parsing %s: %w", name, parseErr)
		}
		parsed = append(parsed, f)
	}

	builders := builderBodies(parsed)
	for _, f := range parsed {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok || len(value.Names) != 1 || len(value.Values) != 1 {
					continue
				}
				name := value.Names[0].Name
				if name == "" || name[0] < 'A' || name[0] > 'Z' {
					continue // token defs co-located in card files are unexported
				}
				if body := cardBody(value.Values[0], builders); body != nil {
					cardName, found := nameInBody(body)
					if !found {
						return nil, nil, fmt.Errorf("no CardFace.Name found for %s", name)
					}
					cards = append(cards, Entry{Name: cardName, Builder: name})
					continue
				}
				if isEagerDef(value.Values[0], builders) {
					eager = append(eager, name)
				}
			}
		}
	}
	slices.SortFunc(cards, func(a, b Entry) int { return strings.Compare(a.Builder, b.Builder) })
	return cards, eager, nil
}

// cardBody returns the func body to read a card's Name from when expr is a lazy
// card registration — a bare builder identifier (var X = newX) or a func literal
// (var X = func() *game.CardDef { ... }) — or nil otherwise.
func cardBody(expr ast.Expr, builders map[string]*ast.FuncDecl) ast.Node {
	if ident, ok := expr.(*ast.Ident); ok {
		if fn, isBuilder := builders[ident.Name]; isBuilder {
			return fn.Body
		}
		return nil
	}
	if lit, ok := expr.(*ast.FuncLit); ok && returnsCardDefPointer(lit.Type) {
		return lit.Body
	}
	return nil
}

// isEagerDef reports whether expr is an eager CardDef registration — a builder
// call (var X = newX()) or a direct composite literal (var X = &game.CardDef{}) —
// which marks a token def.
func isEagerDef(expr ast.Expr, builders map[string]*ast.FuncDecl) bool {
	if call, ok := expr.(*ast.CallExpr); ok {
		if ident, ok := call.Fun.(*ast.Ident); ok {
			_, isBuilder := builders[ident.Name]
			return isBuilder
		}
		return false
	}
	return compositeOf(expr) != nil
}

// builderBodies returns each no-arg *game.CardDef builder func by name.
func builderBodies(files []*ast.File) map[string]*ast.FuncDecl {
	builders := map[string]*ast.FuncDecl{}
	for _, f := range files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && fn.Recv == nil && fn.Name != nil && fn.Body != nil && returnsCardDefPointer(fn.Type) {
				builders[fn.Name.Name] = fn
			}
		}
	}
	return builders
}

// render returns gofmt-formatted cards.go source registering entries lazily.
func render(pkgName, generator string, entries []Entry) ([]byte, error) {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "// Code generated by %s; DO NOT EDIT.\n\n", generator)
	_, _ = fmt.Fprintf(&b, "package %s\n\n", pkgName)
	_, _ = fmt.Fprintf(&b, "import %q\n\n", cardsetImportPath)
	_, _ = b.WriteString("// Cards lists all card definitions in this package, each paired with a\n")
	_, _ = b.WriteString("// constructor so the registry can build them lazily.\n")
	_, _ = b.WriteString("var Cards = []cardset.Entry{\n")
	for _, entry := range entries {
		_, _ = fmt.Fprintf(&b, "\t{Name: %q, New: %s},\n", entry.Name, entry.Builder)
	}
	_, _ = b.WriteString("}\n")
	return format.Source([]byte(b.String()))
}

// renderEager returns gofmt-formatted cards.go source listing eager,
// pre-constructed CardDef package vars as a []*game.CardDef, for token packages.
func renderEager(pkgName, generator string, varNames []string) ([]byte, error) {
	slices.Sort(varNames)
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "// Code generated by %s; DO NOT EDIT.\n\n", generator)
	_, _ = fmt.Fprintf(&b, "package %s\n\n", pkgName)
	_, _ = b.WriteString("import \"github.com/natefinch/council4/mtg/game\"\n\n")
	_, _ = b.WriteString("// Cards lists all card definitions in this package.\n")
	_, _ = b.WriteString("var Cards = []*game.CardDef{\n")
	for _, name := range varNames {
		_, _ = fmt.Fprintf(&b, "\t%s,\n", name)
	}
	_, _ = b.WriteString("}\n")
	return format.Source([]byte(b.String()))
}

func returnsCardDefPointer(fn *ast.FuncType) bool {
	if fn.Results == nil || len(fn.Results.List) != 1 {
		return false
	}
	star, ok := fn.Results.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "game" && sel.Sel.Name == "CardDef"
}

// nameInBody finds the first &game.CardDef{...} (or game.CardDef{...}) literal in
// node and returns its CardFace.Name.
func nameInBody(node ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	var lit *ast.CompositeLit
	ast.Inspect(node, func(n ast.Node) bool {
		if lit != nil {
			return false
		}
		if cl, ok := n.(*ast.CompositeLit); ok && isSelector(cl.Type, "game", "CardDef") {
			lit = cl
			return false
		}
		return true
	})
	if lit == nil {
		return "", false
	}
	return nameFromCardDef(lit)
}

// compositeOf returns the composite literal for &game.CardDef{...} or
// game.CardDef{...}, or nil if expr is neither.
func compositeOf(expr ast.Expr) *ast.CompositeLit {
	if unary, ok := expr.(*ast.UnaryExpr); ok {
		expr = unary.X
	}
	lit, ok := expr.(*ast.CompositeLit)
	if !ok || !isSelector(lit.Type, "game", "CardDef") {
		return nil
	}
	return lit
}

// nameFromCardDef extracts the Name from a game.CardDef composite: it looks in
// the CardFace field's literal, falling back to a top-level Name key.
func nameFromCardDef(lit *ast.CompositeLit) (string, bool) {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "CardFace" {
			continue
		}
		if face, ok := kv.Value.(*ast.CompositeLit); ok {
			if name, ok := stringField(face, "Name"); ok {
				return name, true
			}
		}
	}
	return stringField(lit, "Name")
}

func stringField(lit *ast.CompositeLit, field string) (string, bool) {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != field {
			continue
		}
		basic, ok := kv.Value.(*ast.BasicLit)
		if !ok || basic.Kind != token.STRING {
			return "", false
		}
		unquoted, err := strconv.Unquote(basic.Value)
		if err != nil {
			return "", false
		}
		return unquoted, true
	}
	return "", false
}

func isSelector(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == pkg && sel.Sel.Name == name
}
