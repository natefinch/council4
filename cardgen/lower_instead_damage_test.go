package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerBrimstoneVolleyMorbidInsteadDamage verifies that Brimstone Volley's
// base "3 damage to any target" paragraph and its "Morbid — ... 5 damage instead
// if a creature died this turn." paragraph fuse into a single spell over one
// any-target whose 3 damage resolves only when no creature died this turn and
// whose 5 damage resolves only when one did.
func TestLowerBrimstoneVolleyMorbidInsteadDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Brimstone Volley",
		Layout:   "normal",
		TypeLine: "Instant",
		OracleText: "Brimstone Volley deals 3 damage to any target.\n" +
			"Morbid — Brimstone Volley deals 5 damage instead if a creature died this turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Brimstone Volley produced no spell ability")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("modes = %#v, want 1", modes)
	}
	seq := modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2 (base + morbid)", len(seq))
	}
	base, ok := seq[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("base instruction = %#v, want Damage", seq[0].Primitive)
	}
	morbid, ok := seq[1].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("morbid instruction = %#v, want Damage", seq[1].Primitive)
	}
	if base.Recipient != game.AnyTargetDamageRecipient(0) ||
		morbid.Recipient != game.AnyTargetDamageRecipient(0) {
		t.Fatalf("recipients = %#v / %#v, want any-target 0", base.Recipient, morbid.Recipient)
	}
	if got := base.Amount; got != game.Fixed(3) {
		t.Fatalf("base amount = %#v, want 3", got)
	}
	if got := morbid.Amount; got != game.Fixed(5) {
		t.Fatalf("morbid amount = %#v, want 5", got)
	}
	for i, instr := range seq {
		if !instr.Condition.Exists || !instr.Condition.Val.Condition.Exists {
			t.Fatalf("instruction[%d] is ungated: %#v", i, instr)
		}
		if !instr.Condition.Val.Condition.Val.EventHistory.Exists {
			t.Fatalf("instruction[%d] is not event-history gated: %#v", i, instr.Condition.Val)
		}
	}
	if !seq[0].Condition.Val.Condition.Val.Negate {
		t.Fatal("base damage must be gated on NOT(a creature died this turn)")
	}
	if seq[1].Condition.Val.Condition.Val.Negate {
		t.Fatal("morbid damage must be gated on (a creature died this turn)")
	}
}

// TestLowerThermalBlastThresholdInsteadDamage verifies the threshold-condition
// variant fuses the same way: a base 3 damage to target creature and a 5 damage
// "instead" alternative gated on seven or more graveyard cards.
func TestLowerThermalBlastThresholdInsteadDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Thermal Blast",
		Layout:   "normal",
		TypeLine: "Instant",
		OracleText: "Thermal Blast deals 3 damage to target creature.\n" +
			"Threshold — Thermal Blast deals 5 damage instead if there are seven or more cards in your graveyard.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Thermal Blast produced no spell ability")
	}
	seq := face.SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2 (base + threshold)", len(seq))
	}
	if got := seq[0].Condition.Val.Condition.Val.Aggregates; len(got) != 1 || got[0].Aggregate != game.AggregateControllerGraveyardCardCount || got[0].Value != 7 {
		t.Fatalf("base gate graveyard aggregate = %+v, want graveyard-card-count >= 7", seq[0].Condition.Val.Condition.Val.Aggregates)
	}
	if !seq[0].Condition.Val.Condition.Val.Negate {
		t.Fatal("base damage must be gated on NOT(threshold)")
	}
	if seq[1].Condition.Val.Condition.Val.Negate {
		t.Fatal("threshold damage must be gated on threshold")
	}
}
