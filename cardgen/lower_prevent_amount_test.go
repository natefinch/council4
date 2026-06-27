package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerPreventAmountAnyTarget covers the amount-based "Prevent the next N
// damage that would be dealt to any target this turn." shield (Master
// Apothecary), which lowers to a fixed-amount PreventDamage whose recipient is
// resolved through a single any-target slot constrained to a player or
// permanent.
func TestLowerPreventAmountAnyTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Master Apothecary",
		Layout:     "normal",
		TypeLine:   "Creature — Human Cleric",
		OracleText: "Tap an untapped Cleric you control: Prevent the next 2 damage that would be dealt to any target this turn.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want one", face.ActivatedAbilities)
	}
	modes := face.ActivatedAbilities[0].Content.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("ability content = %#v, want one mode with one instruction", modes)
	}
	if len(modes[0].Targets) != 1 {
		t.Fatalf("mode targets = %#v, want one any-target slot", modes[0].Targets)
	}
	if allow := modes[0].Targets[0].Allow; allow != game.TargetAllowPermanent|game.TargetAllowPlayer {
		t.Fatalf("target allow = %v, want permanent|player", allow)
	}
	prevent, ok := modes[0].Sequence[0].Primitive.(game.PreventDamage)
	if !ok {
		t.Fatalf("primitive = %#v, want game.PreventDamage", modes[0].Sequence[0].Primitive)
	}
	if prevent.Global || prevent.All || prevent.CombatOnly || prevent.BySource {
		t.Fatalf("prevent = %#v, want a fixed-amount targeted shield", prevent)
	}
	if got, want := prevent.Amount, game.Fixed(2); got != want {
		t.Fatalf("prevent amount = %#v, want %#v", got, want)
	}
	object, ok := prevent.AnyTarget.AnyTargetObjectReference()
	if !ok || object.TargetIndex() != 0 {
		t.Fatalf("any-target object reference = %#v, %v, want target index 0", object, ok)
	}
	player, ok := prevent.AnyTarget.AnyTargetPlayerReference()
	if !ok || player.TargetIndex() != 0 {
		t.Fatalf("any-target player reference = %#v, %v, want target index 0", player, ok)
	}
}

// TestLowerPreventAmountYou covers the "dealt to you" recipient form, which
// shields the controller directly with no target slot.
func TestLowerPreventAmountYou(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Personal Ward",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Prevent the next 3 damage that would be dealt to you this turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatalf("spell ability = %#v, want present", face.SpellAbility)
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 || len(modes[0].Targets) != 0 {
		t.Fatalf("ability content = %#v, want one targetless mode with one instruction", modes)
	}
	prevent, ok := modes[0].Sequence[0].Primitive.(game.PreventDamage)
	if !ok {
		t.Fatalf("primitive = %#v, want game.PreventDamage", modes[0].Sequence[0].Primitive)
	}
	if prevent.Player.Kind() == game.PlayerReferenceNone {
		t.Fatalf("prevent player = %#v, want controller reference", prevent.Player)
	}
	if got, want := prevent.Amount, game.Fixed(3); got != want {
		t.Fatalf("prevent amount = %#v, want %#v", got, want)
	}
}
