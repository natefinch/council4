package main

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestIsEnvelopeGapDetail(t *testing.T) {
	t.Parallel()
	envelope := []string{
		"the executable source backend supports only exact destruction of one target permanent",
		"the executable source backend supports exact recognized counter placement on one valid target",
		"the executable source backend does not support this group recipient",
		"supports only a single fixed-power/toughness creature token",
	}
	for _, detail := range envelope {
		if !isEnvelopeGapDetail(detail) {
			t.Errorf("isEnvelopeGapDetail(%q) = false, want true", detail)
		}
	}
	notEnvelope := []string{
		"sub-effect — unsupported token creation",
		"structural — per-effect condition spans multiple clauses",
		orderedSequenceUnrecognizedConditionPrefix + "if this spell was kicked",
		"",
	}
	for _, detail := range notEnvelope {
		if isEnvelopeGapDetail(detail) {
			t.Errorf("isEnvelopeGapDetail(%q) = true, want false", detail)
		}
	}
}

func envelopeCard(name, id, summary, detail string, extraSummaries ...string) unsupported {
	diagnostics := []reportDiagnostic{{Summary: summary, Detail: detail}}
	for _, extra := range extraSummaries {
		diagnostics = append(diagnostics, reportDiagnostic{Summary: extra})
	}
	return unsupported{Name: name, ID: id, Diagnostics: diagnostics}
}

func TestAnalyzeEnvelopeGapBacklogRanksAndExcludes(t *testing.T) {
	t.Parallel()
	destroy := "the executable source backend supports only exact destruction of one target permanent"
	creatureCreationDetail := "the executable source backend supports only a single fixed-power/toughness creature with one subtype and at most one color"
	output := report{Unsupported: []unsupported{
		envelopeCard("Destroy1", "d1", "unsupported destroy spell", destroy),
		envelopeCard("Destroy2", "d2", "unsupported destroy spell", destroy),
		// Co-blocked: a second distinct summary excludes it from sole blockers.
		envelopeCard("Destroy3", "d3", "unsupported destroy spell", destroy, "unsupported Oracle construct"),
		envelopeCard("Token1", "t1", "unsupported token creation", creatureCreationDetail),
		// Non-envelope detail must not appear in the backlog.
		envelopeCard("Structural", "s1", "unsupported ordered effect sequence", "sub-effect — unsupported life spell"),
	}}

	gaps := analyzeEnvelopeGapBacklog(output)
	if len(gaps) != 2 {
		t.Fatalf("gap count = %d, want 2: %+v", len(gaps), gaps)
	}
	if gaps[0].summary != "unsupported destroy spell" || gaps[0].affectedCards != 3 || gaps[0].soleBlockerCards != 2 {
		t.Errorf("first gap = %+v, want destroy affected 3 sole 2", gaps[0])
	}
	if gaps[1].summary != "unsupported token creation" || gaps[1].soleBlockerCards != 1 {
		t.Errorf("second gap = %+v, want token sole 1", gaps[1])
	}
}

func TestSpanWording(t *testing.T) {
	t.Parallel()
	full := "Return target creature card from your graveyard to your hand."
	single := &cardgen.ScryfallCard{OracleText: full}
	if got := spanWording(single, shared.Span{End: shared.Position{Offset: len(full)}}); got != full {
		t.Errorf("single-face wording = %q, want %q", got, full)
	}
	// Multi-face card carries per-face text and empty top-level text.
	faceText := "Destroy target creature."
	multi := &cardgen.ScryfallCard{CardFaces: []cardgen.ScryfallCardFace{{OracleText: faceText}}}
	if got := spanWording(multi, shared.Span{End: shared.Position{Offset: len(faceText)}}); got != faceText {
		t.Errorf("multi-face wording = %q, want %q", got, faceText)
	}
	// Out-of-bounds span yields no wording.
	if got := spanWording(single, shared.Span{End: shared.Position{Offset: len(full) + 5}}); got != "" {
		t.Errorf("out-of-bounds wording = %q, want empty", got)
	}
}

func TestWriteEnvelopeGapBacklogRendersTableWithSamples(t *testing.T) {
	t.Parallel()
	detail := "the executable source backend supports only exact return of one target permanent to its owner's hand"
	wording := "return target creature card from your graveyard to your hand."
	output := report{Unsupported: []unsupported{{
		Name: "Reanimate",
		ID:   "r1",
		Diagnostics: []reportDiagnostic{{
			Summary: "unsupported return spell",
			Detail:  detail,
			Span:    shared.Span{End: shared.Position{Offset: len(wording)}},
		}},
	}}}
	results := []result{{card: cardgen.ScryfallCard{ID: "r1", OracleText: wording}}}

	var builder strings.Builder
	writeEnvelopeGapBacklog(&builder, output, results)
	rendered := builder.String()
	for _, wanted := range []string{
		"## Modeled-capability envelope gaps (parameter backlog)",
		"| Capability | Supported envelope (blocker) | Affected cards | Sole blockers | Example wordings |",
		"unsupported return spell",
		"from your graveyard to your hand",
	} {
		if !strings.Contains(rendered, wanted) {
			t.Errorf("rendered output missing %q:\n%s", wanted, rendered)
		}
	}
}

func TestWriteEnvelopeGapBacklogEmptyWhenNoEnvelopeGaps(t *testing.T) {
	t.Parallel()
	output := report{Unsupported: []unsupported{
		{Name: "Other", Diagnostics: []reportDiagnostic{{Summary: "unsupported Oracle construct"}}},
	}}
	var builder strings.Builder
	writeEnvelopeGapBacklog(&builder, output, nil)
	if builder.Len() != 0 {
		t.Errorf("expected no output, got:\n%s", builder.String())
	}
}
