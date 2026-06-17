// Command cardbacklog joins the parser-coverage signal with compilecards'
// authoritative lowering/compile signal for every eligible Scryfall corpus card
// and routes each unsupported card to the layer that blocks it. It emits two
// ranked, actionable task queues (a lowering queue of parser-complete cards that
// do not yet lower, and a parser queue of cards the grammar cannot yet
// represent) plus a headline that partitions the eligible corpus. Generated
// membership is read from compilecards' canonical JSON report, never decided by
// cardbacklog alone; an independent per-card recompile cross-checks that report
// and fails the run loudly on any divergence.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type config struct {
	inputPath     string
	outputPath    string
	reportPath    string
	compileReport string
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
	flags := flag.NewFlagSet("cardbacklog", flag.ContinueOnError)
	flags.StringVar(&cfg.inputPath, "in", "", "Scryfall Oracle Cards bulk-data JSON file")
	flags.StringVar(&cfg.outputPath, "out", "card-backlog.md", "card-backlog Markdown path")
	flags.StringVar(&cfg.reportPath, "report", "-", "JSON report path, or - for stdout")
	flags.StringVar(&cfg.compileReport, "compile-report", "",
		"compilecards JSON report providing the authoritative generated/unsupported set (required)")
	flags.IntVar(&cfg.workers, "workers", 0, "number of workers (0 selects NumCPU)")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func run(cfg config) error {
	if cfg.inputPath == "" {
		return errors.New("-in is required")
	}
	if cfg.compileReport == "" {
		return errors.New("-compile-report is required: run compilecards first to produce the authoritative report")
	}

	authority, raw, err := loadCompileAuthority(cfg.compileReport)
	if err != nil {
		return err
	}

	input, err := os.Open(cfg.inputPath)
	if err != nil {
		return fmt.Errorf("opening input: %w", err)
	}
	defer input.Close()

	outcomes, err := parseCorpus(input, cfg.workers)
	if err != nil {
		return err
	}

	rec := applyAuthority(outcomes, authority)
	output := buildReport(outcomes)
	output.Reconciliation = buildReconciliationReport(rec, raw.GeneratedCount, perCardGeneratedCount(outcomes))
	reportSummary(output)

	if err := writeMarkdown(cfg.outputPath, output); err != nil {
		return err
	}
	if err := writeReport(cfg.reportPath, output); err != nil {
		return err
	}

	if !output.PartitionOK {
		return fmt.Errorf("partition check failed: %d + %d + %d != %d eligible",
			output.SupportedCards, output.LoweringBacklog, output.ParserBacklog, output.EligibleCards)
	}
	if !output.Reconciliation.OK {
		return fmt.Errorf("reconciliation failed: per-card recompile diverged from compilecards' report "+
			"(%d generated divergences, %d exclusion conflicts, %d missing from corpus)",
			len(output.Reconciliation.GeneratedDivergences),
			len(output.Reconciliation.ExclusionConflicts),
			len(output.Reconciliation.MissingFromCorpus))
	}
	return nil
}

// perCardGeneratedCount counts eligible cards the independent per-card recompile
// considered generated, the left-hand side of the reconciliation cross-check.
func perCardGeneratedCount(outcomes []cardOutcome) int {
	count := 0
	for i := range outcomes {
		if outcomes[i].eligible && outcomes[i].perCardGenerated {
			count++
		}
	}
	return count
}

func buildReconciliationReport(rec reconciliation, reportGenerated, perCardGenerated int) reconciliationReport {
	return reconciliationReport{
		OK:                   rec.ok(),
		ReportGenerated:      reportGenerated,
		PerCardGenerated:     perCardGenerated,
		GeneratedDivergences: rec.generatedDivergences,
		ExclusionConflicts:   rec.exclusionConflicts,
		MissingFromCorpus:    rec.missingFromCorpus,
	}
}
