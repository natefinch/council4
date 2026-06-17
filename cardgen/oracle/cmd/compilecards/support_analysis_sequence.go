package main

import (
	"cmp"
	"slices"
)

// orderedSequenceReasonSummary is the diagnostic summary emitted whenever an
// ordered effect sequence cannot be lowered. Its per-diagnostic Detail carries
// the specific blocker category (see unsupportedEffectSequenceDiagnostic).
const orderedSequenceReasonSummary = "unsupported ordered effect sequence"

// orderedSequenceCategory aggregates how many cards a single ordered-sequence
// blocker category affects, and how many of those cards it sole-blocks.
type orderedSequenceCategory struct {
	category         string
	affectedCards    int
	soleBlockerCards int
}

type orderedSequenceCounts struct {
	affectedCards    int
	soleBlockerCards int
}

// analyzeOrderedSequenceCategories breaks the otherwise-opaque
// "unsupported ordered effect sequence" reason into its constituent blocker
// categories. A card contributes to a category's affected count once per
// distinct category it carries, and to the sole-blocker count only when the
// ordered-sequence reason is the card's only distinct diagnostic summary.
func analyzeOrderedSequenceCategories(output report) []orderedSequenceCategory {
	countsByCategory := make(map[string]*orderedSequenceCounts)
	for _, card := range output.Unsupported {
		summaries := distinctDiagnosticSummaries(card.Diagnostics)
		if !slices.Contains(summaries, orderedSequenceReasonSummary) {
			continue
		}
		soleBlocker := len(summaries) == 1
		categories := make(map[string]bool)
		for _, diagnostic := range card.Diagnostics {
			if diagnostic.Summary != orderedSequenceReasonSummary {
				continue
			}
			category := diagnostic.Detail
			if category == "" {
				category = "(uncategorized)"
			}
			categories[category] = true
		}
		for category := range categories {
			counts := countsByCategory[category]
			if counts == nil {
				counts = &orderedSequenceCounts{}
				countsByCategory[category] = counts
			}
			counts.affectedCards++
			if soleBlocker {
				counts.soleBlockerCards++
			}
		}
	}

	result := make([]orderedSequenceCategory, 0, len(countsByCategory))
	for category, counts := range countsByCategory {
		result = append(result, orderedSequenceCategory{
			category:         category,
			affectedCards:    counts.affectedCards,
			soleBlockerCards: counts.soleBlockerCards,
		})
	}
	slices.SortFunc(result, func(a, b orderedSequenceCategory) int {
		if compared := cmp.Compare(b.soleBlockerCards, a.soleBlockerCards); compared != 0 {
			return compared
		}
		if compared := cmp.Compare(b.affectedCards, a.affectedCards); compared != 0 {
			return compared
		}
		return cmp.Compare(a.category, b.category)
	})
	return result
}
