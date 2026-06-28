package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerEnchantedControllerLosesLife proves the standalone "its controller
// loses N life" triggered effect lowers to a fixed LoseLife on the controller of
// the triggering permanent (Corrupted Roots, Sinister Possession,
// Contaminated Ground). The drained player resolves through the same
// event-controller reference the life-drain path uses, but here the life loss
// stands alone without an accompanying "you gain" rider.
func TestLowerEnchantedControllerLosesLife(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sinister Possession",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		ManaCost:   "{1}{B}",
		OracleText: "Enchant creature\nWhenever enchanted creature attacks or blocks, its controller loses 2 life.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	prim := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive
	lose, ok := prim.(game.LoseLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.LoseLife", prim)
	}
	if lose.Amount.IsDynamic() || lose.Amount.Value() != 2 {
		t.Fatalf("amount = %#v, want fixed 2", lose.Amount)
	}
	if lose.Player != game.ObjectControllerReference(game.EventPermanentReference()) {
		t.Fatalf("player = %#v, want event-permanent controller", lose.Player)
	}
}

// TestLowerAttackingCreatureControllerLosesLife proves the "that creature's
// controller loses N life" combat form lowers to the same event-controller
// LoseLife (Blood Reckoning, Carnage Gladiator). The drained player is the
// controller of the creature that fired the attack/block trigger.
func TestLowerAttackingCreatureControllerLosesLife(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Blood Reckoning",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{2}{R}",
		OracleText: "Whenever a creature attacks you or a planeswalker you control, that creature's controller loses 1 life.",
	})
	prim := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive
	lose, ok := prim.(game.LoseLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.LoseLife", prim)
	}
	if lose.Amount.IsDynamic() || lose.Amount.Value() != 1 {
		t.Fatalf("amount = %#v, want fixed 1", lose.Amount)
	}
	if lose.Player != game.ObjectControllerReference(game.EventPermanentReference()) {
		t.Fatalf("player = %#v, want event-permanent controller", lose.Player)
	}
}
