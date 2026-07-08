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

// TestParseDieThisTurnExileReplacementDamagedCreatureSubject verifies the rider
// recognizer accepts the "a creature dealt damage this way" burn subject
// (Yamabushi's Flame, Demonfire) and marks it with ExileDieSubjectDamagedCreature.
// The parser owns the wording; the single-target binding feasibility (rejecting
// the mass "each creature" form) is enforced by the lowering, covered by
// TestLowerDieToExileRequiresSingleTarget.
func TestParseDieThisTurnExileReplacementDamagedCreatureSubject(t *testing.T) {
	t.Parallel()
	source := "Yamabushi's Flame deals 3 damage to any target. " +
		"If a creature dealt damage this way would die this turn, exile it instead."
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
	if effect.Kind != EffectExileIfWouldDieThisTurn || !effect.Exact {
		t.Fatalf("effect = %v exact=%v, want exact EffectExileIfWouldDieThisTurn", effect.Kind, effect.Exact)
	}
	if !effect.ExileDieSubjectDamagedCreature {
		t.Fatal("damaged-creature subject must set ExileDieSubjectDamagedCreature")
	}
}
