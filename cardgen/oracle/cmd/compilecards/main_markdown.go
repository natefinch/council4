package main

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const (
	readmeSupportStart = "<!-- card-support:start -->"
	readmeSupportEnd   = "<!-- card-support:end -->"
)

func writeSupportDocumentation(cfg config, output report, results []result) error {
	if cfg.supportedPath != "" {
		if err := writeSupportedMarkdown(cfg.supportedPath, output, results); err != nil {
			return err
		}
	}
	if cfg.unsupportedPath != "" {
		if err := writeUnsupportedMarkdown(cfg.unsupportedPath, output); err != nil {
			return err
		}
	}
	if cfg.unsupportedReasonsPath != "" {
		if err := writeUnsupportedReasonsMarkdown(cfg.unsupportedReasonsPath, output, results); err != nil {
			return err
		}
	}
	if cfg.readmePath != "" {
		if err := updateReadmeSupport(cfg.readmePath, output); err != nil {
			return err
		}
	}
	return nil
}

func writeSupportedMarkdown(path string, output report, results []result) error {
	names := make([]string, 0, output.GeneratedCount)
	for _, result := range results {
		if result.exclusion == "" && result.err == nil && len(result.diagnostics) == 0 {
			names = append(names, result.card.Name)
		}
	}
	slices.SortFunc(names, func(a, b string) int {
		if compared := cmp.Compare(strings.ToLower(a), strings.ToLower(b)); compared != 0 {
			return compared
		}
		return cmp.Compare(a, b)
	})

	var builder strings.Builder
	_, _ = builder.WriteString("# Supported Cards\n\n")
	_, _ = builder.WriteString(supportSummary(output))
	_, _ = builder.WriteString("\n\n")
	for _, name := range names {
		_, _ = fmt.Fprintf(&builder, "- %s\n", markdownInline(name))
	}
	return writeDocumentationFile(path, builder.String())
}

func writeUnsupportedMarkdown(path string, output report) error {
	cards := append([]unsupported(nil), output.Unsupported...)
	slices.SortFunc(cards, func(a, b unsupported) int {
		if compared := cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); compared != 0 {
			return compared
		}
		if compared := cmp.Compare(a.Name, b.Name); compared != 0 {
			return compared
		}
		return cmp.Compare(a.OracleID, b.OracleID)
	})

	var builder strings.Builder
	_, _ = builder.WriteString("# Unsupported Cards\n\n")
	_, _ = builder.WriteString(supportSummary(output))
	_, _ = builder.WriteString("\n\n")
	_, _ = builder.WriteString(
		"These cards are eligible for paper support but cardgen cannot yet generate them. " +
			"Cards excluded by the corpus policy are not listed.\n\n",
	)
	for _, card := range cards {
		reasons := make([]string, 0, len(card.Diagnostics))
		for _, diagnostic := range card.Diagnostics {
			reason := markdownInline(diagnostic.Summary)
			if diagnostic.Detail != "" {
				reason += ": " + markdownInline(diagnostic.Detail)
			}
			reasons = append(reasons, reason)
		}
		_, _ = fmt.Fprintf(
			&builder,
			"- **%s** — %s\n",
			markdownInline(card.Name),
			strings.Join(reasons, "; "),
		)
	}
	return writeDocumentationFile(path, builder.String())
}

func writeUnsupportedReasonsMarkdown(path string, output report, results []result) error {
	analysis := analyzeSupport(output)
	var builder strings.Builder
	_, _ = builder.WriteString("# Card-Support Planning Report\n\n")
	_, _ = builder.WriteString(
		"Capability-aware blockers for eligible paper cards that cannot yet be generated. " +
			"Each distinct diagnostic summary and capability is counted at most once per card.\n\n",
	)
	_, _ = builder.WriteString("## Diagnostic reasons\n\n")
	_, _ = builder.WriteString(
		"A sole blocker is the card's only distinct diagnostic summary. " +
			"The most common co-blocker excludes the reason in its own row.\n\n",
	)
	_, _ = builder.WriteString(
		"| Rank | Reason | Affected cards | Sole blockers | Sole blocker % | Most common co-blocker |\n",
	)
	_, _ = builder.WriteString("| ---: | --- | ---: | ---: | ---: | --- |\n")
	for index, reason := range analysis.reasons {
		coBlocker := "-"
		if reason.mostCommonCoBlocker != "" {
			coBlocker = markdownTableCell(reason.mostCommonCoBlocker)
		}
		_, _ = fmt.Fprintf(
			&builder,
			"| %d | %s | %s | %s | %.1f%% | %s |\n",
			index+1,
			markdownTableCell(reason.summary),
			formatCount(reason.affectedCards),
			formatCount(reason.soleBlockerCards),
			reason.soleBlockerPercentage(),
			coBlocker,
		)
	}
	_, _ = builder.WriteString("\n## Capability clusters\n\n")
	_, _ = builder.WriteString(
		"A fully unlockable card has every distinct diagnostic summary in one capability cluster. " +
			"Constituent summaries list the diagnostics currently observed in that cluster.\n\n",
	)
	_, _ = builder.WriteString(
		"| Capability | Affected cards | Fully unlockable cards | Constituent diagnostic summaries |\n",
	)
	_, _ = builder.WriteString("| --- | ---: | ---: | --- |\n")
	for _, capability := range analysis.capabilities {
		summaries := make([]string, len(capability.summaries))
		for index, summary := range capability.summaries {
			summaries[index] = markdownTableCell(summary)
		}
		_, _ = fmt.Fprintf(
			&builder,
			"| %s | %s | %s | %s |\n",
			capability.id,
			formatCount(capability.affectedCards),
			formatCount(capability.fullyUnlockableCards),
			strings.Join(summaries, "; "),
		)
	}
	writeUnblockRoadmap(&builder, output)
	writeOrderedSequenceCategories(&builder, output)
	writeConditionRecognitionBacklog(&builder, output)
	writeEnvelopeGapBacklog(&builder, output, results)
	return writeDocumentationFile(path, builder.String())
}

