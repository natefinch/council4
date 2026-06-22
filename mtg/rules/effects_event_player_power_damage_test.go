package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestEventPlayerSourcePowerDamageHitsTriggeringPlayer proves the Gleeful
// Arsonist shape: "Whenever an opponent casts a noncreature spell, this creature
// deals damage equal to its power to that player." The amount reads the source
// creature's power, the damage source is the source permanent, and the recipient
// is the triggering event's player (the opponent who cast the spell). That
// player loses life equal to the source's power; the ability's own controller is
// untouched.
func TestEventPlayerSourcePowerDamageHitsTriggeringPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	startP1 := g.Players[game.Player1].Life
	startP2 := g.Players[game.Player2].Life

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		SourceID:        source.ObjectID,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:       game.EventSpellCast,
			Controller: game.Player2,
		},
	}

	resolveInstruction(engine, g, obj, game.Damage{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountObjectPower,
			Multiplier: 1,
			Object:     game.SourcePermanentReference(),
		}),
		Recipient:    game.PlayerDamageRecipient(game.EventPlayerReference()),
		DamageSource: opt.Val(game.SourcePermanentReference()),
	}, &TurnLog{})

	if got := startP2 - g.Players[game.Player2].Life; got != 5 {
		t.Fatalf("triggering player life lost = %d, want 5 (source power)", got)
	}
	if got := g.Players[game.Player1].Life; got != startP1 {
		t.Fatalf("controller life = %d, want %d (only the triggering player is dealt damage)", got, startP1)
	}
}
