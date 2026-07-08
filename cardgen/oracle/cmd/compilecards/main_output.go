package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/cardgen/cardlist"
)

func writeSupported(root string, results []result) error {
	affected := make(map[string]bool)
	finalPaths := make(map[string]bool)
	generatedPrefixes := make(map[string]bool)
	tokenPrefixes := make(map[string]bool)
	for _, result := range results {
		if result.err != nil || len(result.diagnostics) > 0 {
			continue
		}
		finalPaths[result.relative] = true
		directory := filepath.Dir(result.relative)
		base := cardgen.CardNameToSafeFileName(result.card.Name)
		if result.card.Layout == "token" || result.card.Layout == "double_faced_token" {
			tokenPrefixes[filepath.Join(directory, base+"_")] = true
		} else {
			generatedPrefixes[filepath.Join(directory, base+"_scryfall")] = true
		}
	}
	for _, result := range results {
		if result.err != nil || len(result.diagnostics) > 0 || result.superseded == "" {
			continue
		}
		path := filepath.Join(root, result.superseded)
		err := os.Remove(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("removing superseded source for %s: %w", result.card.Name, err)
		}
		if err == nil {
			affected[filepath.Dir(path)] = true
		}
	}
	for prefix := range generatedPrefixes {
		matches, err := filepath.Glob(filepath.Join(root, prefix+"*.go"))
		if err != nil {
			return fmt.Errorf("matching generated identity paths for %s: %w", prefix, err)
		}
		for _, path := range matches {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return fmt.Errorf("resolving generated identity path %s: %w", path, err)
			}
			if finalPaths[relative] {
				continue
			}
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("removing obsolete generated identity path %s: %w", path, err)
			}
			affected[filepath.Dir(path)] = true
		}
	}
	for prefix := range tokenPrefixes {
		matches, err := filepath.Glob(filepath.Join(root, prefix+"*.go"))
		if err != nil {
			return fmt.Errorf("matching generated token identity paths for %s: %w", prefix, err)
		}
		for _, path := range matches {
			if !isTokenIdentityPath(path, filepath.Join(root, prefix)) {
				continue
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return fmt.Errorf("resolving generated token identity path %s: %w", path, err)
			}
			if finalPaths[relative] {
				continue
			}
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("removing obsolete generated token identity path %s: %w", path, err)
			}
			affected[filepath.Dir(path)] = true
		}
	}
	for _, result := range results {
		if result.exclusion != "" || result.err != nil || len(result.diagnostics) > 0 {
			continue
		}
		path := filepath.Join(root, result.relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return fmt.Errorf("creating package directory for %s: %w", result.card.Name, err)
		}
		if err := os.WriteFile(path, []byte(result.source), 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
		affected[filepath.Dir(path)] = true
	}
	directories := make([]string, 0, len(affected))
	for directory := range affected {
		directories = append(directories, directory)
	}
	slices.Sort(directories)
	for _, directory := range directories {
		if err := writeCardList(directory); err != nil {
			return err
		}
	}
	return writeTokenPackages(root, results)
}

func isTokenIdentityPath(path, prefix string) bool {
	suffix := strings.TrimSuffix(strings.TrimPrefix(path, prefix), ".go")
	if len(suffix) != 32 {
		return false
	}
	for _, r := range suffix {
		if !strings.ContainsRune("0123456789abcdef", r) {
			return false
		}
	}
	return true
}

