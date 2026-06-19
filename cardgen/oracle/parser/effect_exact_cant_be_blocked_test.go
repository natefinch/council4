package parser

import "testing"

// cantBeBlockedEffect parses a single can't-be-blocked-this-turn sentence and
// returns its resolving effect, asserting that the parser recognized exactly one
// EffectCantBeBlocked effect.
func cantBeBlockedEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectCantBeBlocked {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0]
}

func TestExactCantBeBlockedThisTurnAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Target creature can't be blocked this turn.",
		"Target creature you control can't be blocked this turn.",
		"Target creature an opponent controls can't be blocked this turn.",
		"Target attacking creature can't be blocked this turn.",
	}
	for _, source := range accepted {
		effect := cantBeBlockedEffect(t, source)
		if !effect.Exact {
			t.Errorf("cantBeBlockedEffect(%q).Exact = false, want true", source)
		}
		if effect.Context != EffectContextTarget {
			t.Errorf("cantBeBlockedEffect(%q).Context = %s, want EffectContextTarget", source, effect.Context)
		}
		if effect.Duration != EffectDurationThisTurn {
			t.Errorf("cantBeBlockedEffect(%q).Duration = %s, want EffectDurationThisTurn", source, effect.Duration)
		}
		if len(effect.Targets) != 1 {
			t.Errorf("cantBeBlockedEffect(%q) targets = %d, want 1", source, len(effect.Targets))
		}
	}
}

func TestExactCantBeBlockedThisTurnFailsClosed(t *testing.T) {
	t.Parallel()
	// Each wording deviates from the exact temporary "Target creature can't be
	// blocked this turn." restriction, so its round-trip must not reach an exact,
	// lowerable production: a different duration, an "except by ..." qualifier, a
	// group recipient, the inverse "can't block" / "can't attack" operations, and
	// any extra trailing qualifier. The conditional "... if it's tapped." form
	// is split into a separate condition clause by the parser, so its fail-closed
	// rejection is covered at the lowering layer rather than here.
	rejected := []string{
		"Target creature can't be blocked.",
		"Target creature can't be blocked until end of turn.",
		"Target creature can't be blocked this turn except by Walls.",
		"Target creature can't be blocked by more than one creature this turn.",
		"Up to two target creatures can't be blocked this turn.",
		"Target creature can't block this turn.",
		"Target creature can't attack this turn.",
	}
	for _, source := range rejected {
		document, _ := Parse(source, Context{})
		if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) == 0 {
			continue
		}
		for _, sentence := range document.Abilities[0].Sentences {
			for _, effect := range sentence.Effects {
				if effect.Kind == EffectCantBeBlocked && effect.Exact {
					t.Errorf("Parse(%q) produced an exact EffectCantBeBlocked, want fail closed", source)
				}
			}
		}
	}
}
