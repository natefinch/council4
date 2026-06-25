package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestObjectControllerGetsPoisonCounter covers the referenced-object-controller
// gain-player-counter recipient ("its controller gets a poison counter"): the
// AddPlayerCounter resolves ObjectControllerReference(EventPermanentReference())
// to the controller of the triggering event permanent and adds the counter to
// that player, not to the ability's controller. Relic Putrescence enchants an
// opponent's artifact, so its controller — the opponent — is poisoned.
func TestObjectControllerGetsPoisonCounter(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The triggering permanent is controlled by Player2 (the opponent); the
	// resolving ability is controlled by Player1.
	triggering := addCreaturePermanent(g, game.Player2)
	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventPermanentEnteredBattlefield,
			PermanentID: triggering.ObjectID,
		},
	}

	resolveInstruction(engine, g, obj, game.AddPlayerCounter{
		Amount:      game.Fixed(1),
		Player:      game.ObjectControllerReference(game.EventPermanentReference()),
		CounterKind: counter.Poison,
	}, &TurnLog{})

	if got := g.Players[game.Player2].PoisonCounters; got != 1 {
		t.Fatalf("object controller (Player2) poison counters = %d, want 1", got)
	}
	if got := g.Players[game.Player1].PoisonCounters; got != 0 {
		t.Fatalf("ability controller (Player1) poison counters = %d, want 0", got)
	}
}
