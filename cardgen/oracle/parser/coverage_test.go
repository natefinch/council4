package parser

import "testing"

func firstAbility(t *testing.T, source string, context Context) *Ability {
	t.Helper()
	document, _ := Parse(source, context)
	if len(document.Abilities) == 0 {
		t.Fatalf("Parse(%q) produced no abilities", source)
	}
	return &document.Abilities[0]
}

func TestAbilityCoverageKeywordIsComplete(t *testing.T) {
	report := AbilityCoverage(firstAbility(t, "Flying", Context{}))
	if !report.Complete {
		t.Fatalf("keyword ability incomplete: blockers=%v uncovered=%v", report.Blockers, report.Uncovered)
	}
	if len(report.Uncovered) != 0 {
		t.Errorf("keyword ability has uncovered runs: %v", report.Uncovered)
	}
	if len(report.Components) != 0 {
		t.Errorf("keyword ability has uncovered components: %v", report.Components)
	}
}

func TestAbilityCoverageSimpleSpellEffectIsComplete(t *testing.T) {
	report := AbilityCoverage(firstAbility(t, "Draw a card.", Context{InstantOrSorcery: true}))
	if !report.Complete {
		t.Fatalf("simple spell effect incomplete: blockers=%v uncovered=%v", report.Blockers, report.Uncovered)
	}
	if report.ResolvingEffects != 1 || report.ExactEffects != 1 {
		t.Errorf("resolving=%d exact=%d, want 1 and 1", report.ResolvingEffects, report.ExactEffects)
	}
}

func TestAbilityCoverageUnknownEffectIsIncomplete(t *testing.T) {
	report := AbilityCoverage(firstAbility(t, "Goad target creature.", Context{InstantOrSorcery: true}))
	if report.Complete {
		t.Fatal("unknown effect ability reported complete")
	}
	if len(report.Uncovered) == 0 {
		t.Fatal("unknown effect ability reported no uncovered runs")
	}
	if got, want := report.Uncovered[0].Text, "Goad target creature"; got != want {
		t.Errorf("uncovered run text = %q, want %q", got, want)
	}
	if len(report.Components) != 1 {
		t.Fatalf("components = %v, want exactly one", report.Components)
	}
	if got, want := report.Components[0].Text, "Goad target creature."; got != want {
		t.Errorf("uncovered component text = %q, want %q", got, want)
	}
	if got := report.Components[0].Blocker; got != CoverageBlockerEffect {
		t.Errorf("uncovered component blocker = %q, want %q", got, CoverageBlockerEffect)
	}
}

func TestAbilityCoverageTriggerWithExactEffectIsComplete(t *testing.T) {
	report := AbilityCoverage(firstAbility(t, "Whenever this creature attacks, draw a card.", Context{}))
	if !report.Complete {
		t.Fatalf("recognized trigger + exact effect incomplete: blockers=%v uncovered=%v",
			report.Blockers, report.Uncovered)
	}
	if report.ResolvingEffects != 1 || report.ExactEffects != 1 {
		t.Errorf("resolving=%d exact=%d, want 1 and 1", report.ResolvingEffects, report.ExactEffects)
	}
}

func TestDocumentCoverageAggregatesAbilities(t *testing.T) {
	document, _ := Parse("Flying\nGoad target creature.", Context{})
	report := DocumentCoverage(document)
	if report.Complete {
		t.Fatal("document with an unknown effect reported complete")
	}
	if len(report.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(report.Abilities))
	}
	if len(report.Components) == 0 {
		t.Fatal("document reported no uncovered components")
	}
}

// orphanProbe reconstructs an instant/sorcery ability so the orphan-clause
// regression cases below share one parse path.
func orphanProbe(t *testing.T, source string) AbilityCoverageReport {
	t.Helper()
	return AbilityCoverage(firstAbility(t, source, Context{InstantOrSorcery: true}))
}

func TestAbilityCoverageTrailingOrphanClauseIsIncomplete(t *testing.T) {
	cases := []struct {
		source    string
		uncovered string
	}{
		{"Draw a card and goad target creature.", "and goad target creature"},
		{"Draw a card, then goad target creature.", "then goad target creature"},
		{"Draw a card and frobnicate.", "and frobnicate"},
	}
	for _, test := range cases {
		report := orphanProbe(t, test.source)
		if report.Complete {
			t.Errorf("%q reported complete; a trailing unrepresented clause must stay uncovered", test.source)
		}
		if report.ExactEffects != 0 {
			t.Errorf("%q exact=%d, want 0 (sentence has an unrepresented clause)", test.source, report.ExactEffects)
		}
		if len(report.Uncovered) == 0 || report.Uncovered[0].Text != test.uncovered {
			t.Errorf("%q uncovered = %v, want first run %q", test.source, report.Uncovered, test.uncovered)
		}
	}
}

func TestAbilityCoverageLeadingOrphanClauseIsIncomplete(t *testing.T) {
	report := orphanProbe(t, "Goad target creature and draw a card.")
	if report.Complete {
		t.Fatal("leading unrepresented clause reported complete")
	}
	if report.ExactEffects != 0 {
		t.Errorf("exact=%d, want 0", report.ExactEffects)
	}
	if len(report.Uncovered) == 0 || report.Uncovered[0].Text != "Goad target creature and" {
		t.Errorf("uncovered = %v, want first run %q", report.Uncovered, "Goad target creature and")
	}
}

func TestAbilityCoverageLeadingOrphanViaCommaOrThenIsIncomplete(t *testing.T) {
	cases := []struct {
		source    string
		uncovered string
	}{
		{"Goad target creature, then draw a card.", "Goad target creature"},
		{"Goad target creature, draw a card.", "Goad target creature"},
		{"Frobnicate, then draw a card.", "Frobnicate"},
	}
	for _, test := range cases {
		report := orphanProbe(t, test.source)
		if report.Complete {
			t.Errorf("%q reported complete; a leading unrepresented clause joined by comma/then must stay uncovered", test.source)
		}
		if report.ExactEffects != 0 {
			t.Errorf("%q exact=%d, want 0 (sentence has an unrepresented leading clause)", test.source, report.ExactEffects)
		}
		if len(report.Uncovered) == 0 || report.Uncovered[0].Text != test.uncovered {
			t.Errorf("%q uncovered = %v, want first run %q", test.source, report.Uncovered, test.uncovered)
		}
	}
}

func TestAbilityCoverageTriggerWithTrailingOrphanIsIncomplete(t *testing.T) {
	report := AbilityCoverage(firstAbility(t,
		"Whenever this creature attacks, draw a card and goad target creature.", Context{}))
	if report.Complete {
		t.Fatal("trigger with a trailing unrepresented clause reported complete")
	}
	if len(report.Uncovered) == 0 || report.Uncovered[0].Text != "and goad target creature" {
		t.Errorf("uncovered = %v, want first run %q", report.Uncovered, "and goad target creature")
	}
}

func TestAbilityCoverageCompoundRecognizedClausesStayComplete(t *testing.T) {
	cases := []struct {
		source string
		exact  int
	}{
		{"Draw a card and gain 2 life.", 2},
		{"Destroy all artifacts and enchantments.", 1},
	}
	for _, test := range cases {
		report := orphanProbe(t, test.source)
		if !report.Complete {
			t.Errorf("%q reported incomplete: uncovered=%v", test.source, report.Uncovered)
		}
		if report.ExactEffects != test.exact {
			t.Errorf("%q exact=%d, want %d", test.source, report.ExactEffects, test.exact)
		}
	}
}
