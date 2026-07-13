package parser

import "testing"

func TestParseCounterThatSpellOrAbilityExact(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever this creature becomes the target of a spell or ability, counter that spell or ability.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectCounter || !effect.Exact ||
		effect.Context != EffectContextController || len(effect.Targets) != 0 {
		t.Fatalf("effect = %#v", effect)
	}
	if !effect.CounterTriggeringStackObject {
		t.Fatalf("effect = %#v, want triggering-stack counter marker", effect)
	}
}

// TestParseCounterThatSpellExact proves the singular "counter that spell." body
// of a spell-cast trigger (Boromir, Warden of the Tower; Jin-Gitaxias, Progress
// Tyrant; Vexing Bauble) is typed as the exact reference-form counter that names
// the triggering stack object, mirroring the "or ability" form so no bare "that
// spell" target reaches the strict backend.
func TestParseCounterThatSpellExact(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever an opponent casts a spell, if no mana was spent to cast it, counter that spell.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectCounter || !effect.Exact ||
		effect.Context != EffectContextController || len(effect.Targets) != 0 {
		t.Fatalf("effect = %#v", effect)
	}
	if !effect.CounterTriggeringStackObject {
		t.Fatalf("effect = %#v, want triggering-stack counter marker", effect)
	}
}
