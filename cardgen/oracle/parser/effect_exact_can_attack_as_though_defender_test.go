package parser

import "testing"

// canAttackAsThoughDefenderEffect parses a single "<subject> can attack this
// turn as though it didn't have defender." sentence and returns its resolving
// effect, asserting that the parser recognized exactly one
// EffectCanAttackAsThoughDefender effect.
func canAttackAsThoughDefenderEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectCanAttackAsThoughDefender {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0]
}

// TestExactCanAttackAsThoughDefenderAccepts covers the self subject form
// "This creature can attack this turn as though it didn't have defender."
// (EffectContextSource), an activated-ability self grant (Glade Watcher).
func TestExactCanAttackAsThoughDefenderAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"{G}: This creature can attack this turn as though it didn't have defender.",
		"Discard a card: This creature can attack this turn as though it didn't have defender.",
	}
	for _, source := range accepted {
		effect := canAttackAsThoughDefenderEffect(t, source)
		if !effect.Exact {
			t.Errorf("canAttackAsThoughDefenderEffect(%q).Exact = false, want true", source)
		}
		if effect.Context != EffectContextSource {
			t.Errorf("canAttackAsThoughDefenderEffect(%q).Context = %s, want EffectContextSource", source, effect.Context)
		}
		if effect.Duration != EffectDurationThisTurn {
			t.Errorf("canAttackAsThoughDefenderEffect(%q).Duration = %s, want EffectDurationThisTurn", source, effect.Duration)
		}
	}
}

// TestExactCanAttackAsThoughDefenderFailsClosed ensures wordings that deviate
// from the temporary "<subject> can attack this turn as though it didn't have
// defender." restriction do not reach an exact, lowerable production: a
// different duration, an inverse "can't attack" operation, and a permission that
// is not about defender.
func TestExactCanAttackAsThoughDefenderFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"{G}: This creature can attack as though it didn't have defender.",
		"{G}: This creature can't attack this turn.",
		"{G}: This creature can attack this turn as though it weren't tapped.",
	}
	for _, source := range rejected {
		document, _ := Parse(source, Context{})
		if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) == 0 {
			continue
		}
		for _, sentence := range document.Abilities[0].Sentences {
			for _, effect := range sentence.Effects {
				if effect.Kind == EffectCanAttackAsThoughDefender && effect.Exact {
					t.Errorf("Parse(%q) produced an exact EffectCanAttackAsThoughDefender, want fail closed", source)
				}
			}
		}
	}
}
