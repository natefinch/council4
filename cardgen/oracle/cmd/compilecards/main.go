// Command compilecards generates fully executable CardDef source files for the
// strictly supported subset of a Scryfall Oracle Cards bulk-data file.
package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/cardgen/oracle"
)

type config struct {
	inputPath  string
	outputRoot string
	reportPath string
	format     string
	workers    int
}

type job struct {
	index int
	card  cardgen.ScryfallCard
}

type result struct {
	index       int
	card        cardgen.ScryfallCard
	relative    string
	source      string
	diagnostics []oracle.Diagnostic
	err         error
}

type report struct {
	CardCount        int           `json:"card_count"`
	GeneratedCount   int           `json:"generated_count"`
	UnsupportedCount int           `json:"unsupported_count"`
	Unsupported      []unsupported `json:"unsupported"`
}

type unsupported struct {
	ID          string             `json:"id,omitempty"`
	OracleID    string             `json:"oracle_id,omitempty"`
	Name        string             `json:"name"`
	Layout      string             `json:"layout,omitempty"`
	Diagnostics []reportDiagnostic `json:"diagnostics"`
}

type reportDiagnostic struct {
	Severity string      `json:"severity"`
	Summary  string      `json:"summary"`
	Detail   string      `json:"detail,omitempty"`
	Span     oracle.Span `json:"span"`
}

func main() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	if err := run(cfg); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func parseFlags(args []string) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("compilecards", flag.ContinueOnError)
	flags.StringVar(&cfg.inputPath, "in", "", "Scryfall Oracle Cards bulk-data JSON file")
	flags.StringVar(&cfg.outputRoot, "out", filepath.Join("mtg", "cards"), "output cards package root")
	flags.StringVar(&cfg.reportPath, "report", "-", "unsupported report path, or - for stdout")
	flags.StringVar(&cfg.format, "format", "json", "report format: json or text")
	flags.IntVar(&cfg.workers, "workers", runtime.NumCPU(), "number of compiler workers")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func run(cfg config) error {
	if cfg.inputPath == "" {
		return errors.New("-in is required")
	}
	if cfg.workers < 1 {
		return errors.New("-workers must be at least 1")
	}
	if cfg.format != "json" && cfg.format != "text" {
		return fmt.Errorf("unsupported -format %q", cfg.format)
	}
	input, err := os.Open(cfg.inputPath)
	if err != nil {
		return fmt.Errorf("opening input: %w", err)
	}
	defer input.Close()

	results, err := compileCorpus(input, cfg.workers)
	if err != nil {
		return err
	}
	report := buildReport(results)
	if err := writeSupported(cfg.outputRoot, results); err != nil {
		return err
	}
	return writeReport(cfg.reportPath, cfg.format, report)
}

func compileCorpus(input io.Reader, workers int) ([]result, error) {
	decoder := json.NewDecoder(input)
	tkn, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("reading bulk-data array: %w", err)
	}
	if delimiter, ok := tkn.(json.Delim); !ok || delimiter != '[' {
		return nil, errors.New("bulk data from Scryfall must be a top-level JSON array")
	}

	jobs := make(chan job)
	results := make(chan result)
	var workersDone sync.WaitGroup
	workersDone.Add(workers)
	for range workers {
		go func() {
			defer workersDone.Done()
			for item := range jobs {
				results <- compileCard(item)
			}
		}()
	}
	go func() {
		workersDone.Wait()
		close(results)
	}()

	decodeError := make(chan error, 1)
	go func() {
		defer close(jobs)
		sent := 0
		for decoder.More() {
			var card cardgen.ScryfallCard
			if err := decoder.Decode(&card); err != nil {
				decodeError <- fmt.Errorf("decoding card %d: %w", sent, err)
				return
			}
			jobs <- job{index: sent, card: card}
			sent++
		}
		if _, err := decoder.Token(); err != nil {
			decodeError <- fmt.Errorf("closing bulk-data array: %w", err)
			return
		}
		decodeError <- nil
	}()

	var all []result
	for compiled := range results {
		all = append(all, compiled)
	}
	if err := <-decodeError; err != nil {
		return nil, err
	}
	rejectPathCollisions(all)
	rejectIdentifierCollisions(all)
	slices.SortFunc(all, func(a, b result) int {
		return cmp.Compare(a.index, b.index)
	})
	return all, nil
}

