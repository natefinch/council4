package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestEventPlayerDamageHitsTriggeringPlayer proves the Underworld Dreams /
// Megrim shape: "Whenever an opponent draws a card, ~ deals N damage to that
// player." The damage recipient is the player named by the triggering event
// (the opponent who drew), resolved through EventPlayerReference, so only that
// player loses life while the source's controller is untouched.
func TestEventPlayerDamageHitsTriggeringPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	startP1 := g.Players[game.Player1].Life
	startP2 := g.Players[game.Player2].Life

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		SourceID:        source.ObjectID,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:   game.EventCardDrawn,
			Player: game.Player2,
		},
	}

	resolveInstruction(engine, g, obj, game.Damage{
		Amount:       game.Fixed(2),
		Recipient:    game.PlayerDamageRecipient(game.EventPlayerReference()),
		DamageSource: opt.Val(game.SourcePermanentReference()),
	}, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != startP2-2 {
		t.Fatalf("triggering player life = %d, want %d (2 damage)", got, startP2-2)
	}
	if got := g.Players[game.Player1].Life; got != startP1 {
		t.Fatalf("source controller life = %d, want %d (untouched)", got, startP1)
	}
}
