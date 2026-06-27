package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// conditionalDamageReplacementSpell lowers a single-face conditional damage
// amount-replacement spell and returns its single spell mode, asserting it fused
// into one spell over one target with exactly two gated Damage instructions.
func conditionalDamageReplacementSpell(t *testing.T, card *ScryfallCard) game.Mode {
	t.Helper()
	face := lowerSingleFace(t, card)
	if !face.SpellAbility.Exists {
		t.Fatalf("%s produced no spell ability", card.Name)
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("%s modes = %d, want 1", card.Name, len(modes))
	}
	mode := modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("%s targets = %d, want 1", card.Name, len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("%s sequence = %d, want 2 (base + instead)", card.Name, len(mode.Sequence))
	}
	return mode
}

// assertGatedDamage asserts the instruction is a fixed-amount Damage to the sole
// target (slot 0) gated on a wrapped condition with the expected negation.
func assertGatedDamage(t *testing.T, label string, instruction game.Instruction, amount int, negated bool) {
	t.Helper()
	damage, ok := instruction.Primitive.(game.Damage)
	if !ok {
		t.Fatalf("%s primitive = %#v, want Damage", label, instruction.Primitive)
	}
	if damage.Recipient != game.AnyTargetDamageRecipient(0) {
		t.Fatalf("%s recipient = %#v, want any-target slot 0", label, damage.Recipient)
	}
	if got := damage.Amount; got != game.Fixed(amount) {
		t.Fatalf("%s amount = %#v, want fixed %d", label, got, amount)
	}
	if !instruction.Condition.Exists || !instruction.Condition.Val.Condition.Exists {
		t.Fatalf("%s must be gated on a wrapped condition", label)
	}
	if instruction.Condition.Val.Condition.Val.Negate != negated {
		t.Fatalf("%s gate Negate = %v, want %v", label, instruction.Condition.Val.Condition.Val.Negate, negated)
	}
}

// TestLowerShivanFireConditionalDamageReplacement verifies the kicked
// amount-replacement form ("deals 2 to target creature. If kicked, it deals 4
// instead.") fuses into one spell: the base 2 damage gated on not-kicked and the
// 4 damage gated on kicked, so exactly one resolves.
func TestLowerShivanFireConditionalDamageReplacement(t *testing.T) {
	t.Parallel()
	mode := conditionalDamageReplacementSpell(t, &ScryfallCard{
		Name:     "Shivan Fire",
		Layout:   "normal",
		TypeLine: "Instant",
		OracleText: "Kicker {4} (You may pay an additional {4} as you cast this spell.)\n" +
			"Shivan Fire deals 2 damage to target creature. If this spell was kicked, it deals 4 damage instead.",
	})
	if mode.Targets[0].Allow&game.TargetAllowPermanent == 0 {
		t.Fatalf("target Allow = %#v, want permanent", mode.Targets[0].Allow)
	}
	assertGatedDamage(t, "base", mode.Sequence[0], 2, true)
	assertGatedDamage(t, "instead", mode.Sequence[1], 4, false)
	baseGate := mode.Sequence[1].Condition.Val.Condition.Val
	if !baseGate.SpellWasKicked {
		t.Fatalf("instead gate = %+v, want SpellWasKicked", baseGate)
	}
}

// TestLowerBurstLightningAnyTargetReplacement verifies the "any target" variant
// lowers with an any-target permanent-or-player slot shared by both damage
// amounts.
func TestLowerBurstLightningAnyTargetReplacement(t *testing.T) {
	t.Parallel()
	mode := conditionalDamageReplacementSpell(t, &ScryfallCard{
		Name:     "Burst Lightning",
		Layout:   "normal",
		TypeLine: "Instant",
		OracleText: "Kicker {4} (You may pay an additional {4} as you cast this spell.)\n" +
			"Burst Lightning deals 2 damage to any target. If this spell was kicked, it deals 4 damage instead.",
	})
	if mode.Targets[0].Allow&(game.TargetAllowPermanent|game.TargetAllowPlayer) == 0 {
		t.Fatalf("target Allow = %#v, want any target", mode.Targets[0].Allow)
	}
	assertGatedDamage(t, "base", mode.Sequence[0], 2, true)
	assertGatedDamage(t, "instead", mode.Sequence[1], 4, false)
}

