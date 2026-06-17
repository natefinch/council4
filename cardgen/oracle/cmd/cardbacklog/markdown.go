package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	loweringQueueLimit = 60
	parserQueueLimit   = 60
)

func writeMarkdown(path string, output report) error {
	var builder strings.Builder
	writeIntro(&builder)
	writeHeadline(&builder, output)
	writeLoweringQueue(&builder, output)
	writeParserQueue(&builder, output)

	if err := os.WriteFile(path, []byte(builder.String()), 0o600); err != nil {
		return fmt.Errorf("writing markdown: %w", err)
	}
	return nil
}

func writeIntro(builder *strings.Builder) {
	_, _ = builder.WriteString("# Card-Support Backlog\n\n")
	_, _ = builder.WriteString("Every eligible Scryfall corpus card is evaluated with two signals and routed ")
	_, _ = builder.WriteString("to the layer that blocks it:\n\n")
	_, _ = builder.WriteString("- **Parser signal** (parser-only): `cardgen.ParseCardFaces` + ")
	_, _ = builder.WriteString("`parser.DocumentCoverage` \u2014 is the card parser-complete, and which ")
	_, _ = builder.WriteString("uncovered components remain?\n")
	_, _ = builder.WriteString("- **Lowering signal** (full compile): compilecards' canonical report ")
	_, _ = builder.WriteString("\u2014 did the card generate, and if not, which distinct diagnostic ")
	_, _ = builder.WriteString("summaries blocked lowering? compilecards is the authority; an ")
	_, _ = builder.WriteString("independent per-card recompile reconciles against it.\n\n")
	_, _ = builder.WriteString("This produces two ranked, actionable queues. Regenerate with `mage cardBacklog`.\n\n")
}

func writeHeadline(builder *strings.Builder, output report) {
	_, _ = builder.WriteString("## Headline\n\n")
	_, _ = fmt.Fprintf(builder, "- Eligible cards: %d\n", output.EligibleCards)
	_, _ = fmt.Fprintf(builder, "- Supported (generated): %d\n", output.SupportedCards)
	_, _ = fmt.Fprintf(builder, "- Parser-complete: %d\n", output.ParserCompleteCards)
	_, _ = fmt.Fprintf(builder,
		"- **Lowering backlog** (parser-complete, not generated): %d\n", output.LoweringBacklog)
	_, _ = fmt.Fprintf(builder,
		"- **Parser backlog** (not parser-complete, not generated): %d\n", output.ParserBacklog)
	_, _ = builder.WriteString("\n")

	if output.PartitionOK {
		_, _ = fmt.Fprintf(builder,
			"Partition check: %d supported + %d lowering-backlog + %d parser-backlog = %d eligible. \u2713\n\n",
			output.SupportedCards, output.LoweringBacklog, output.ParserBacklog, output.EligibleCards)
	} else {
		_, _ = fmt.Fprintf(builder,
			"Partition check FAILED: %d + %d + %d \u2260 %d eligible.\n\n",
			output.SupportedCards, output.LoweringBacklog, output.ParserBacklog, output.EligibleCards)
	}

	if output.GeneratedIncomplete > 0 {
		writeGeneratedIncompleteNote(builder, output)
	}
	writeReconciliationNote(builder, output.Reconciliation)
}

// writeReconciliationNote records the independent guard outcome: how
// cardbacklog's own per-card recompile compares to compilecards' authoritative
// generated set. It documents that routing authority is compilecards' report,
// not the recompile, and surfaces any divergence inline.
func writeReconciliationNote(builder *strings.Builder, rec reconciliationReport) {
	_, _ = builder.WriteString("### Reconciliation guard\n\n")
	_, _ = builder.WriteString("Generated membership is read from compilecards' canonical report. An ")
	_, _ = builder.WriteString("independent per-card recompile cross-checks it; the run fails if they ")
	_, _ = builder.WriteString("diverge.\n\n")
	_, _ = fmt.Fprintf(builder,
		"- Authoritative generated (compilecards report): %d\n", rec.ReportGenerated)
	_, _ = fmt.Fprintf(builder,
		"- Independent per-card recompile generated: %d\n", rec.PerCardGenerated)
	if rec.OK {
		_, _ = builder.WriteString("- Divergences: 0 \u2014 the two pipelines agree. \u2713\n\n")
		return
	}
	_, _ = fmt.Fprintf(builder,
		"- **Divergences: %d generated, %d exclusion conflicts, %d missing from corpus.**\n",
		len(rec.GeneratedDivergences), len(rec.ExclusionConflicts), len(rec.MissingFromCorpus))
	for _, d := range rec.GeneratedDivergences {
		_, _ = fmt.Fprintf(builder, "  - %s (`%s`): per-card=%s, authoritative=%s\n",
			markdownCell(d.Name), d.ID, d.PerCard, d.Authoritative)
	}
	for _, d := range rec.ExclusionConflicts {
		_, _ = fmt.Fprintf(builder, "  - %s (`%s`): per-card=%s, authoritative=%s\n",
			markdownCell(d.Name), d.ID, d.PerCard, d.Authoritative)
	}
	_, _ = builder.WriteString("\n")
}

