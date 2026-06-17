package main

import (
	"cmp"
	"slices"
)

const unsupportedReasonLimit = 100

type supportCapabilityID string

const (
	capabilityRecognitionFallback  supportCapabilityID = "recognition-fallback"
	capabilitySharedAbilityContent supportCapabilityID = "shared-ability-content"
	capabilityTriggerPattern       supportCapabilityID = "trigger-pattern"
	capabilityStaticDeclaration    supportCapabilityID = "static-declaration"
	capabilityActivation           supportCapabilityID = "activation"
	capabilityReplacement          supportCapabilityID = "replacement"
	capabilityOther                supportCapabilityID = "other"
)

var supportCapabilityIDs = []supportCapabilityID{
	capabilityRecognitionFallback,
	capabilitySharedAbilityContent,
	capabilityTriggerPattern,
	capabilityStaticDeclaration,
	capabilityActivation,
	capabilityReplacement,
	capabilityOther,
}

// diagnosticCapabilities is intentionally exact: new summaries remain visible
// under other until their capability is chosen deliberately.
var diagnosticCapabilities = map[string]supportCapabilityID{
	"unsupported Oracle construct":                      capabilityRecognitionFallback,
	"unsupported ability word":                          capabilityRecognitionFallback,
	"unsupported unknown ability":                       capabilityRecognitionFallback,
	"unsupported reminder ability":                      capabilityRecognitionFallback,
	"unsupported ordered effect sequence":               capabilitySharedAbilityContent,
	"unsupported spell ability":                         capabilitySharedAbilityContent,
	"unsupported ability content":                       capabilitySharedAbilityContent,
	"unsupported ability modes":                         capabilitySharedAbilityContent,
	"unsupported mana effect":                           capabilitySharedAbilityContent,
	"unsupported enter trigger effect":                  capabilitySharedAbilityContent,
	"unsupported phase/step trigger phrase effect":      capabilitySharedAbilityContent,
	"unsupported triggered ability effect":              capabilitySharedAbilityContent,
	"unsupported damage spell":                          capabilitySharedAbilityContent,
	"unsupported modal ability":                         capabilitySharedAbilityContent,
	"unsupported destroy spell":                         capabilitySharedAbilityContent,
	"unsupported dies trigger effect":                   capabilitySharedAbilityContent,
	"unsupported return spell":                          capabilitySharedAbilityContent,
	"unsupported search effect":                         capabilitySharedAbilityContent,
	"unsupported exile spell":                           capabilitySharedAbilityContent,
	"unsupported power/toughness spell":                 capabilitySharedAbilityContent,
	"unsupported counter placement":                     capabilitySharedAbilityContent,
	"unsupported library placement":                     capabilitySharedAbilityContent,
	"unsupported life spell":                            capabilitySharedAbilityContent,
	"unsupported keyword or ability grant":              capabilitySharedAbilityContent,
	"unsupported keyword or ability loss":               capabilitySharedAbilityContent,
	"unsupported dies trigger body effect":              capabilitySharedAbilityContent,
	"unsupported multiple spell abilities":              capabilitySharedAbilityContent,
	"unsupported counter spell":                         capabilitySharedAbilityContent,
	"unsupported temporary keyword spell":               capabilitySharedAbilityContent,
	"unsupported dies trigger body":                     capabilitySharedAbilityContent,
	"unsupported gain-control spell":                    capabilitySharedAbilityContent,
	"unsupported draw spell":                            capabilitySharedAbilityContent,
	"unsupported draw/discard trigger effect":           capabilitySharedAbilityContent,
	"unsupported discard spell":                         capabilitySharedAbilityContent,
	"unsupported tap spell":                             capabilitySharedAbilityContent,
	"unsupported untap spell":                           capabilitySharedAbilityContent,
	"unsupported mill spell":                            capabilitySharedAbilityContent,
	"unsupported cycling trigger effect":                capabilitySharedAbilityContent,
	"unsupported fight spell":                           capabilitySharedAbilityContent,
	"unsupported group power/toughness spell":           capabilitySharedAbilityContent,
	"unsupported delayed effect":                        capabilitySharedAbilityContent,
	"unsupported investigate spell":                     capabilitySharedAbilityContent,
	"unsupported manifest spell":                        capabilitySharedAbilityContent,
	"unsupported proliferate spell":                     capabilitySharedAbilityContent,
	"unsupported explore spell":                         capabilitySharedAbilityContent,
	"unsupported regenerate spell":                      capabilitySharedAbilityContent,
	"unsupported scry spell":                            capabilitySharedAbilityContent,
	"unsupported triggered ability":                     capabilityTriggerPattern,
	"unsupported enter trigger":                         capabilityTriggerPattern,
	"unsupported phase/step trigger phrase":             capabilityTriggerPattern,
	"unsupported dies trigger":                          capabilityTriggerPattern,
	"unsupported dies trigger phrase":                   capabilityTriggerPattern,
	"unsupported draw/discard trigger":                  capabilityTriggerPattern,
	"unsupported cycling trigger":                       capabilityTriggerPattern,
	"unsupported static ability":                        capabilityStaticDeclaration,
	"unsupported mixed keyword ability":                 capabilityStaticDeclaration,
	"unsupported Enchant ability":                       capabilityStaticDeclaration,
	"unsupported Saga chapter ability":                  capabilityStaticDeclaration,
	"unsupported keyword ability":                       capabilityStaticDeclaration,
	"unsupported parameterized keyword":                 capabilityStaticDeclaration,
	"unsupported Protection ability":                    capabilityStaticDeclaration,
	"unsupported Read ahead ability":                    capabilityStaticDeclaration,
	"unsupported static rule declaration":               capabilityStaticDeclaration,
	"unsupported hand Cycling grant":                    capabilityStaticDeclaration,
	"unsupported static declaration shell":              capabilityStaticDeclaration,
	"unsupported static declaration condition":          capabilityStaticDeclaration,
	"unsupported static declaration duration":           capabilityStaticDeclaration,
	"unsupported static declaration group":              capabilityStaticDeclaration,
	"unsupported static declaration operation":          capabilityStaticDeclaration,
	"unsupported activated ability":                     capabilityActivation,
	"unsupported mana ability":                          capabilityActivation,
	"unsupported activation cost":                       capabilityActivation,
	"unsupported activation timing":                     capabilityActivation,
	"unsupported activation zone":                       capabilityActivation,
	"unsupported activation condition":                  capabilityActivation,
	"unsupported activation references":                 capabilityActivation,
	"unsupported activation modes":                      capabilityActivation,
	"unsupported activation structure":                  capabilityActivation,
	"unsupported activation ability word":               capabilityActivation,
	"unsupported loyalty ability":                       capabilityActivation,
	"unsupported cost":                                  capabilityActivation,
	"unsupported Equip ability":                         capabilityActivation,
	"unsupported mana symbol":                           capabilitySharedAbilityContent,
	"unsupported Cycling ability":                       capabilityActivation,
	"unsupported Mutate ability":                        capabilityActivation,
	"unsupported Ninjutsu ability":                      capabilityActivation,
	"unsupported enters-tapped replacement":             capabilityReplacement,
	"unsupported enters-with-counters replacement":      capabilityReplacement,
	"unsupported damage replacement":                    capabilityReplacement,
	"unsupported conditional enters-tapped replacement": capabilityReplacement,
	"unsupported counter-placement replacement":         capabilityReplacement,
	"unsupported token-creation replacement":            capabilityReplacement,
	"unsupported self zone-destination replacement":     capabilityReplacement,
	"unsupported type line":                             capabilityOther,
	"incomplete executable lowering":                    capabilityOther,
	"unsupported card layout":                           capabilityOther,
	"unsupported package letter":                        capabilityOther,
	"validation failed: oracle-without-abilities":       capabilityOther,
}

