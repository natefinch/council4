package main

import (
	"fmt"
	"os"
	"strings"
)

const markdownClusterLimit = 40

func writeMarkdown(path string, output report) error {
	var builder strings.Builder
	_, _ = builder.WriteString("# Parser Coverage\n\n")
	_, _ = builder.WriteString("Parser-only coverage across the eligible Scryfall corpus, measured without ")
	_, _ = builder.WriteString("running the compiler or lowering. Two distinct metrics are reported:\n\n")
	_, _ = builder.WriteString("- **Parser-complete (typed coverage):** every must-cover token of every ")
	_, _ = builder.WriteString("ability is accounted for by a kind-recognized typed element. This is an ")
	_, _ = builder.WriteString("upper bound on what the lowerer could consume \u2014 it does not require byte-")
	_, _ = builder.WriteString("exact reconstruction.\n")
	_, _ = builder.WriteString("- **Exact round-trip:** the parser reconstructs the original text byte-for-")
	_, _ = builder.WriteString("byte (`effect.Exact`). Strictly stronger than typed coverage.\n\n")
	_, _ = builder.WriteString("Regenerate with `mage parserCoverage`.\n\n")

	_, _ = builder.WriteString("## Headline\n\n")
	_, _ = fmt.Fprintf(&builder, "- Eligible cards: %d\n", output.EligibleCards)
	_, _ = fmt.Fprintf(&builder, "- Parser-complete cards (typed coverage): %d (%.2f%%)\n", output.ParserComplete, output.CompletePercent)
	_, _ = fmt.Fprintf(&builder, "- Exact round-trip cards (complete and every effect exact): %d (%.2f%%)\n", output.CardExact, output.CardExactPercent)
	_, _ = fmt.Fprintf(&builder, "- Resolving effects: %d\n", output.ResolvingEffects)
	_, _ = fmt.Fprintf(&builder, "- Exact round-trip effects: %d (%.2f%%)\n", output.ExactEffects, output.ExactPercent)
	_, _ = builder.WriteString("\n")

	if output.Validation != nil {
		_, _ = builder.WriteString("## Generated \u2286 Parser-complete\n\n")
		_, _ = fmt.Fprintf(&builder, "- Generated cards: %d\n", output.Validation.GeneratedCards)
		_, _ = fmt.Fprintf(&builder, "- Violations: %d\n", output.Validation.Violations)
		for _, name := range output.Validation.ViolationNames {
			_, _ = fmt.Fprintf(&builder, "  - %s\n", name)
		}
		_, _ = builder.WriteString("\n")
		if output.Validation.Violations > 0 {
			writeResidueNote(&builder)
		}
	}

	writeBlockerSummary(&builder, output)
	writeClusterQueue(&builder, output)

	if err := os.WriteFile(path, []byte(builder.String()), 0o600); err != nil {
		return fmt.Errorf("writing markdown: %w", err)
	}
	return nil
}

func writeResidueNote(builder *strings.Builder) {
	_, _ = builder.WriteString("These generated cards are not parser-complete because the parser ")
	_, _ = builder.WriteString("recognizes the construct semantically (the effect round-trips or the trigger ")
	_, _ = builder.WriteString("event is typed) but does not expose a source span covering all of its ")
	_, _ = builder.WriteString("must-cover tokens, so the coverage harness cannot credit them without ")
	_, _ = builder.WriteString("over-crediting an adjacent clause. The unspanned material falls into three ")
	_, _ = builder.WriteString("groups: coordinated trigger/condition lists whose typed span stops at the ")
	_, _ = builder.WriteString("first list item (e.g. \"instant, sorcery, or Wizard spell\"); \"for each ")
	_, _ = builder.WriteString("X\" iteration prefixes on a create-token effect; and reflexive/delayed ")
	_, _ = builder.WriteString("trigger preambles (\"When you do,\" / \"Whenever ... this turn,\"). Widening ")
	_, _ = builder.WriteString("those parser spans is tracked separately; they are reported here rather ")
	_, _ = builder.WriteString("than hidden by loosening the metric.\n\n")
}

func writeBlockerSummary(builder *strings.Builder, output report) {
	_, _ = builder.WriteString("## Uncovered components by blocker\n\n")
	if len(output.BlockerSummary) == 0 {
		_, _ = builder.WriteString("None.\n\n")
		return
	}
	_, _ = builder.WriteString("| Blocker | Components |\n| --- | --- |\n")
	for _, entry := range output.BlockerSummary {
		_, _ = fmt.Fprintf(builder, "| %s | %d |\n", entry.Blocker, entry.Count)
	}
	_, _ = builder.WriteString("\n")
}

func writeClusterQueue(builder *strings.Builder, output report) {
	_, _ = builder.WriteString("## Uncovered grammar work queue\n\n")
	if len(output.UncoveredClusters) == 0 {
		_, _ = builder.WriteString("None.\n")
		return
	}
	_, _ = builder.WriteString("Top uncovered span clusters (normalized), ranked by occurrence.\n\n")
	_, _ = builder.WriteString("| Rank | Count | Blocker | Cluster | Examples |\n| --- | --- | --- | --- | --- |\n")
	limit := min(markdownClusterLimit, len(output.UncoveredClusters))
	for i := range limit {
		entry := output.UncoveredClusters[i]
		_, _ = fmt.Fprintf(builder, "| %d | %d | %s | %s | %s |\n",
			i+1, entry.Count, entry.Blocker,
			markdownCell(entry.Text), markdownCell(strings.Join(entry.Examples, "; ")))
	}
	_, _ = builder.WriteString("\n")
}

func markdownCell(text string) string {
	text = strings.ReplaceAll(text, "|", "\\|")
	text = strings.ReplaceAll(text, "\n", " ")
	return text
}
