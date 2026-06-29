package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// excessDamageToControllerInstruction lowers a single-face card and returns its
// spell ability's sole Damage instruction, asserting the card lowered without
// diagnostics into a one-instruction redirect sequence.
func excessDamageToControllerInstruction(t *testing.T, card *ScryfallCard) game.Damage {
	t.Helper()
	face := lowerSingleFace(t, card)
	if !face.SpellAbility.Exists {
		t.Fatal("card has no spell ability")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(modes))
	}
	sequence := modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	damage, ok := sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("sequence[0] primitive = %T, want game.Damage", sequence[0].Primitive)
	}
	if !damage.Recipient.Valid() {
		t.Fatal("damage has no valid recipient")
	}
	return damage
}

// TestLowerExcessDamageToControllerFixed proves Flame Spill's "deals 4 damage to
// target creature. Excess damage is dealt to that creature's controller instead."
// lowers to a single fixed-damage instruction whose excess is redirected to the
// target's controller.
func TestLowerExcessDamageToControllerFixed(t *testing.T) {
	t.Parallel()
	damage := excessDamageToControllerInstruction(t, &ScryfallCard{
		Name:       "Flame Spill",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{R}",
		OracleText: "Flame Spill deals 4 damage to target creature. Excess damage is dealt to that creature's controller instead.",
	})
	if damage.Amount.IsDynamic() || damage.Amount.Value() != 4 {
		t.Fatalf("damage amount = %#v, want fixed 4", damage.Amount)
	}
	if _, ok := damage.Recipient.AnyTargetObjectReference(); !ok {
		t.Fatalf("recipient = %#v, want any-target permanent", damage.Recipient)
	}
	want := game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0)))
	if damage.ExcessRecipient != want {
		t.Fatalf("excess recipient = %#v, want %#v", damage.ExcessRecipient, want)
	}
}

// TestLowerExcessDamageToControllerDynamicAmount proves Gandalf's Sanction's
// graveyard-count X damage lowers to the same count-cards-in-zone amount as the
// plain single-target spell, with the excess redirected to the controller.
func TestLowerExcessDamageToControllerDynamicAmount(t *testing.T) {
	t.Parallel()
	damage := excessDamageToControllerInstruction(t, &ScryfallCard{
		Name:     "Gandalf's Sanction",
		Layout:   "normal",
		TypeLine: "Sorcery",
		ManaCost: "{1}{U}{R}",
		OracleText: "Gandalf's Sanction deals X damage to target creature, where X is the number " +
			"of instant and sorcery cards in your graveyard. Excess damage is dealt to that creature's controller instead.",
	})
	dyn := damage.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountCountCardsInZone {
		t.Fatalf("damage amount = %#v, want DynamicAmountCountCardsInZone", damage.Amount)
	}
	want := game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0)))
	if damage.ExcessRecipient != want {
		t.Fatalf("excess recipient = %#v, want %#v", damage.ExcessRecipient, want)
	}
}

// TestLowerExcessDamageToControllerOwner proves the "owner instead" wording
// redirects the excess to the target's owner rather than its controller.
func TestLowerExcessDamageToControllerOwner(t *testing.T) {
	t.Parallel()
	damage := excessDamageToControllerInstruction(t, &ScryfallCard{
		Name:       "Owner Spill",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{R}",
		OracleText: "Owner Spill deals 4 damage to target creature. Excess damage is dealt to that creature's owner instead.",
	})
	want := game.PlayerDamageRecipient(game.ObjectOwnerReference(game.TargetPermanentReference(0)))
	if damage.ExcessRecipient != want {
		t.Fatalf("excess recipient = %#v, want %#v", damage.ExcessRecipient, want)
	}
}