type supportAnalysis struct {
	reasons      []unsupportedReason
	capabilities []unsupportedCapability
}

type unsupportedReason struct {
	summary             string
	affectedCards       int
	soleBlockerCards    int
	mostCommonCoBlocker string
}

func (r unsupportedReason) soleBlockerPercentage() float64 {
	if r.affectedCards == 0 {
		return 0
	}
	return 100 * float64(r.soleBlockerCards) / float64(r.affectedCards)
}

type unsupportedCapability struct {
	id                   supportCapabilityID
	affectedCards        int
	fullyUnlockableCards int
	summaries            []string
}

type reasonCounts struct {
	affectedCards    int
	soleBlockerCards int
	coBlockers       map[string]int
}

type capabilityCounts struct {
	affectedCards        int
	fullyUnlockableCards int
	summaries            map[string]bool
}

func analyzeSupport(output report) supportAnalysis {
	reasonCountsBySummary := make(map[string]*reasonCounts)
	capabilityCountsByID := make(map[supportCapabilityID]*capabilityCounts, len(supportCapabilityIDs))
	for _, id := range supportCapabilityIDs {
		capabilityCountsByID[id] = &capabilityCounts{summaries: make(map[string]bool)}
	}

	for _, card := range output.Unsupported {
		summaries := distinctDiagnosticSummaries(card.Diagnostics)
		capabilities := make(map[supportCapabilityID]bool)
		for _, summary := range summaries {
			counts := reasonCountsBySummary[summary]
			if counts == nil {
				counts = &reasonCounts{coBlockers: make(map[string]int)}
				reasonCountsBySummary[summary] = counts
			}
			counts.affectedCards++
			if len(summaries) == 1 {
				counts.soleBlockerCards++
			}
			for _, coBlocker := range summaries {
				if coBlocker != summary {
					counts.coBlockers[coBlocker]++
				}
			}

			id := capabilityForDiagnostic(summary)
			capabilities[id] = true
			capabilityCountsByID[id].summaries[summary] = true
		}
		for id := range capabilities {
			capabilityCountsByID[id].affectedCards++
		}
		if len(capabilities) == 1 {
			for id := range capabilities {
				capabilityCountsByID[id].fullyUnlockableCards++
			}
		}
	}

	return supportAnalysis{
		reasons:      buildUnsupportedReasons(reasonCountsBySummary),
		capabilities: buildUnsupportedCapabilities(capabilityCountsByID),
	}
}

