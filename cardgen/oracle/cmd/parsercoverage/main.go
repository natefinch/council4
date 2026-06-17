// Command parsercoverage measures how completely the Oracle parser represents
// the eligible Scryfall corpus as typed syntax, without running the compiler or
// lowering. It reports the effect-level exact percentage, the card-level
// parser-complete percentage, and a ranked queue of the unrepresented grammar.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type config struct {
	inputPath     string
	reportPath    string
	outputPath    string
	generatedPath string
	workers       int
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
	flags := flag.NewFlagSet("parsercoverage", flag.ContinueOnError)
	flags.StringVar(&cfg.inputPath, "in", "", "Scryfall Oracle Cards bulk-data JSON file")
	flags.StringVar(&cfg.reportPath, "report", "-", "JSON report path, or - for stdout")
	flags.StringVar(&cfg.outputPath, "out", "parser-coverage.md", "parser-coverage Markdown path")
	flags.StringVar(&cfg.generatedPath, "generated", "",
		"optional supported-card Markdown (one \"- <name>\" per line) to assert generated cards are a subset of parser-complete cards")
	flags.IntVar(&cfg.workers, "workers", 0, "number of parser workers (0 selects NumCPU)")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func run(cfg config) error {
	if cfg.inputPath == "" {
		return errors.New("-in is required")
	}
	input, err := os.Open(cfg.inputPath)
	if err != nil {
		return fmt.Errorf("opening input: %w", err)
	}
	defer input.Close()

	cards, err := parseCorpus(input, cfg.workers)
	if err != nil {
		return err
	}
	report := buildReport(cards)

	if cfg.generatedPath != "" {
		generatedCount, violations, err := generatedSubsetViolations(cfg.generatedPath, cards)
		if err != nil {
			return err
		}
		report.Validation = buildValidation(generatedCount, violations)
		reportSubsetViolations(violations)
	}

	if err := writeMarkdown(cfg.outputPath, report); err != nil {
		return err
	}
	return writeReport(cfg.reportPath, report)
}