func compileCard(item job) result {
	card := item.card
	letter := cardgen.CardNameToPackageLetter(card.Name)
	compiled := result{index: item.index, card: card}
	if len(letter) != 1 || letter[0] < 'a' || letter[0] > 'z' {
		compiled.diagnostics = []oracle.Diagnostic{{
			Severity: oracle.SeverityWarning,
			Summary:  "unsupported package letter",
			Detail:   fmt.Sprintf("card name %q does not map to an ASCII a-z package", card.Name),
		}}
		return compiled
	}
	compiled.relative = filepath.Join(letter, safeFileName(card.Name)+".go")
	compiled.source, compiled.diagnostics, compiled.err =
		cardgen.GenerateExecutableCardSource(&card, letter)
	return compiled
}

func rejectPathCollisions(results []result) {
	byPath := make(map[string][]int)
	for i := range results {
		if results[i].err == nil && len(results[i].diagnostics) == 0 {
			byPath[results[i].relative] = append(byPath[results[i].relative], i)
		}
	}
	for path, indexes := range byPath {
		if len(indexes) < 2 {
			continue
		}
		for _, index := range indexes {
			results[index].source = ""
			results[index].diagnostics = []oracle.Diagnostic{{
				Severity: oracle.SeverityWarning,
				Summary:  "generated path collision",
				Detail:   fmt.Sprintf("%d Oracle cards map to %s", len(indexes), path),
			}}
		}
	}
}

func rejectIdentifierCollisions(results []result) {
	byName := make(map[string][]int)
	for i := range results {
		if results[i].err != nil || len(results[i].diagnostics) > 0 {
			continue
		}
		file, err := parser.ParseFile(
			token.NewFileSet(),
			results[i].relative,
			results[i].source,
			0,
		)
		if err != nil {
			results[i].err = fmt.Errorf("parsing generated source: %w", err)
			continue
		}
		seen := make(map[string]bool)
		for _, name := range cardDefNames(file) {
			if seen[name] {
				results[i].source = ""
				results[i].diagnostics = []oracle.Diagnostic{{
					Severity: oracle.SeverityWarning,
					Summary:  "duplicate generated identifier",
					Detail:   fmt.Sprintf("generated source declares %s more than once", name),
				}}
				continue
			}
			seen[name] = true
			byName[name] = append(byName[name], i)
		}
	}
	for name, indexes := range byName {
		if len(indexes) < 2 {
			continue
		}
		for _, index := range indexes {
			results[index].source = ""
			results[index].diagnostics = []oracle.Diagnostic{{
				Severity: oracle.SeverityWarning,
				Summary:  "generated identifier collision",
				Detail: fmt.Sprintf(
					"%d Oracle cards declare %s in package %s",
					len(indexes),
					name,
					filepath.Dir(results[index].relative),
				),
			}}
		}
	}
}

func buildReport(results []result) report {
	output := report{CardCount: len(results)}
	for _, result := range results {
		if result.err == nil && len(result.diagnostics) == 0 {
			output.GeneratedCount++
			continue
		}
		diagnostics := result.diagnostics
		if result.err != nil {
			diagnostics = []oracle.Diagnostic{{
				Severity: oracle.SeverityError,
				Summary:  "source generation failed",
				Detail:   result.err.Error(),
			}}
		}
		output.Unsupported = append(output.Unsupported, unsupported{
			ID:          result.card.ID,
			OracleID:    result.card.OracleID,
			Name:        result.card.Name,
			Layout:      result.card.Layout,
			Diagnostics: reportDiagnostics(diagnostics),
		})
	}
	output.UnsupportedCount = len(output.Unsupported)
	return output
}

func reportDiagnostics(diagnostics []oracle.Diagnostic) []reportDiagnostic {
	output := make([]reportDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		output = append(output, reportDiagnostic{
			Severity: diagnosticSeverityName(diagnostic.Severity),
			Summary:  diagnostic.Summary,
			Detail:   diagnostic.Detail,
			Span:     diagnostic.Span,
		})
	}
	return output
}

func diagnosticSeverityName(severity oracle.Severity) string {
	switch severity {
	case oracle.SeverityError:
		return "error"
	case oracle.SeverityWarning:
		return "warning"
	default:
		return "unknown"
	}
}

func safeFileName(name string) string {
	base := cardgen.CardNameToFileName(name)
	if base == "cards" || strings.HasSuffix(base, "_test") {
		return base + "_card"
	}
	parts := strings.Split(base, "_")
	for _, suffix := range goFileSuffixes {
		if len(parts) >= len(suffix) && slices.Equal(parts[len(parts)-len(suffix):], suffix) {
			return base + "_card"
		}
	}
	return base
}