// TestLowerFrostBiteSnowReplacement verifies a non-kicked gate ("if you control
// three or more snow permanents") drives the amount replacement.
func TestLowerFrostBiteSnowReplacement(t *testing.T) {
	t.Parallel()
	mode := conditionalDamageReplacementSpell(t, &ScryfallCard{
		Name:       "Frost Bite",
		Layout:     "normal",
		TypeLine:   "Snow Instant",
		OracleText: "Frost Bite deals 2 damage to target creature or planeswalker. If you control three or more snow permanents, it deals 3 damage instead.",
	})
	assertGatedDamage(t, "base", mode.Sequence[0], 2, true)
	assertGatedDamage(t, "instead", mode.Sequence[1], 3, false)
}

// TestLowerFirebendingLessonRestatedTargetReplacement verifies the form that
// restates the recipient ("it deals 5 damage to that creature instead") fuses
// onto the same chosen target rather than introducing a new one.
func TestLowerFirebendingLessonRestatedTargetReplacement(t *testing.T) {
	t.Parallel()
	mode := conditionalDamageReplacementSpell(t, &ScryfallCard{
		Name:     "Firebending Lesson",
		Layout:   "normal",
		TypeLine: "Instant — Lesson",
		OracleText: "Kicker {4} (You may pay an additional {4} as you cast this spell.)\n" +
			"Firebending Lesson deals 2 damage to target creature. If this spell was kicked, it deals 5 damage to that creature instead.",
	})
	assertGatedDamage(t, "base", mode.Sequence[0], 2, true)
	assertGatedDamage(t, "instead", mode.Sequence[1], 5, false)
}

// TestLowerInvasiveManeuversTrailingConditionReplacement verifies the trailing
// "It deals 5 damage instead if <condition>." word order also fuses.
func TestLowerInvasiveManeuversTrailingConditionReplacement(t *testing.T) {
	t.Parallel()
	mode := conditionalDamageReplacementSpell(t, &ScryfallCard{
		Name:       "Invasive Maneuvers",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Invasive Maneuvers deals 3 damage to target creature. It deals 5 damage instead if you control a Spacecraft.",
	})
	assertGatedDamage(t, "base", mode.Sequence[0], 3, true)
	assertGatedDamage(t, "instead", mode.Sequence[1], 5, false)
}

// TestLowerConditionalDamageReplacementFailsClosed verifies shapes outside the
// single-target amount-replacement family stay rejected: a second distinct
// target ("another target"), an additive "also deals" rider, and group "each
// creature" damage.
func TestLowerConditionalDamageReplacementFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		card ScryfallCard
	}{
		{
			name: "another target",
			card: ScryfallCard{
				Name:     "Magma Burst",
				Layout:   "normal",
				TypeLine: "Instant",
				OracleText: "Kicker—Sacrifice two lands. (You may sacrifice two lands in addition to any other costs as you cast this spell.)\n" +
					"Magma Burst deals 3 damage to any target. If this spell was kicked, it deals 3 damage to another target.",
			},
		},
		{
			name: "also deals second target",
			card: ScryfallCard{
				Name:     "Goblin Barrage",
				Layout:   "normal",
				TypeLine: "Sorcery",
				OracleText: "Kicker—Sacrifice an artifact or Goblin. (You may sacrifice an artifact or Goblin in addition to any other costs as you cast this spell.)\n" +
					"Goblin Barrage deals 4 damage to target creature. If this spell was kicked, it also deals 4 damage to target player or planeswalker.",
			},
		},
		{
			name: "each creature group damage",
			card: ScryfallCard{
				Name:     "Cinderclasm",
				Layout:   "normal",
				TypeLine: "Instant",
				OracleText: "Kicker {R} (You may pay an additional {R} as you cast this spell.)\n" +
					"Cinderclasm deals 1 damage to each creature. If it was kicked, it deals 2 damage to each creature instead.",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			lowerSingleFaceExpectingUnsupported(t, &tc.card)
		})
	}
}
