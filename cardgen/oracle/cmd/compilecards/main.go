// Command compilecards generates fully executable CardDef source files for the
// strictly supported subset of a Scryfall Oracle Cards bulk-data file.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

type config struct {
	inputPath              string
	outputRoot             string
	reportPath             string
	format                 string
	supportedPath          string
	unsupportedPath        string
	unsupportedReasonsPath string
	readmePath             string
	workers                int
}

type job struct {
	index int
	card  cardgen.ScryfallCard
}

type result struct {
	index       int
	card        cardgen.ScryfallCard
	relative    string
	superseded  string
	source      string
	exclusion   cardgen.CorpusExclusionReason
	diagnostics []shared.Diagnostic
	err         error
}

type report struct {
	CardCount        int           `json:"card_count"`
	EligibleCount    int           `json:"eligible_count"`
	GeneratedCount   int           `json:"generated_count"`
	UnsupportedCount int           `json:"unsupported_count"`
	ExcludedCount    int           `json:"excluded_count"`
	Unsupported      []unsupported `json:"unsupported"`
	Excluded         []excluded    `json:"excluded"`
}

type unsupported struct {
	ID          string             `json:"id,omitempty"`
	OracleID    string             `json:"oracle_id,omitempty"`
	Name        string             `json:"name"`
	Layout      string             `json:"layout,omitempty"`
	Diagnostics []reportDiagnostic `json:"diagnostics"`
}

type excluded struct {
	ID       string                        `json:"id,omitempty"`
	OracleID string                        `json:"oracle_id,omitempty"`
	Name     string                        `json:"name"`
	Layout   string                        `json:"layout,omitempty"`
	Reason   cardgen.CorpusExclusionReason `json:"reason"`
}

type reportDiagnostic struct {
	Severity string      `json:"severity"`
	Summary  string      `json:"summary"`
	Detail   string      `json:"detail,omitempty"`
	Span     shared.Span `json:"span"`
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
	flags.StringVar(&cfg.supportedPath, "supported", "", "supported-card Markdown path")
	flags.StringVar(&cfg.unsupportedPath, "unsupported", "", "unsupported-card Markdown path")
	flags.StringVar(&cfg.unsupportedReasonsPath, "unsupported-reasons", "", "card-support planning Markdown path")
	flags.StringVar(&cfg.readmePath, "readme", "", "README path whose card-support block should be updated")
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
	if err := writeSupportDocumentation(cfg, report, results); err != nil {
		return err
	}
	return writeReport(cfg.reportPath, cfg.format, report)
}