// writeUnblockRoadmap renders the greedy set-cover priority of fixes: the reasons
// that, applied in order, fully unblock the most cards. It is the direct answer to
// "how do we unblock the most cards at once?" — because a card is generated only
// when every one of its distinct diagnostic reasons is resolved.
func writeUnblockRoadmap(builder *strings.Builder, output report) {
	steps := analyzeUnblockRoadmap(output)
	if len(steps) == 0 {
		return
	}
	_, _ = builder.WriteString("\n## Unblock roadmap\n\n")
	_, _ = builder.WriteString(
		"Greedy set-cover priority: each step fixes the reason that — given the reasons already " +
			"fixed in the steps above it — newly fully unblocks the most still-blocked cards. " +
			"Cumulative is the running total of cards fully unblocked. Fan-out lowerers (ordered " +
			"sequence, modal, optional) now report every independent blocker they carry, so these " +
			"counts account for co-blockers rather than crediting a fix with cards that need other " +
			"fixes too. A few remaining lowerers still short-circuit within an ability, so the " +
			"counts stay a slight over-estimate.\n\n",
	)
	_, _ = builder.WriteString(
		"| Step | Fix this reason | Capability | Newly unblocked | Cumulative | Sample cards |\n",
	)
	_, _ = builder.WriteString("| ---: | --- | --- | ---: | ---: | --- |\n")
	for index, step := range steps {
		_, _ = fmt.Fprintf(
			builder,
			"| %d | %s | %s | %s | %s | %s |\n",
			index+1,
			markdownTableCell(step.summary),
			step.capability,
			formatCount(step.newlyUnblocked),
			formatCount(step.cumulativeUnblocked),
			markdownTableCell(strings.Join(step.sampleCards, ", ")),
		)
	}
}

// writeOrderedSequenceCategories renders the sub-category breakdown of the
// "unsupported ordered effect sequence" reason, so the largest sole-blocker
// bucket is legible rather than opaque.
func writeOrderedSequenceCategories(builder *strings.Builder, output report) {
	categories := analyzeOrderedSequenceCategories(output)
	if len(categories) == 0 {
		return
	}
	_, _ = builder.WriteString("\n## Ordered effect sequence sub-categories\n\n")
	_, _ = builder.WriteString(
		"Breakdown of the `unsupported ordered effect sequence` reason by the specific blocker " +
			"within the sequence. A `sub-effect` row names the single-effect lowering a clause needs " +
			"before its sequence can compile; a `structural` row names a sequence-machinery limitation. " +
			"Counts mirror the diagnostic-reasons table: affected cards include co-blocked cards, sole " +
			"blockers do not.\n\n",
	)
	_, _ = builder.WriteString("| Category | Affected cards | Sole blockers |\n")
	_, _ = builder.WriteString("| --- | ---: | ---: |\n")
	for _, category := range categories {
		_, _ = fmt.Fprintf(
			builder,
			"| %s | %s | %s |\n",
			markdownTableCell(category.category),
			formatCount(category.affectedCards),
			formatCount(category.soleBlockerCards),
		)
	}
}

// writeConditionRecognitionBacklog renders the ranked list of unrecognized
// per-effect condition wordings that block ordered-sequence lowering. It is the
// actionable drill-down beneath the "structural — per-effect condition
// unrecognized" sub-category: each row names a condition the compiler does not
// yet recognize and how many cards recognizing it would unblock, so coverage
// work can be prioritized by leverage.
func writeConditionRecognitionBacklog(builder *strings.Builder, output report) {
	backlog := analyzeConditionRecognitionBacklog(output)
	if len(backlog) == 0 {
		return
	}
	_, _ = builder.WriteString("\n## Unrecognized per-effect conditions (recognition backlog)\n\n")
	_, _ = builder.WriteString(
		"Distinct `if <condition>` wordings inside ordered sequences whose predicate the compiler " +
			"does not yet recognize. Recognizing a wording unblocks ordered-sequence lowering for the " +
			"listed cards. Rows are ranked by sole blockers (cards a single wording is the only blocker " +
			"for) then affected cards.\n\n",
	)
	_, _ = builder.WriteString("| Unrecognized condition | Affected cards | Sole blockers |\n")
	_, _ = builder.WriteString("| --- | ---: | ---: |\n")
	for _, entry := range backlog {
		_, _ = fmt.Fprintf(
			builder,
			"| %s | %s | %s |\n",
			markdownTableCell(entry.condition),
			formatCount(entry.affectedCards),
			formatCount(entry.soleBlockerCards),
		)
	}
}

