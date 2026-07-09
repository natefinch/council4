package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/opt"
)

// TestAddCounterEventPermanentDiedPowerReadsLastKnown proves the Death's
// Presence shape: "Whenever a creature you control dies, put X +1/+1 counters on
// target creature you control, where X is the power of the creature that died."
// The counted amount reads the dying creature's power from its last-known
// information (CR 603.10, CR 608.2h) after it has left the battlefield, and the
// counters land on the chosen target rather than the dead creature.
func TestAddCounterEventPermanentDiedPowerReadsLastKnown(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	recipient := addCreaturePermanent(g, game.Player1)
	diedID := g.IDGen.Next()
	rememberLastKnown(g, &game.ObjectSnapshot{
		ObjectID: diedID,
		Power:    opt.Val(5),
	})

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventPermanentDied,
			PermanentID: diedID,
		},
		Targets: []game.Target{game.PermanentTarget(recipient.ObjectID)},
	}

	resolveInstruction(engine, g, obj, game.AddCounter{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountObjectPower,
			Multiplier: 1,
			Object:     game.EventPermanentReference(),
		}),
		Object:      game.TargetPermanentReference(0),
		CounterKind: counter.PlusOnePlusOne,
	}, &TurnLog{})

	if got := recipient.Counters.Get(counter.PlusOnePlusOne); got != 5 {
		t.Fatalf("recipient +1/+1 counters = %d, want 5 (dead creature's last-known power)", got)
	}
}
