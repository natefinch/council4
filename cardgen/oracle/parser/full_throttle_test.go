package parser

import "testing"

func TestParseEachCombatThisTurnDelayedUntap(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"At the beginning of each combat this turn, untap all creatures that attacked this turn.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || document.Abilities[0].Kind != AbilitySpell {
		t.Fatalf("abilities = %#v, want one delayed-trigger spell effect", document.Abilities)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectDelayedTrigger || effect.DelayedTriggerAbility == nil || effect.DelayedTriggerOneShot {
		t.Fatalf("effect = %#v, want repeating delayed trigger", effect)
	}
	inner, innerDiagnostics := effect.DelayedTriggerAbility.Inner()
	if len(innerDiagnostics) != 0 {
		t.Fatalf("inner diagnostics = %#v", innerDiagnostics)
	}
	untap := inner.Abilities[0].Sentences[0].Effects[0]
	if untap.Kind != EffectUntap || !untap.UntapAttackedThisTurn || !untap.Exact {
		t.Fatalf("inner untap = %#v, want typed attacked-this-turn group", untap)
	}
}
