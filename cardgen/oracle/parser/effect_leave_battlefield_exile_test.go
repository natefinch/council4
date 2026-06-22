package parser

import "testing"

// TestParseLeaveBattlefieldExileReplacement verifies that the
// leaves-the-battlefield exile replacement recognizer emits a single exact
// EffectExileIfLeaveBattlefield effect for the "it" back-reference form (Whip of
// Erebos), the "this <type>" self-applied form, and the shorter
// "...exile it instead." wording, and that no spurious condition surfaces.
func TestParseLeaveBattlefieldExileReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source  string
		context EffectContextKind
	}{
		{
			"If it would leave the battlefield, exile it instead of putting it anywhere else.",
			EffectContextReferencedObject,
		},
		{
			"If it would leave the battlefield, exile it instead.",
			EffectContextReferencedObject,
		},
		{
			"If this creature would leave the battlefield, exile it instead of putting it anywhere else.",
			EffectContextSource,
		},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v, want none", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			effect := effects[0]
			if effect.Kind != EffectExileIfLeaveBattlefield {
				t.Fatalf("effect Kind = %v, want EffectExileIfLeaveBattlefield", effect.Kind)
			}
			if !effect.Exact {
				t.Fatal("effect Exact = false, want true")
			}
			if effect.Context != test.context {
				t.Fatalf("effect Context = %v, want %v", effect.Context, test.context)
			}
		})
	}
}
