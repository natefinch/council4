package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func partyCreature(name string, subtypes ...types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Subtypes:  subtypes,
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

func TestControllerPartySizeAssignsEachCreatureOneRole(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, partyCreature("Cleric Wizard", types.Cleric, types.Wizard))
	addCombatPermanent(g, game.Player1, partyCreature("Cleric", types.Cleric))
	addCombatPermanent(g, game.Player1, partyCreature("Rogue Warrior", types.Rogue, types.Warrior))

	if got := controllerPartySize(g, game.Player1); got != 3 {
		t.Fatalf("party size = %d, want 3", got)
	}
	warrior := addCombatPermanent(g, game.Player1, partyCreature("Warrior", types.Warrior))
	if got := controllerPartySize(g, game.Player1); got != 4 {
		t.Fatalf("party size after Warrior = %d, want 4", got)
	}
	warrior.PhasedOut = true
	if got := controllerPartySize(g, game.Player1); got != 3 {
		t.Fatalf("party size with phased-out Warrior = %d, want 3", got)
	}
	warrior.PhasedOut = false
	addCombatPermanent(g, game.Player2, partyCreature("Opponent Wizard", types.Wizard))
	if got := controllerPartySize(g, game.Player1); got != 4 {
		t.Fatalf("opponent changed party size to %d, want 4", got)
	}
}