func writeTokenPackages(root string, results []result) error {
	letters := make(map[string]bool)
	for _, result := range results {
		if result.err != nil || len(result.diagnostics) > 0 ||
			(result.card.Layout != "token" && result.card.Layout != "double_faced_token") {
			continue
		}
		letters[filepath.Base(filepath.Dir(result.relative))] = true
	}
	if len(letters) == 0 {
		return nil
	}
	tokenRoot := filepath.Join(root, "tokens")
	if err := os.MkdirAll(tokenRoot, 0o750); err != nil {
		return fmt.Errorf("creating token package: %w", err)
	}
	rootReadme := "# Tokens\n\n" +
		"Package `tokens` collects generated definitions for playable paper tokens. " +
		"Token definitions live in letter subpackages and use their complete Oracle ID " +
		"in filenames and Go identifiers so same-name tokens remain distinct.\n\n" +
		"Tokens are not included in `cards.Registry`. In the repository tree, use " +
		"`tokens.Cards` when all token definitions are needed.\n"
	if err := os.WriteFile(filepath.Join(tokenRoot, "README.md"), []byte(rootReadme), 0o600); err != nil {
		return fmt.Errorf("writing token package README: %w", err)
	}
	ordered := make([]string, 0, len(letters))
	for letter := range letters {
		ordered = append(ordered, letter)
	}
	slices.Sort(ordered)
	for _, letter := range ordered {
		doc := fmt.Sprintf(
			"// Package %s contains generated playable token definitions.\npackage %s\n\n"+
				"//go:generate go run github.com/natefinch/council4/cardgen/cmd/gencardlist\n",
			letter,
			letter,
		)
		docPath := filepath.Join(tokenRoot, letter, "doc.go")
		if err := os.WriteFile(docPath, []byte(doc), 0o600); err != nil {
			return fmt.Errorf("writing token letter package documentation: %w", err)
		}
		readme := fmt.Sprintf(
			"# %s tokens\n\nPackage `%s` contains generated playable token definitions whose names begin with %s. "+
				"Use `Cards` to iterate over every token definition in this package.\n",
			strings.ToUpper(letter), letter, strings.ToUpper(letter),
		)
		path := filepath.Join(tokenRoot, letter, "README.md")
		if err := os.WriteFile(path, []byte(readme), 0o600); err != nil {
			return fmt.Errorf("writing token letter package README: %w", err)
		}
	}
	if !isRepositoryCardsRoot(root) {
		return nil
	}

	var builder strings.Builder
	_, _ = builder.WriteString("// Code generated by compilecards; DO NOT EDIT.\n\n")
	_, _ = builder.WriteString("// Package tokens provides playable token definitions.\n")
	_, _ = builder.WriteString("package tokens\n\n")
	_, _ = builder.WriteString("import (\n\t\"slices\"\n\n")
	for _, letter := range ordered {
		_, _ = fmt.Fprintf(
			&builder,
			"\t\"github.com/natefinch/council4/mtg/cards/tokens/%s\"\n",
			letter,
		)
	}
	_, _ = builder.WriteString(")\n\n")
	_, _ = builder.WriteString("// Cards lists all generated playable token definitions.\n")
	_, _ = builder.WriteString("var Cards = slices.Concat(\n")
	for _, letter := range ordered {
		_, _ = fmt.Fprintf(&builder, "\t%s.Cards,\n", letter)
	}
	_, _ = builder.WriteString(")\n")
	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return fmt.Errorf("formatting token package: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tokenRoot, "cards.go"), formatted, 0o600); err != nil {
		return fmt.Errorf("writing token package: %w", err)
	}
	return nil
}

func isRepositoryCardsRoot(root string) bool {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	repositoryRoot, err := filepath.Abs(filepath.Join("mtg", "cards"))
	return err == nil && absoluteRoot == repositoryRoot
}

func writeCardList(directory string) error {
	source, err := cardlist.Generate(directory, filepath.Base(directory), "compilecards")
	if err != nil {
		return fmt.Errorf("generating %s/cards.go: %w", directory, err)
	}
	if err := os.WriteFile(filepath.Join(directory, "cards.go"), source, 0o600); err != nil {
		return fmt.Errorf("writing %s/cards.go: %w", directory, err)
	}
	return nil
}

// cardDefNames returns the builder-function names of the card definitions in a
// generated file, excluding token defs. A card registers only its builder and
// has no package-level var; a token def has an unexported var that calls its
// builder, so a builder a var calls is a token and is excluded. The builder name
// is the card's stable generated identifier, used to detect two Oracle cards
// that would compile to the same identifier in one package.
func cardDefNames(file *ast.File) []string {
	builders := cardDefBuilderFuncs(file)
	tokenBuilders := varReferencedBuilders(file)
	var names []string
	for builder := range builders {
		if !tokenBuilders[builder] {
			names = append(names, builder)
		}
	}
	return names
}

// varReferencedBuilders returns the builder-function names that a package-level
// var calls (e.g. `var xToken = newXToken()`), identifying token defs.
func varReferencedBuilders(file *ast.File) map[string]bool {
	referenced := map[string]bool{}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, specification := range general.Specs {
			values, ok := specification.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, value := range values.Values {
				call, ok := value.(*ast.CallExpr)
				if !ok {
					continue
				}
				if identifier, ok := call.Fun.(*ast.Ident); ok {
					referenced[identifier.Name] = true
				}
			}
		}
	}
	return referenced
}

// cardDefBuilderFuncs returns the set of function names in file that take no
// parameters and return *game.CardDef. Generated card files declare each card's
// construction as such a builder function.
func cardDefBuilderFuncs(file *ast.File) map[string]bool {
	builders := map[string]bool{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Name == nil {
			continue
		}
		if returnsCardDefPointer(function.Type) {
			builders[function.Name.Name] = true
		}
	}
	return builders
}

// returnsCardDefPointer reports whether fn returns exactly one value of type
// *game.CardDef.
func returnsCardDefPointer(fn *ast.FuncType) bool {
	if fn.Results == nil || len(fn.Results.List) != 1 {
		return false
	}
	star, ok := fn.Results.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	return ok && packageName.Name == "game" && selector.Sel.Name == "CardDef"
}
