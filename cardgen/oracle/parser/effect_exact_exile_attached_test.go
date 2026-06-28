package parser

import "testing"

// TestParseExileAttachedExact recognizes the attached-recipient exile form
// "Exile enchanted creature." (Aura) and "Exile equipped creature." (Equipment)
// as exact and records the typed ExileAttached marker. The recipient is the
// permanent the source is attached to, so the clause carries no target or
// reference.
func TestParseExileAttachedExact(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Exile enchanted creature.",
		"Exile equipped creature.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			effect := effects[0]
			if effect.Kind != EffectExile {
				t.Fatalf("kind = %q, want EffectExile", effect.Kind)
			}
			if !effect.Exact {
				t.Fatalf("effect not exact: %#v", effect)
			}
			if !effect.ExileAttached {
				t.Fatalf("ExileAttached = false, want true: %#v", effect)
			}
			if len(effect.Targets) != 0 || len(effect.References) != 0 {
				t.Fatalf("targets=%d references=%d, want zero", len(effect.Targets), len(effect.References))
			}
		})
	}
}

// TestParseExileAttachedFailsClosed documents that exile shapes other than the
// bare attached-creature recipient are not recognized as attached exiles, so
// lowering fails closed.
func TestParseExileAttachedFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Exile target creature.",
		"Exile enchanted permanent.",
		"Exile enchanted creature and all Auras attached to it.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{InstantOrSorcery: true})
			for _, sentence := range document.Abilities[0].Sentences {
				for i := range sentence.Effects {
					if sentence.Effects[i].ExileAttached {
						t.Fatalf("source %q recognized as attached exile: %#v", source, sentence.Effects[i])
					}
				}
			}
		})
	}
}
