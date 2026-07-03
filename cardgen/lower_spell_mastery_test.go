package cardgen

import (
	"testing"
)

// TestLowerSpellMasteryInstantOrSorceryGraveyardGate proves the "Spell mastery"
// ability word plus the "two or more instant and/or sorcery cards in your
// graveyard" count condition gate an escalation. Fiery Impulse lowers to two
// gated Damage instructions: 2 damage unless the graveyard holds 2+ instant/
// sorcery cards, 3 damage when it does.
func TestLowerSpellMasteryInstantOrSorceryGraveyardGate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Fiery Impulse",
		Layout:   "normal",
		TypeLine: "Instant",
		OracleText: "This spell deals 2 damage to target creature.\n" +
			"Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, " +
			"it deals 3 damage instead.",
	})
	if len(face.SpellAbility.Val.Modes) != 1 {
		t.Fatalf("modes = %#v", face.SpellAbility.Val.Modes)
	}
	seq := face.SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence = %#v, want two gated damage instructions", seq)
	}
	var sawBaseGate, sawEscalationGate bool
	for _, instr := range seq {
		if !instr.Condition.Exists || !instr.Condition.Val.Condition.Exists {
			t.Fatalf("instruction not gated: %#v", instr)
		}
		cond := instr.Condition.Val.Condition.Val
		if cond.ControllerGraveyardInstantOrSorceryCountAtLeast != 2 {
			t.Fatalf("gate = %#v, want InstantOrSorcery count 2", cond)
		}
		if cond.Negate {
			sawBaseGate = true
		} else {
			sawEscalationGate = true
		}
	}
	if !sawBaseGate || !sawEscalationGate {
		t.Fatalf("want both a negated base gate and a plain escalation gate: %#v", seq)
	}
}
