package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestAddCounterOnEventPermanentTargetsTriggeringPermanent proves the runtime
// resolution of AddCounter{Object: EventPermanentReference()} places the
// counters on the permanent involved in the triggering event and on no other
// permanent. This is the recipient the "put a +1/+1 counter on it" / "on that
// creature" trigger lowering relies on.
func TestAddCounterOnEventPermanentTargetsTriggeringPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	triggering := addCreaturePermanent(g, game.Player1)
	bystander := addCreaturePermanent(g, game.Player1)

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventPermanentEnteredBattlefield,
			PermanentID: triggering.ObjectID,
		},
	}

	resolveInstruction(engine, g, obj, game.AddCounter{
		Amount:      game.Fixed(2),
		Object:      game.EventPermanentReference(),
		CounterKind: counter.PlusOnePlusOne,
	}, &TurnLog{})

	if got := triggering.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("triggering creature +1/+1 counters = %d, want 2", got)
	}
	if got := bystander.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("bystander creature +1/+1 counters = %d, want 0", got)
	}
}
