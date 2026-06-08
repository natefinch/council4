// Command checklexer checks every Oracle text in a Scryfall card bulk-data
// file and reports cards containing text the oracle lexer cannot tokenize.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/natefinch/council4/cardgen/oracle"
	"github.com/natefinch/council4/cardgen/oracle/internal/corpuscheck"
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
	flags := flag.NewFlagSet("checklexer", flag.ContinueOnError)
	flags.StringVar(&cfg.inputPath, "in", "", "Scryfall card bulk-data JSON file")
	flags.StringVar(&cfg.outputPath, "out", "-", "report path, or - for stdout")
	flags.StringVar(&cfg.format, "format", "json", "report format: json or text")
	flags.IntVar(&cfg.workers, "workers", runtime.NumCPU(), "number of lexer workers")
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

	report, err := corpuscheck.Check(input, cfg.workers, checkLexer)
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

func checkLexer(text corpuscheck.Text) []corpuscheck.Issue {
	lexer := oracle.NewLexer(text.OracleText)
	var issues []corpuscheck.Issue
	for {
		token := lexer.Next()
		if token.Kind == oracle.Invalid {
			issues = append(issues, corpuscheck.Issue{
				Severity: "error",
				Reason:   oracle.InvalidReason(token),
				Text:     token.Text,
				Span:     token.Span,
			})
		}
		if token.Kind == oracle.EOF {
			return issues
		}
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
