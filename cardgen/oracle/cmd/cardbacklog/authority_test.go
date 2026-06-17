package main

import (
	"testing"
)

// TestApplyAuthorityRoutesOnReport verifies that generated membership is taken
// from the compile report keyed by id, not from the per-card recompile: a card
// the report lists as unsupported is marked ungenerated and carries the report's
// distinct diagnostic summaries, while a card absent from the report is generated.
func TestApplyAuthorityRoutesOnReport(t *testing.T) {
	outcomes := []cardOutcome{
		{index: 0, id: "gen", name: "Gen", eligible: true, perCardGenerated: true},
		{index: 1, id: "unsup", name: "Unsup", eligible: true, perCardGenerated: false},
	}
	authority := compileAuthority{
		unsupported: map[string][]string{"unsup": {"unsupported damage spell"}},
		excluded:    map[string]bool{},
	}

	rec := applyAuthority(outcomes, authority)

	if !rec.ok() {
		t.Fatalf("expected clean reconciliation, got %+v", rec)
	}
	if !outcomes[0].generated {
		t.Error("card absent from report should be generated")
	}
	if outcomes[1].generated {
		t.Error("card in report.unsupported should be ungenerated")
	}
	if got := outcomes[1].loweringSummaries; len(got) != 1 || got[0] != "unsupported damage spell" {
		t.Errorf("lowering summaries = %v, want [unsupported damage spell]", got)
	}
}

// TestApplyAuthorityGuardDetectsDivergence proves the reconciliation guard is
// non-tautological: when compilecards demotes a card (lists its id as
// unsupported) that cardbacklog's independent per-card recompile still considers
// generated, applyAuthority routes it out of supported AND records a generated
// divergence. A tautological check could never see this; this test would fail if
// the guard were removed.
func TestApplyAuthorityGuardDetectsDivergence(t *testing.T) {
	outcomes := []cardOutcome{
		{
			index: 0, id: "collide", name: "Collide", eligible: true,
			parserComplete:   true,
			perCardGenerated: true, // per-card recompile: clean
		},
	}
	// compilecards demoted "collide" through a collision pass: it is in the
	// report's unsupported set with no lowering diagnostic of its own.
	authority := compileAuthority{
		unsupported: map[string][]string{"collide": {}},
		excluded:    map[string]bool{},
	}

	rec := applyAuthority(outcomes, authority)

	if rec.ok() {
		t.Fatal("guard failed to detect divergence between report and per-card recompile")
	}
	if len(rec.generatedDivergences) != 1 {
		t.Fatalf("generated divergences = %d, want 1: %+v", len(rec.generatedDivergences), rec.generatedDivergences)
	}
	d := rec.generatedDivergences[0]
	if d.ID != "collide" || d.PerCard != "generated" || d.Authoritative != "ungenerated" {
		t.Errorf("divergence = %+v, want collide per-card=generated authoritative=ungenerated", d)
	}
	if outcomes[0].generated {
		t.Error("demoted card must not be routed as generated")
	}

	// And it routes into the lowering queue under the synthetic collision reason.
	output := buildReport(outcomes)
	if output.SupportedCards != 0 || output.LoweringBacklog != 1 {
		t.Fatalf("demoted card mis-routed: supported=%d lowering=%d", output.SupportedCards, output.LoweringBacklog)
	}
	if got := output.LoweringQueue[0].Reason; got != reasonCollisionOrParseRejection {
		t.Errorf("reason = %q, want %q", got, reasonCollisionOrParseRejection)
	}
}

// TestApplyAuthorityExclusionConflict verifies a card our policy treats as
// eligible but the report marks excluded is flagged as an exclusion conflict and
// kept out of supported.
func TestApplyAuthorityExclusionConflict(t *testing.T) {
	outcomes := []cardOutcome{
		{index: 0, id: "x", name: "X", eligible: true, perCardGenerated: true},
	}
	authority := compileAuthority{
		unsupported: map[string][]string{},
		excluded:    map[string]bool{"x": true},
	}

	rec := applyAuthority(outcomes, authority)

	if rec.ok() {
		t.Fatal("expected exclusion conflict to be recorded")
	}
	if len(rec.exclusionConflicts) != 1 {
		t.Fatalf("exclusion conflicts = %d, want 1", len(rec.exclusionConflicts))
	}
	if outcomes[0].generated {
		t.Error("report-excluded card must not be generated")
	}
}

// TestApplyAuthorityMissingFromCorpus verifies report ids never seen in the
// streamed corpus are surfaced rather than ignored.
func TestApplyAuthorityMissingFromCorpus(t *testing.T) {
	outcomes := []cardOutcome{
		{index: 0, id: "present", name: "Present", eligible: true, perCardGenerated: false},
	}
	authority := compileAuthority{
		unsupported: map[string][]string{
			"present": {"reason"},
			"ghost":   {"reason"},
		},
		excluded: map[string]bool{},
	}

	rec := applyAuthority(outcomes, authority)

	if len(rec.missingFromCorpus) != 1 || rec.missingFromCorpus[0] != "ghost" {
		t.Fatalf("missing-from-corpus = %v, want [ghost]", rec.missingFromCorpus)
	}
}