// writeEnvelopeGapBacklog renders the ranked list of modeled-capability envelope
// gaps: families the compiler recognizes but lowers only within an exact
// envelope, ranked by how many cards a single envelope detail sole-blocks. Each
// row shows a few example wordings so the specific parameter the envelope must
// grow to cover (a dynamic amount, a graveyard zone, a multi-target count, a
// filter) is visible. It is the effect-family analogue of the unrecognized
// per-effect condition backlog.
func writeEnvelopeGapBacklog(builder *strings.Builder, output report, results []result) {
	gaps := analyzeEnvelopeGapBacklog(output)
	if len(gaps) == 0 {
		return
	}
	samples := collectEnvelopeSamples(output, results)
	_, _ = builder.WriteString("\n## Modeled-capability envelope gaps (parameter backlog)\n\n")
	_, _ = builder.WriteString(
		"Families the compiler recognizes but lowers only within an exact envelope. Each row is " +
			"one supported-envelope blocker ranked by sole blockers (cards it is the only blocker " +
			"for); growing the envelope to cover the example wordings unblocks the listed cards. " +
			"This is the effect-family analogue of the unrecognized-condition backlog above.\n\n",
	)
	_, _ = builder.WriteString("| Capability | Supported envelope (blocker) | Affected cards | Sole blockers | Example wordings |\n")
	_, _ = builder.WriteString("| --- | --- | ---: | ---: | --- |\n")
	for index, gap := range gaps {
		if index >= envelopeGapBacklogLimit {
			break
		}
		examples := "-"
		if rows := samples[envelopeGapKey(gap.summary, gap.detail)]; len(rows) > 0 {
			examples = markdownTableCell(strings.Join(rows, "; "))
		}
		_, _ = fmt.Fprintf(
			builder,
			"| %s | %s | %s | %s | %s |\n",
			markdownTableCell(gap.summary),
			markdownTableCell(gap.detail),
			formatCount(gap.affectedCards),
			formatCount(gap.soleBlockerCards),
			examples,
		)
	}
}

func updateReadmeSupport(path string, output report) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading README support summary: %w", err)
	}
	content := string(data)
	if strings.Count(content, readmeSupportStart) != 1 || strings.Count(content, readmeSupportEnd) != 1 {
		return fmt.Errorf("README %s must contain one %s / %s marker pair", path, readmeSupportStart, readmeSupportEnd)
	}
	start := strings.Index(content, readmeSupportStart)
	end := strings.Index(content, readmeSupportEnd)
	if end < start {
		return fmt.Errorf("README %s must contain an ordered %s / %s marker pair", path, readmeSupportStart, readmeSupportEnd)
	}
	start += len(readmeSupportStart)
	replacement := "\n" + supportSummary(output) +
		" See [`supported.md`](./supported.md) and [`unsupported.md`](./unsupported.md) for the complete lists, " +
		"and [`unsupported-reasons.md`](./unsupported-reasons.md) for capability-aware blocker planning.\n"
	content = content[:start] + replacement + content[end:]
	return writeDocumentationFile(path, content)
}

func supportSummary(output report) string {
	percentage := 0.0
	if output.EligibleCount > 0 {
		percentage = 100 * float64(output.GeneratedCount) / float64(output.EligibleCount)
	}
	return fmt.Sprintf(
		"Council4 currently supports **%s of %s cards eligible for paper support (%.1f%%)**. "+
			"The Scryfall Oracle Cards corpus contains %s additional digital, special-format, memorabilia, "+
			"or non-sanctioned-paper records that are excluded from that total.",
		formatCount(output.GeneratedCount),
		formatCount(output.EligibleCount),
		percentage,
		formatCount(output.ExcludedCount),
	)
}

func formatCount(value int) string {
	text := strconv.Itoa(value)
	for i := len(text) - 3; i > 0; i -= 3 {
		text = text[:i] + "," + text[i:]
	}
	return text
}

func markdownInline(text string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\n", " ",
		"\r", " ",
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(text)
}

func markdownTableCell(text string) string {
	return strings.ReplaceAll(markdownInline(text), "|", "\\|")
}

func writeDocumentationFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("creating documentation directory for %s: %w", path, err)
	}
	if existing, err := os.ReadFile(path); err == nil && string(existing) == content {
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading existing documentation %s: %w", path, err)
	}
	//nolint:gosec // The caller explicitly selects documentation output paths through CLI flags.
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("writing documentation %s: %w", path, err)
	}
	return nil
}
