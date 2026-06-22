package parser

import (
	"testing"
)

// TestParseRollDieEffect proves the parser recognizes "roll a d<N>" and records
// the die size, and that a following "...equal to the result." amount types the
// die-roll-result dynamic amount.
func TestParseRollDieEffect(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		sides  int
	}{
		{"Roll a d20.", 20},
		{"Roll a d6.", 6},
		{"Roll a d100.", 100},
	}
	for _, tc := range cases {
		document, diagnostics := Parse(tc.source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", tc.source, diagnostics)
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 1 {
			t.Fatalf("Parse(%q) effects = %#v, want one", tc.source, effects)
		}
		effect := effects[0]
		if effect.Kind != EffectRollDie {
			t.Errorf("Parse(%q) kind = %v, want EffectRollDie", tc.source, effect.Kind)
		}
		if effect.DieSides != tc.sides {
			t.Errorf("Parse(%q) DieSides = %d, want %d", tc.source, effect.DieSides, tc.sides)
		}
		if !effect.Exact {
			t.Errorf("Parse(%q) Exact = false, want true", tc.source)
		}
	}
}

// TestParseRollDieResultSequence proves the two-sentence dice payoff parses to a
// die roll followed by a token create whose count reads the die result.
func TestParseRollDieResultSequence(t *testing.T) {
	t.Parallel()
	source := "Roll a d20. You create a number of Treasure tokens equal to the result."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	var effects []EffectSyntax
	for _, sentence := range document.Abilities[0].Sentences {
		effects = append(effects, sentence.Effects...)
	}
	if len(effects) != 2 {
		t.Fatalf("Parse(%q) effects = %#v, want two", source, effects)
	}
	if effects[0].Kind != EffectRollDie || effects[0].DieSides != 20 {
		t.Errorf("Parse(%q) effect[0] = %v/d%d, want EffectRollDie/d20", source, effects[0].Kind, effects[0].DieSides)
	}
	create := effects[1]
	if create.Kind != EffectCreate {
		t.Errorf("Parse(%q) effect[1] kind = %v, want EffectCreate", source, create.Kind)
	}
	if create.Amount.DynamicKind != EffectDynamicAmountDieRollResult {
		t.Errorf("Parse(%q) amount kind = %v, want EffectDynamicAmountDieRollResult", source, create.Amount.DynamicKind)
	}
	if !create.Exact {
		t.Errorf("Parse(%q) create Exact = false, want true", source)
	}
}

// TestParseRollDieEffectFailsClosed proves non-die wordings do not match the
// die-roll recognizer.
func TestParseRollDieEffectFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Roll a die.",
		"Roll two d20.",
		"Roll a d20 and a d6.",
	} {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			continue
		}
		for _, ability := range document.Abilities {
			for _, sentence := range ability.Sentences {
				for _, effect := range sentence.Effects {
					if effect.Kind == EffectRollDie {
						t.Errorf("Parse(%q) matched EffectRollDie, want fail-closed", source)
					}
				}
			}
		}
	}
}
