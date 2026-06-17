package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// report is the machine-readable card-support backlog: the headline partition
// counts, the reconciliation guard result, plus the two ranked queues.
type report struct {
	TotalCards               int                  `json:"total_cards"`
	ExcludedCards            int                  `json:"excluded_cards"`
	EligibleCards            int                  `json:"eligible_cards"`
	SupportedCards           int                  `json:"supported_cards"`
	ParserCompleteCards      int                  `json:"parser_complete_cards"`
	LoweringBacklog          int                  `json:"lowering_backlog_cards"`
	ParserBacklog            int                  `json:"parser_backlog_cards"`
	GeneratedIncomplete      int                  `json:"generated_but_incomplete_cards"`
	PartitionOK              bool                 `json:"partition_ok"`
	Reconciliation           reconciliationReport `json:"reconciliation"`
	LoweringQueue            []loweringReason     `json:"lowering_queue"`
	ParserQueue              []parserClusterRow   `json:"parser_queue"`
	GeneratedIncompleteNames []string             `json:"generated_but_incomplete_names,omitempty"`
}

// reconciliationReport records the independent cross-check of cardbacklog's
// per-card recompile against compilecards' authoritative report. ReportGenerated
// is the canonical generated count; PerCardGenerated is the per-card recompute.
// A clean run has Divergences == 0 and OK == true.
type reconciliationReport struct {
	OK                   bool         `json:"ok"`
	ReportGenerated      int          `json:"report_generated_count"`
	PerCardGenerated     int          `json:"per_card_generated_count"`
	GeneratedDivergences []divergence `json:"generated_divergences,omitempty"`
	ExclusionConflicts   []divergence `json:"exclusion_conflicts,omitempty"`
	MissingFromCorpus    []string     `json:"missing_from_corpus,omitempty"`
}

// loweringReason is one lowering-queue row: a distinct lowering diagnostic
// summary and how many parser-complete-but-ungenerated cards it blocks.
type loweringReason struct {
	Reason        string   `json:"reason"`
	AffectedCards int      `json:"affected_cards"`
	SoleBlockers  int      `json:"sole_blocker_cards"`
	Examples      []string `json:"examples"`
}

// parserClusterRow is one parser-queue row: an owning component family and a
// normalized uncovered-span cluster, with its occurrence count.
type parserClusterRow struct {
	Component string   `json:"component"`
	Cluster   string   `json:"cluster"`
	Count     int      `json:"count"`
	Examples  []string `json:"examples"`
}

func writeReport(path string, output report) error {
	encoded, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding report: %w", err)
	}
	encoded = append(encoded, '\n')
	if path == "-" {
		_, err := os.Stdout.Write(encoded)
		return err
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}
	return nil
}

// reportSummary prints the headline partition counts to stderr so a run is
// verifiable without opening the report.
func reportSummary(output report) {
	_, _ = fmt.Fprintf(os.Stderr, "eligible: %d\n", output.EligibleCards)
	_, _ = fmt.Fprintf(os.Stderr, "supported (generated): %d\n", output.SupportedCards)
	_, _ = fmt.Fprintf(os.Stderr, "parser-complete: %d\n", output.ParserCompleteCards)
	_, _ = fmt.Fprintf(os.Stderr, "lowering-backlog (parser-complete, ungenerated): %d\n", output.LoweringBacklog)
	_, _ = fmt.Fprintf(os.Stderr, "parser-backlog (parser-incomplete, ungenerated): %d\n", output.ParserBacklog)
	_, _ = fmt.Fprintf(os.Stderr, "generated-but-incomplete (in supported): %d\n", output.GeneratedIncomplete)
	if output.PartitionOK {
		_, _ = fmt.Fprintf(os.Stderr,
			"partition OK: %d + %d + %d = %d\n",
			output.SupportedCards, output.LoweringBacklog, output.ParserBacklog, output.EligibleCards)
	} else {
		_, _ = fmt.Fprintf(os.Stderr,
			"partition MISMATCH: %d + %d + %d != %d\n",
			output.SupportedCards, output.LoweringBacklog, output.ParserBacklog, output.EligibleCards)
	}
	reportReconciliation(output.Reconciliation)
}

// reportReconciliation prints the independent guard result to stderr. It compares
// the authoritative generated count (from compilecards' report) against
// cardbacklog's own per-card recompute and lists any disagreement, so a future
// divergence between the two pipelines fails visibly instead of silently
// mis-routing a card.
func reportReconciliation(rec reconciliationReport) {
	_, _ = fmt.Fprintf(os.Stderr,
		"reconciliation: report-generated %d vs per-card-generated %d\n",
		rec.ReportGenerated, rec.PerCardGenerated)
	if rec.OK {
		_, _ = fmt.Fprintln(os.Stderr, "reconciliation OK: 0 divergences")
		return
	}
	_, _ = fmt.Fprintf(os.Stderr,
		"reconciliation DIVERGENCE: %d generated, %d exclusion conflicts, %d missing from corpus\n",
		len(rec.GeneratedDivergences), len(rec.ExclusionConflicts), len(rec.MissingFromCorpus))
	for _, d := range rec.GeneratedDivergences {
		_, _ = fmt.Fprintf(os.Stderr, "  generated: %s (%s) per-card=%s authoritative=%s\n",
			d.Name, d.ID, d.PerCard, d.Authoritative)
	}
	for _, d := range rec.ExclusionConflicts {
		_, _ = fmt.Fprintf(os.Stderr, "  exclusion: %s (%s) per-card=%s authoritative=%s\n",
			d.Name, d.ID, d.PerCard, d.Authoritative)
	}
	for _, id := range rec.MissingFromCorpus {
		_, _ = fmt.Fprintf(os.Stderr, "  missing from corpus: %s\n", id)
	}
}
