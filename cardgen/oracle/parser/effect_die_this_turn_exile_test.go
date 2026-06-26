package parser

import "testing"

// TestParseDieThisTurnExileReplacement verifies that the would-die exile rider
// "If <subject> would die this turn, exile it instead." emits a single exact
// EffectExileIfWouldDieThisTurn effect for the "that creature", "that creature
// or planeswalker", and "it" subject forms, and that the leading would-die
// clause does not surface as an unrecognized sibling effect or diagnostic.
func TestParseDieThisTurnExileReplacement(t *testing.T) {
	t.Parallel()
	tests := []string{
		"Lava Coil deals 4 damage to target creature. If that creature would die this turn, exile it instead.",
		"Obliterating Bolt deals 4 damage to target creature or planeswalker. If that creature or planeswalker would die this turn, exile it instead.",
		"Magma Spray deals 2 damage to target creature. If it would die this turn, exile it instead.",
	}
	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v, want none", diagnostics)
			}
			sentences := document.Abilities[0].Sentences
			rider := sentences[len(sentences)-1]
			if len(rider.Effects) != 1 {
				t.Fatalf("rider effects = %#v, want one", rider.Effects)
			}
			effect := rider.Effects[0]
			if effect.Kind != EffectExileIfWouldDieThisTurn {
				t.Fatalf("effect Kind = %v, want EffectExileIfWouldDieThisTurn", effect.Kind)
			}
			if !effect.Exact {
				t.Fatal("effect Exact = false, want true")
			}
		})
	}
}

// TestParseDieThisTurnExileReplacementRejectsGroupSubject verifies the rider
// recognizer fails closed for the group "a creature dealt damage this way"
// subject, which the single-target replacement cannot bind, so it does not
// masquerade as the single-target form.
func TestParseDieThisTurnExileReplacementRejectsGroupSubject(t *testing.T) {
	t.Parallel()
	source := "Anger of the Gods deals 3 damage to each creature. " +
		"If a creature dealt damage this way would die this turn, exile it instead."
	document, _ := Parse(source, Context{InstantOrSorcery: true})
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if effect.Kind == EffectExileIfWouldDieThisTurn {
					t.Fatal("group subject must not match the single-target die-this-turn rider")
				}
			}
		}
	}
}