func writeGeneratedIncompleteNote(builder *strings.Builder, output report) {
	_, _ = fmt.Fprintf(builder,
		"%d generated cards are not parser-complete. The lowerer fully generates them, ",
		output.GeneratedIncomplete)
	_, _ = builder.WriteString("but the parser-coverage harness does not span all their must-cover tokens ")
	_, _ = builder.WriteString("(the residue tracked in `parser-coverage.md`). They are counted as ")
	_, _ = builder.WriteString("**supported**, not routed to either backlog queue:\n\n")
	for _, name := range output.GeneratedIncompleteNames {
		_, _ = fmt.Fprintf(builder, "- %s\n", name)
	}
	_, _ = builder.WriteString("\n")
}

func writeLoweringQueue(builder *strings.Builder, output report) {
	_, _ = builder.WriteString("## Lowering queue\n\n")
	_, _ = builder.WriteString("Parser-complete cards that do not yet lower, bucketed by distinct lowering ")
	_, _ = builder.WriteString("diagnostic summary and ranked by affected-card count. Parsing is already ")
	_, _ = builder.WriteString("done for these cards, so they are the lowest-risk backlog: this is ")
	_, _ = builder.WriteString("`unsupported-reasons.md` restricted to the parser-complete subset.\n\n")
	if len(output.LoweringQueue) == 0 {
		_, _ = builder.WriteString("None.\n\n")
		return
	}
	_, _ = builder.WriteString("| Rank | Reason | Affected (parser-complete) cards | Sole blockers | Example cards |\n")
	_, _ = builder.WriteString("| --- | --- | --- | --- | --- |\n")
	limit := min(loweringQueueLimit, len(output.LoweringQueue))
	for i := range limit {
		row := output.LoweringQueue[i]
		_, _ = fmt.Fprintf(builder, "| %d | %s | %d | %d | %s |\n",
			i+1, markdownCell(row.Reason), row.AffectedCards, row.SoleBlockers,
			markdownCell(strings.Join(row.Examples, "; ")))
	}
	_, _ = builder.WriteString("\n")
}

func writeParserQueue(builder *strings.Builder, output report) {
	_, _ = builder.WriteString("## Parser queue\n\n")
	_, _ = builder.WriteString("Cards that are not parser-complete (and do not lower), bucketed by owning ")
	_, _ = builder.WriteString("component family and normalized uncovered-span cluster, ranked by ")
	_, _ = builder.WriteString("occurrence. This is the grammar-recognition backlog.\n\n")
	if len(output.ParserQueue) == 0 {
		_, _ = builder.WriteString("None.\n\n")
		return
	}
	_, _ = builder.WriteString("| Rank | Component | Cluster | Count | Example cards |\n")
	_, _ = builder.WriteString("| --- | --- | --- | --- | --- |\n")
	limit := min(parserQueueLimit, len(output.ParserQueue))
	for i := range limit {
		row := output.ParserQueue[i]
		_, _ = fmt.Fprintf(builder, "| %d | %s | %s | %d | %s |\n",
			i+1, markdownCell(row.Component), markdownCell(row.Cluster), row.Count,
			markdownCell(strings.Join(row.Examples, "; ")))
	}
	_, _ = builder.WriteString("\n")
}

func markdownCell(text string) string {
	text = strings.ReplaceAll(text, "|", "\\|")
	text = strings.ReplaceAll(text, "\n", " ")
	return text
}
