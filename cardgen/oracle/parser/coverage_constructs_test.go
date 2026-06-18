package parser

import "testing"

func assertComplete(t *testing.T, source string, context Context) {
	t.Helper()
	report := AbilityCoverage(firstAbility(t, source, context))
	if !report.Complete {
		t.Fatalf("ability incomplete for %q: blockers=%v uncovered=%v",
			source, report.Blockers, report.Uncovered)
	}
	if len(report.Uncovered) != 0 {
		t.Errorf("ability for %q has uncovered runs: %v", source, report.Uncovered)
	}
}

func TestAbilityCoverageCoordinatedTriggerSpellListIsComplete(t *testing.T) {
	assertComplete(t,
		"Whenever you cast an instant, sorcery, or Wizard spell, this creature deals 1 damage to any target.",
		Context{})
}

func TestAbilityCoverageCoordinatedConditionSubjectListIsComplete(t *testing.T) {
	assertComplete(t,
		"Target creature gets -4/-0 until end of turn. If you control a Fish, Octopus, Otter, Seal, Serpent, or Whale, draw a card.",
		Context{InstantOrSorcery: true})
}

func TestAbilityCoverageForEachIterationPrefixIsComplete(t *testing.T) {
	assertComplete(t,
		"Whenever this creature attacks, for each token you control, create a 1/1 white Rabbit creature token.",
		Context{})
}

func TestAbilityCoverageReflexiveTriggerPreambleIsComplete(t *testing.T) {
	assertComplete(t,
		"Create two 1/1 blue Faerie creature tokens with flying. When you do, tap target creature an opponent controls.",
		Context{InstantOrSorcery: true})
}

func TestAbilityCoverageDelayedTriggerPreambleIsComplete(t *testing.T) {
	assertComplete(t,
		"All creatures get -2/-2 until end of turn. Whenever a creature dies this turn, you gain 1 life.",
		Context{InstantOrSorcery: true})
}

// TestAbilityCoverageConstructSpansDoNotOverCredit guards that a recognized
// construct span never completes an ability whose effect verb itself is
// unrepresented: the coordinated type list "creature or artifact" is credited,
// but the unknown "Goad" effect leaves its verb and target uncovered.
func TestAbilityCoverageConstructSpansDoNotOverCredit(t *testing.T) {
	report := AbilityCoverage(firstAbility(t, "Goad target creature or artifact.", Context{InstantOrSorcery: true}))
	if report.Complete {
		t.Fatal("unknown effect with a coordinated type list reported complete")
	}
}
