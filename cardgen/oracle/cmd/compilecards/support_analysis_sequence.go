package main

import (
	"cmp"
	"slices"
	"strings"
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
			category := normalizeOrderedSequenceCategory(diagnostic.Detail)
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

// orderedSequenceUnrecognizedConditionPrefix mirrors the closed diagnostic
// detail the cardgen lowering emits when an ordered sequence carries a per-effect
// "if <condition>" whose predicate the compiler never recognized
// (ConditionPredicateUnsupported). The lowering appends the recognized condition
// wording after this prefix so the report can rank which unrecognized conditions
// block the most cards. Keep this in sync with effectGateCategoryUnrecognizedPrefix
// in cardgen/lower_spell_sequence.go (guarded by TestNormalizeOrderedSequenceCategory).
const (
	orderedSequenceUnrecognizedConditionPrefix   = "structural — per-effect condition unrecognized: "
	orderedSequenceUnrecognizedConditionCategory = "structural — per-effect condition unrecognized"
)

// normalizeOrderedSequenceCategory collapses the per-card unrecognized-condition
// details (which embed the specific condition wording) into one stable category
// so the sub-category breakdown stays legible. All other details are categories
// already and pass through unchanged.
func normalizeOrderedSequenceCategory(detail string) string {
	if strings.HasPrefix(detail, orderedSequenceUnrecognizedConditionPrefix) {
		return orderedSequenceUnrecognizedConditionCategory
	}
	return detail
}

// conditionRecognitionEntry ranks one unrecognized per-effect condition wording
// by how many cards it blocks within ordered sequences.
type conditionRecognitionEntry struct {
	condition        string
	affectedCards    int
	soleBlockerCards int
}

// analyzeConditionRecognitionBacklog ranks the distinct unrecognized per-effect
// condition wordings that block ordered-sequence lowering. It is the actionable
// drill-down beneath the "structural — per-effect condition unrecognized"
// sub-category: each entry names a condition the compiler does not yet recognize
// and how many cards recognizing it would unblock. A card contributes once per
// distinct wording it carries, and to the sole-blocker count only when the
// ordered-sequence reason is its only distinct diagnostic summary.
func analyzeConditionRecognitionBacklog(output report) []conditionRecognitionEntry {
	countsByCondition := make(map[string]*orderedSequenceCounts)
	for _, card := range output.Unsupported {
		summaries := distinctDiagnosticSummaries(card.Diagnostics)
		if !slices.Contains(summaries, orderedSequenceReasonSummary) {
			continue
		}
		soleBlocker := len(summaries) == 1
		conditions := make(map[string]bool)
		for _, diagnostic := range card.Diagnostics {
			if diagnostic.Summary != orderedSequenceReasonSummary {
				continue
			}
			wording, ok := strings.CutPrefix(diagnostic.Detail, orderedSequenceUnrecognizedConditionPrefix)
			if !ok || wording == "" {
				continue
			}
			conditions[wording] = true
		}
		for condition := range conditions {
			counts := countsByCondition[condition]
			if counts == nil {
				counts = &orderedSequenceCounts{}
				countsByCondition[condition] = counts
			}
			counts.affectedCards++
			if soleBlocker {
				counts.soleBlockerCards++
			}
		}
	}

	result := make([]conditionRecognitionEntry, 0, len(countsByCondition))
	for condition, counts := range countsByCondition {
		result = append(result, conditionRecognitionEntry{
			condition:        condition,
			affectedCards:    counts.affectedCards,
			soleBlockerCards: counts.soleBlockerCards,
		})
	}
	slices.SortFunc(result, func(a, b conditionRecognitionEntry) int {
		if compared := cmp.Compare(b.soleBlockerCards, a.soleBlockerCards); compared != 0 {
			return compared
		}
		if compared := cmp.Compare(b.affectedCards, a.affectedCards); compared != 0 {
			return compared
		}
		return cmp.Compare(a.condition, b.condition)
	})
	return result
}
