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
		if err := writeUnsupportedReasonsMarkdown(cfg.unsupportedReasonsPath, output); err != nil {
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

func writeUnsupportedReasonsMarkdown(path string, output report) error {
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
	return writeDocumentationFile(path, builder.String())
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
