package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerTargetPlayerVariableRandomDiscard proves the variable-count targeted
// random discard family ("Target player discards X cards at random.", Mind
// Twist, Mind Shatter) lowers to a single Discard whose recipient is the target
// player, whose amount is the spell's {X}, and which discards at random.
func TestLowerTargetPlayerVariableRandomDiscard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mind Twist",
		Layout:     "normal",
		ManaCost:   "{X}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Target player discards X cards at random.",
		Colors:     []string{"B"},
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not found")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowPlayer {
		t.Fatalf("targets = %#v, want one player target", mode.Targets)
	}
	prim, ok := mode.Sequence[0].Primitive.(game.Discard)
	if !ok {
		t.Fatalf("primitive = %T, want game.Discard", mode.Sequence[0].Primitive)
	}
	if !prim.AtRandom {
		t.Error("AtRandom = false, want true")
	}
	if prim.Player.Kind() != game.PlayerReferenceTargetPlayer || prim.Player.TargetIndex() != 0 {
		t.Fatalf("player = %#v, want TargetPlayerReference(0)", prim.Player)
	}
	dyn := prim.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountX {
		t.Fatalf("amount = %#v, want DynamicAmountX", prim.Amount)
	}
}

// TestLowerEachPlayerVariableDiscard proves the group variable-count discard
// ("Each player discards X cards.") lowers to a Discard over the all-players
// group with the spell's {X} count.
func TestLowerEachPlayerVariableDiscard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Each Discard",
		Layout:     "normal",
		ManaCost:   "{X}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Each player discards X cards.",
		Colors:     []string{"B"},
	})
	mode := face.SpellAbility.Val.Modes[0]
	prim, ok := mode.Sequence[0].Primitive.(game.Discard)
	if !ok {
		t.Fatalf("primitive = %T, want game.Discard", mode.Sequence[0].Primitive)
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Fatalf("group = %v, want all players", prim.PlayerGroup.Kind)
	}
	dyn := prim.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountX {
		t.Fatalf("amount = %#v, want DynamicAmountX", prim.Amount)
	}
}
