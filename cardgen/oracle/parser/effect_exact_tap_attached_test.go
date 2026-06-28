package parser

import "testing"

// TestParseTapAttachedExact recognizes the attached-recipient tap form "Tap
// enchanted creature." / "Tap enchanted permanent." (Aura) and "Tap equipped
// creature." (Equipment) as exact and records the typed TapAttached marker. The
// recipient is the permanent the source is attached to, so the clause carries no
// target or reference.
func TestParseTapAttachedExact(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Tap enchanted creature.",
		"Tap enchanted permanent.",
		"Tap equipped creature.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			effect := effects[0]
			if effect.Kind != EffectTap {
				t.Fatalf("kind = %q, want EffectTap", effect.Kind)
			}
			if !effect.Exact {
				t.Fatalf("effect not exact: %#v", effect)
			}
			if !effect.TapAttached {
				t.Fatalf("TapAttached = false, want true: %#v", effect)
			}
			if len(effect.Targets) != 0 || len(effect.References) != 0 {
				t.Fatalf("targets=%d references=%d, want zero", len(effect.Targets), len(effect.References))
			}
		})
	}
}

// TestParseUntapAttachedExact recognizes the attached-recipient untap form
// "Untap enchanted creature." / "Untap enchanted permanent." (Aura) and "Untap
// equipped creature." (Equipment) as exact and records the typed UntapAttached
// marker. The recipient is the permanent the source is attached to, so the
// clause carries no target or reference.
func TestParseUntapAttachedExact(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Untap enchanted creature.",
		"Untap enchanted permanent.",
		"Untap equipped creature.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			effect := effects[0]
			if effect.Kind != EffectUntap {
				t.Fatalf("kind = %q, want EffectUntap", effect.Kind)
			}
			if !effect.Exact {
				t.Fatalf("effect not exact: %#v", effect)
			}
			if !effect.UntapAttached {
				t.Fatalf("UntapAttached = false, want true: %#v", effect)
			}
			if len(effect.Targets) != 0 || len(effect.References) != 0 {
				t.Fatalf("targets=%d references=%d, want zero", len(effect.Targets), len(effect.References))
			}
		})
	}
}

// TestParseTapUntapAttachedFailsClosed documents that tap/untap shapes other
// than the bare attached recipient are not recognized as attached tap/untap, so
// lowering fails closed.
func TestParseTapUntapAttachedFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Tap target creature.",
		"Untap target permanent.",
		"Tap all creatures.",
		"Untap all creatures you control.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{InstantOrSorcery: true})
			for _, sentence := range document.Abilities[0].Sentences {
				for i := range sentence.Effects {
					if sentence.Effects[i].TapAttached || sentence.Effects[i].UntapAttached {
						t.Fatalf("source %q recognized as attached tap/untap: %#v", source, sentence.Effects[i])
					}
				}
			}
		})
	}
}
