package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
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

// TestLowerRamThroughTrampleExcessRedirect proves Ram Through's conditional bite
// with trample overflow lowers to two mutually exclusive Damage instructions
// gated on the dealer having trample: a redirecting branch whose excess flows to
// the bitten creature's controller when the dealer has trample, and a plain
// branch otherwise. Both bites deal the dealer's power (target 0) to the bitten
// creature (target 1).
func TestLowerRamThroughTrampleExcessRedirect(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Ram Through",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{1}{G}",
		OracleText: "Target creature you control deals damage equal to its power to target creature you don't control. " +
			"If the creature you control has trample, excess damage is dealt to that creature's controller instead.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("card has no spell ability")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(modes))
	}
	sequence := modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(sequence))
	}

	trampleSelection := opt.Val(game.Selection{Keyword: game.Trample})
	assertDealerPower := func(label string, amount game.Quantity) {
		t.Helper()
		dyn := amount.DynamicAmount()
		if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountObjectPower ||
			dyn.Val.Object != game.TargetPermanentReference(0) {
			t.Fatalf("%s amount = %#v, want dealer (target 0) power", label, amount)
		}
	}

	redirect, ok := sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("sequence[0] primitive = %T, want game.Damage", sequence[0].Primitive)
	}
	assertDealerPower("redirect", redirect.Amount)
	if got, ok := redirect.Recipient.AnyTargetObjectReference(); !ok || got != game.TargetPermanentReference(1) {
		t.Fatalf("redirect recipient = %#v, want any-target 1", redirect.Recipient)
	}
	if redirect.DamageSource != opt.Val(game.TargetPermanentReference(0)) {
		t.Fatalf("redirect source = %#v, want target 0", redirect.DamageSource)
	}
	wantExcess := game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(1)))
	if redirect.ExcessRecipient != wantExcess {
		t.Fatalf("redirect excess = %#v, want controller of target 1", redirect.ExcessRecipient)
	}
	wantTrampleGate := opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
		Object:        opt.Val(game.TargetPermanentReference(0)),
		ObjectMatches: trampleSelection,
	})})
	if !reflect.DeepEqual(sequence[0].Condition, wantTrampleGate) {
		t.Fatalf("redirect gate = %#v, want dealer-has-trample", sequence[0].Condition)
	}

	plain, ok := sequence[1].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("sequence[1] primitive = %T, want game.Damage", sequence[1].Primitive)
	}
	assertDealerPower("plain", plain.Amount)
	if got, ok := plain.Recipient.AnyTargetObjectReference(); !ok || got != game.TargetPermanentReference(1) {
		t.Fatalf("plain recipient = %#v, want any-target 1", plain.Recipient)
	}
	if plain.DamageSource != opt.Val(game.TargetPermanentReference(0)) {
		t.Fatalf("plain source = %#v, want target 0", plain.DamageSource)
	}
	if plain.ExcessRecipient != (game.DamageRecipient{}) {
		t.Fatalf("plain excess = %#v, want none", plain.ExcessRecipient)
	}
	wantNoTrampleGate := opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
		Negate:        true,
		Object:        opt.Val(game.TargetPermanentReference(0)),
		ObjectMatches: trampleSelection,
	})})
	if !reflect.DeepEqual(sequence[1].Condition, wantNoTrampleGate) {
		t.Fatalf("plain gate = %#v, want dealer-lacks-trample", sequence[1].Condition)
	}
}