func capabilityForDiagnostic(summary string) supportCapabilityID {
	if id, ok := diagnosticCapabilities[summary]; ok {
		return id
	}
	return capabilityOther
}

func distinctDiagnosticSummaries(diagnostics []reportDiagnostic) []string {
	seen := make(map[string]bool, len(diagnostics))
	summaries := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		if !seen[diagnostic.Summary] {
			seen[diagnostic.Summary] = true
			summaries = append(summaries, diagnostic.Summary)
		}
	}
	slices.Sort(summaries)
	return summaries
}

func buildUnsupportedReasons(countsBySummary map[string]*reasonCounts) []unsupportedReason {
	reasons := make([]unsupportedReason, 0, len(countsBySummary))
	for summary, counts := range countsBySummary {
		reasons = append(reasons, unsupportedReason{
			summary:             summary,
			affectedCards:       counts.affectedCards,
			soleBlockerCards:    counts.soleBlockerCards,
			mostCommonCoBlocker: mostCommonCoBlocker(counts.coBlockers),
		})
	}
	slices.SortFunc(reasons, func(a, b unsupportedReason) int {
		if compared := cmp.Compare(b.affectedCards, a.affectedCards); compared != 0 {
			return compared
		}
		if compared := cmp.Compare(b.soleBlockerCards, a.soleBlockerCards); compared != 0 {
			return compared
		}
		return cmp.Compare(a.summary, b.summary)
	})
	return reasons[:min(len(reasons), unsupportedReasonLimit)]
}

func mostCommonCoBlocker(counts map[string]int) string {
	var best string
	var bestCount int
	for summary, count := range counts {
		if count > bestCount || count == bestCount && summary < best {
			best = summary
			bestCount = count
		}
	}
	return best
}

func buildUnsupportedCapabilities(countsByID map[supportCapabilityID]*capabilityCounts) []unsupportedCapability {
	capabilities := make([]unsupportedCapability, 0, len(supportCapabilityIDs))
	for _, id := range supportCapabilityIDs {
		counts := countsByID[id]
		summaries := make([]string, 0, len(counts.summaries))
		for summary := range counts.summaries {
			summaries = append(summaries, summary)
		}
		slices.Sort(summaries)
		capabilities = append(capabilities, unsupportedCapability{
			id:                   id,
			affectedCards:        counts.affectedCards,
			fullyUnlockableCards: counts.fullyUnlockableCards,
			summaries:            summaries,
		})
	}
	slices.SortFunc(capabilities, func(a, b unsupportedCapability) int {
		if compared := cmp.Compare(b.fullyUnlockableCards, a.fullyUnlockableCards); compared != 0 {
			return compared
		}
		return cmp.Compare(a.id, b.id)
	})
	return capabilities
}
