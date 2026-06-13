// Command checkparser checks every Oracle text in a Scryfall card bulk-data
// file with the oracle lexer and syntax parser.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/internal/corpuscheck"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

type config struct {
	inputPath  string
	outputPath string
	format     string
	workers    int
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
	flags := flag.NewFlagSet("checkparser", flag.ContinueOnError)
	flags.StringVar(&cfg.inputPath, "in", "", "Scryfall card bulk-data JSON file")
	flags.StringVar(&cfg.outputPath, "out", "-", "report path, or - for stdout")
	flags.StringVar(&cfg.format, "format", "json", "report format: json or text")
	flags.IntVar(&cfg.workers, "workers", runtime.NumCPU(), "number of parser workers")
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
	report, err := corpuscheck.Check(input, cfg.workers, checkParser)
	if err != nil {
		return err
	}
	output, closeOutput, err := openOutput(cfg.outputPath)
	if err != nil {
		return err
	}
	defer closeOutput()
	switch cfg.format {
	case "json":
		encoder := json.NewEncoder(output)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("writing JSON report: %w", err)
		}
	case "text":
		if err := corpuscheck.WriteText(output, report); err != nil {
			return fmt.Errorf("writing text report: %w", err)
		}
	default:
		return fmt.Errorf("unsupported report format %q", cfg.format)
	}
	return nil
}

func checkParser(text corpuscheck.Text) []corpuscheck.Issue {
	_, diagnostics := parser.Parse(text.OracleText, parser.Context{
		InstantOrSorcery: hasCardType(text.TypeLine, "Instant") || hasCardType(text.TypeLine, "Sorcery"),
		Planeswalker:     hasCardType(text.TypeLine, "Planeswalker"),
		Saga:             hasSubtype(text.TypeLine, "Saga"),
	})
	issues := make([]corpuscheck.Issue, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		issues = append(issues, corpuscheck.Issue{
			Severity: severityName(diagnostic.Severity),
			Reason:   diagnostic.Summary,
			Detail:   diagnostic.Detail,
			Span:     diagnostic.Span,
		})
	}
	return issues
}

func hasCardType(typeLine, wanted string) bool {
	mainType, _, _ := strings.Cut(typeLine, "—")
	for word := range strings.FieldsSeq(mainType) {
		if word == wanted {
			return true
		}
	}
	return false
}

func hasSubtype(typeLine, wanted string) bool {
	_, subtypes, ok := strings.Cut(typeLine, "—")
	if !ok {
		return false
	}
	for word := range strings.FieldsSeq(subtypes) {
		if word == wanted {
			return true
		}
	}
	return false
}

func severityName(severity shared.Severity) string {
	switch severity {
	case shared.SeverityError:
		return "error"
	case shared.SeverityWarning:
		return "warning"
	default:
		return "unknown"
	}
}

func openOutput(path string) (io.Writer, func(), error) {
	if path == "-" {
		return os.Stdout, func() {}, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, func() {}, fmt.Errorf("creating report: %w", err)
	}
	return file, func() { _ = file.Close() }, nil
}