var goFileSuffixes = [][]string{
	{"aix"}, {"android"}, {"darwin"}, {"dragonfly"}, {"freebsd"}, {"illumos"},
	{"ios"}, {"js"}, {"linux"}, {"netbsd"}, {"openbsd"}, {"plan9"}, {"solaris"},
	{"wasip1"}, {"windows"},
	{"386"}, {"amd64"}, {"arm"}, {"arm64"}, {"loong64"}, {"mips"}, {"mips64"},
	{"mips64le"}, {"mipsle"}, {"ppc64"}, {"ppc64le"}, {"riscv64"}, {"s390x"},
	{"wasm"},
}

func writeSupported(root string, results []result) error {
	affected := make(map[string]bool)
	for _, result := range results {
		if result.err != nil || len(result.diagnostics) > 0 {
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
	return nil
}

func writeCardList(directory string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("reading %s: %w", directory, err)
	}
	varNames := make([]string, 0, len(entries))
	files := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || name == "cards.go" || strings.HasSuffix(name, "_test.go") ||
			!strings.HasSuffix(name, ".go") {
			continue
		}
		file, err := parser.ParseFile(files, filepath.Join(directory, name), nil, 0)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", name, err)
		}
		varNames = append(varNames, cardDefNames(file)...)
	}
	slices.Sort(varNames)

	var builder strings.Builder
	_, _ = builder.WriteString("// Code generated by compilecards; DO NOT EDIT.\n\n")
	_, _ = fmt.Fprintf(&builder, "package %s\n\n", filepath.Base(directory))
	_, _ = builder.WriteString("import \"github.com/natefinch/council4/mtg/game\"\n\n")
	_, _ = builder.WriteString("// Cards lists all card definitions in this package.\n")
	_, _ = builder.WriteString("var Cards = []*game.CardDef{\n")
	for _, name := range varNames {
		_, _ = fmt.Fprintf(&builder, "\t%s,\n", name)
	}
	_, _ = builder.WriteString("}\n")
	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return fmt.Errorf("formatting %s/cards.go: %w", directory, err)
	}
	if err := os.WriteFile(filepath.Join(directory, "cards.go"), formatted, 0o600); err != nil {
		return fmt.Errorf("writing %s/cards.go: %w", directory, err)
	}
	return nil
}

func cardDefNames(file *ast.File) []string {
	var names []string
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, specification := range general.Specs {
			values, ok := specification.(*ast.ValueSpec)
			if !ok || !isCardDef(values) {
				continue
			}
			for _, name := range values.Names {
				if name.Name != "" && unicode.IsUpper(rune(name.Name[0])) {
					names = append(names, name.Name)
				}
			}
		}
	}
	return names
}

func isCardDef(values *ast.ValueSpec) bool {
	for _, value := range values.Values {
		unary, ok := value.(*ast.UnaryExpr)
		if !ok {
			continue
		}
		composite, ok := unary.X.(*ast.CompositeLit)
		if !ok {
			continue
		}
		selector, ok := composite.Type.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		packageName, ok := selector.X.(*ast.Ident)
		if ok && packageName.Name == "game" && selector.Sel.Name == "CardDef" {
			return true
		}
	}
	return false
}

func writeReport(path, reportFormat string, output report) error {
	writer := io.Writer(os.Stdout)
	var file *os.File
	if path != "-" {
		var err error
		file, err = os.Create(path)
		if err != nil {
			return fmt.Errorf("creating report: %w", err)
		}
		defer file.Close()
		writer = file
	}
	switch reportFormat {
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return fmt.Errorf("writing JSON report: %w", err)
		}
	case "text":
		if _, err := fmt.Fprintf(
			writer,
			"cards: %d\ngenerated: %d\nunsupported: %d\n",
			output.CardCount,
			output.GeneratedCount,
			output.UnsupportedCount,
		); err != nil {
			return fmt.Errorf("writing text report summary: %w", err)
		}
		for _, card := range output.Unsupported {
			for _, diagnostic := range card.Diagnostics {
				if _, err := fmt.Fprintf(
					writer,
					"%s\t%s\t%s\n",
					card.Name,
					diagnostic.Summary,
					diagnostic.Detail,
				); err != nil {
					return fmt.Errorf("writing text report: %w", err)
				}
			}
		}
	default:
		return fmt.Errorf("unsupported report format %q", reportFormat)
	}
	return nil
}
