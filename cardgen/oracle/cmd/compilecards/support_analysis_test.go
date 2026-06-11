package main

import (
	"fmt"
	"slices"
	"testing"
)

func TestAnalyzeSupportCountsReasonsAndCapabilities(t *testing.T) {
	t.Parallel()
	output := report{Unsupported: []unsupported{
		cardWithSummaries("unsupported Oracle construct", "unsupported Oracle construct", "unsupported spell ability"),
		cardWithSummaries("unsupported spell ability"),
		cardWithSummaries("unsupported static ability"),
		cardWithSummaries("new diagnostic"),
	}}

	got := analyzeSupport(output)
	if len(got.reasons) != 4 {
		t.Fatalf("reason count = %d, want 4", len(got.reasons))
	}
	wantReasons := []unsupportedReason{
		{
			summary:             "unsupported spell ability",
			affectedCards:       2,
			soleBlockerCards:    1,
			mostCommonCoBlocker: "unsupported Oracle construct",
		},
		{summary: "new diagnostic", affectedCards: 1, soleBlockerCards: 1},
		{summary: "unsupported static ability", affectedCards: 1, soleBlockerCards: 1},
		{
			summary:             "unsupported Oracle construct",
			affectedCards:       1,
			mostCommonCoBlocker: "unsupported spell ability",
		},
	}
	if !slices.Equal(got.reasons, wantReasons) {
		t.Fatalf("reasons = %#v, want %#v", got.reasons, wantReasons)
	}

	wantCapabilities := []unsupportedCapability{
		{
			id:                   capabilityOther,
			affectedCards:        1,
			fullyUnlockableCards: 1,
			summaries:            []string{"new diagnostic"},
		},
		{
			id:                   capabilitySharedAbilityContent,
			affectedCards:        2,
			fullyUnlockableCards: 1,
			summaries:            []string{"unsupported spell ability"},
		},
		{
			id:                   capabilityStaticDeclaration,
			affectedCards:        1,
			fullyUnlockableCards: 1,
			summaries:            []string{"unsupported static ability"},
		},
		{
			id: capabilityActivation,
		},
		{
			id:            capabilityRecognitionFallback,
			affectedCards: 1,
			summaries:     []string{"unsupported Oracle construct"},
		},
		{id: capabilityReplacement},
		{id: capabilityTriggerPattern},
	}
	if !slices.EqualFunc(got.capabilities, wantCapabilities, equalUnsupportedCapability) {
		t.Fatalf("capabilities = %#v, want %#v", got.capabilities, wantCapabilities)
	}
}

func TestAnalyzeSupportSortsReasonsAndLimitsOutput(t *testing.T) {
	t.Parallel()
	output := report{Unsupported: []unsupported{
		cardWithSummaries("beta", "gamma"),
		cardWithSummaries("alpha", "gamma"),
		cardWithSummaries("beta"),
	}}
	for index := range unsupportedReasonLimit {
		output.Unsupported = append(output.Unsupported, cardWithSummaries(fmt.Sprintf("reason %03d", index)))
	}

	got := analyzeSupport(output).reasons
	if len(got) != unsupportedReasonLimit {
		t.Fatalf("reason count = %d, want %d", len(got), unsupportedReasonLimit)
	}
	for index, want := range []string{"beta", "gamma", "reason 000"} {
		if got[index].summary != want {
			t.Fatalf("reason %d = %q, want %q", index, got[index].summary, want)
		}
	}
	if got[1].mostCommonCoBlocker != "alpha" {
		t.Fatalf("gamma co-blocker = %q, want alpha", got[1].mostCommonCoBlocker)
	}
}

func TestUnsupportedReasonSoleBlockerPercentage(t *testing.T) {
	t.Parallel()
	if got := (unsupportedReason{affectedCards: 4, soleBlockerCards: 1}).soleBlockerPercentage(); got != 25 {
		t.Fatalf("percentage = %v, want 25", got)
	}
	if got := (unsupportedReason{}).soleBlockerPercentage(); got != 0 {
		t.Fatalf("zero percentage = %v, want 0", got)
	}
}

func cardWithSummaries(summaries ...string) unsupported {
	diagnostics := make([]reportDiagnostic, len(summaries))
	for index, summary := range summaries {
		diagnostics[index].Summary = summary
	}
	return unsupported{Diagnostics: diagnostics}
}

func equalUnsupportedCapability(a, b unsupportedCapability) bool {
	return a.id == b.id &&
		a.affectedCards == b.affectedCards &&
		a.fullyUnlockableCards == b.fullyUnlockableCards &&
		slices.Equal(a.summaries, b.summaries)
}
