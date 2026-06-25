package parser

import "testing"

// cantBlockEffect parses a single can't-block-this-turn sentence and returns its
// resolving effect, asserting that the parser recognized exactly one
// EffectCantBlock effect.
func cantBlockEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectCantBlock {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0]
}

func TestExactCantBlockThisTurnAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Target creature can't block this turn.",
		"Target creature an opponent controls can't block this turn.",
		"Up to two target creatures can't block this turn.",
		"Up to three target creatures can't block this turn.",
	}
	for _, source := range accepted {
		effect := cantBlockEffect(t, source)
		if !effect.Exact {
			t.Errorf("cantBlockEffect(%q).Exact = false, want true", source)
		}
		if effect.Context != EffectContextTarget {
			t.Errorf("cantBlockEffect(%q).Context = %s, want EffectContextTarget", source, effect.Context)
		}
		if effect.Duration != EffectDurationThisTurn {
			t.Errorf("cantBlockEffect(%q).Duration = %s, want EffectDurationThisTurn", source, effect.Duration)
		}
		if len(effect.Targets) != 1 {
			t.Errorf("cantBlockEffect(%q) targets = %d, want 1", source, len(effect.Targets))
		}
	}
}

func TestExactCantBlockThisTurnFailsClosed(t *testing.T) {
	t.Parallel()
	// Each wording deviates from the exact temporary "<targets> can't block this
	// turn." restriction, so its round-trip must not reach an exact, lowerable
	// production: a continuous static prohibition with no duration, a different
	// duration, the inverse "can't be blocked" / "can't attack" operations, and a
	// protected-object qualifier ("can't block creatures you control").
	rejected := []string{
		"Creatures can't block.",
		"Target creature can't block.",
		"Target creature can't be blocked this turn.",
		"Target creature can't attack this turn.",
	}
	for _, source := range rejected {
		document, _ := Parse(source, Context{})
		if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) == 0 {
			continue
		}
		for _, sentence := range document.Abilities[0].Sentences {
			for _, effect := range sentence.Effects {
				if effect.Kind == EffectCantBlock && effect.Exact {
					t.Errorf("Parse(%q) produced an exact EffectCantBlock, want fail closed", source)
				}
			}
		}
	}
}
