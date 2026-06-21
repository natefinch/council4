package main

import (
	"cmp"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// envelopeGap ranks one modeled-capability "supported envelope" blocker by how
// many cards it blocks. A modeled family (damage spell, return spell, token
// creation, counter placement, activation, …) is one the compiler recognizes
// but lowers only within an exact envelope; a card outside that envelope fails
// closed with a detail naming the envelope. Ranking these by sole-blocker count
// turns "which parameter should the envelope grow to next?" into data — the
// effect-family analogue of the unrecognized-condition recognition backlog.
type envelopeGap struct {
	summary          string
	detail           string
	affectedCards    int
	soleBlockerCards int
	samples          []string
}

// envelopeGapBacklogLimit caps the rendered backlog so the planning document
// stays focused on the highest-leverage envelope extensions.
const envelopeGapBacklogLimit = 40

// envelopeGapSampleLimit caps the example wordings shown per envelope gap.
const envelopeGapSampleLimit = 3

// isEnvelopeGapDetail reports whether a diagnostic detail describes a modeled
// capability's supported envelope (the family is recognized but the specific
// shape/parameter is outside what the backend lowers). It deliberately keys off
// the envelope phrasing the lowering emits rather than a hard-coded summary
// list, so new modeled families appear automatically. Structural/aggregate
// reasons (ordered-sequence "sub-effect —"/"structural —" categories,
// unrecognized constructs) use other phrasings and are excluded.
func isEnvelopeGapDetail(detail string) bool {
	lower := strings.ToLower(detail)
	return strings.Contains(lower, "supports only") ||
		strings.Contains(lower, "supports exact") ||
		strings.Contains(lower, "only exact") ||
		strings.Contains(lower, "does not support")
}

// envelopeGapKey identifies one (summary, envelope-detail) blocker.
func envelopeGapKey(summary, detail string) string {
	return summary + "\x00" + detail
}

// analyzeEnvelopeGapBacklog ranks modeled-capability envelope gaps by how many
// cards each sole-blocks (a card whose only distinct diagnostic summary is the
// one carrying the envelope detail), then by affected cards. A card contributes
// once per distinct (summary, envelope-detail) it carries.
func analyzeEnvelopeGapBacklog(output report) []envelopeGap {
	type counts struct {
		affected int
		sole     int
		summary  string
		detail   string
	}
	byKey := make(map[string]*counts)
	for _, card := range output.Unsupported {
		summaries := distinctDiagnosticSummaries(card.Diagnostics)
		sole := len(summaries) == 1
		keys := make(map[string]bool)
		for _, diagnostic := range card.Diagnostics {
			if !isEnvelopeGapDetail(diagnostic.Detail) {
				continue
			}
			key := envelopeGapKey(diagnostic.Summary, diagnostic.Detail)
			if keys[key] {
				continue
			}
			keys[key] = true
			entry := byKey[key]
			if entry == nil {
				entry = &counts{summary: diagnostic.Summary, detail: diagnostic.Detail}
				byKey[key] = entry
			}
			entry.affected++
			if sole {
				entry.sole++
			}
		}
	}

	result := make([]envelopeGap, 0, len(byKey))
	for _, entry := range byKey {
		result = append(result, envelopeGap{
			summary:          entry.summary,
			detail:           entry.detail,
			affectedCards:    entry.affected,
			soleBlockerCards: entry.sole,
		})
	}
	slices.SortFunc(result, func(a, b envelopeGap) int {
		if compared := cmp.Compare(b.soleBlockerCards, a.soleBlockerCards); compared != 0 {
			return compared
		}
		if compared := cmp.Compare(b.affectedCards, a.affectedCards); compared != 0 {
			return compared
		}
		if compared := cmp.Compare(a.summary, b.summary); compared != 0 {
			return compared
		}
		return cmp.Compare(a.detail, b.detail)
	})
	return result
}

// collectEnvelopeSamples gathers up to envelopeGapSampleLimit distinct example
// wordings for each envelope gap by slicing the offending clause out of the
// card's Oracle text using the diagnostic span. The wording is illustrative
// metadata only; it lets a reader see the actual parameter variety behind an
// envelope detail (e.g. "from your graveyard", "one or two target") without
// opening each card.
func collectEnvelopeSamples(output report, results []result) map[string][]string {
	byID := make(map[string]*cardgen.ScryfallCard, len(results))
	for i := range results {
		byID[results[i].card.ID] = &results[i].card
	}
	samples := make(map[string][]string)
	seen := make(map[string]map[string]bool)
	for _, card := range output.Unsupported {
		source := byID[card.ID]
		if source == nil {
			continue
		}
		for _, diagnostic := range card.Diagnostics {
			if !isEnvelopeGapDetail(diagnostic.Detail) {
				continue
			}
			key := envelopeGapKey(diagnostic.Summary, diagnostic.Detail)
			if len(samples[key]) >= envelopeGapSampleLimit {
				continue
			}
			wording := spanWording(source, diagnostic.Span)
			if wording == "" {
				continue
			}
			if seen[key] == nil {
				seen[key] = make(map[string]bool)
			}
			if seen[key][wording] {
				continue
			}
			seen[key][wording] = true
			samples[key] = append(samples[key], wording)
		}
	}
	return samples
}

// spanWording returns the normalized Oracle-text substring identified by span.
// It indexes into the single-face Oracle text when present, otherwise the first
// card face whose text covers the span (multi-face cards carry per-face text and
// empty top-level text). It returns "" when no text covers the span.
func spanWording(card *cardgen.ScryfallCard, span shared.Span) string {
	candidates := make([]string, 0, 1+len(card.CardFaces))
	if card.OracleText != "" {
		candidates = append(candidates, card.OracleText)
	}
	for _, face := range card.CardFaces {
		if face.OracleText != "" {
			candidates = append(candidates, face.OracleText)
		}
	}
	start, end := span.Start.Offset, span.End.Offset
	if start < 0 || start > end {
		return ""
	}
	for _, text := range candidates {
		if end <= len(text) {
			return normalizeWording(text[start:end])
		}
	}
	return ""
}

// normalizeWording collapses internal whitespace and truncates a sample wording
// so the planning table stays legible.
func normalizeWording(text string) string {
	collapsed := strings.Join(strings.Fields(text), " ")
	const maxWording = 90
	if len(collapsed) > maxWording {
		collapsed = strings.TrimSpace(collapsed[:maxWording]) + "…"
	}
	return collapsed
}
