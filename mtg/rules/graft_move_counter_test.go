package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestGraftMovesCounterOntoEnteringCreature proves the runtime resolution of the
// Graft move trigger: "Whenever another creature enters, you may move a +1/+1
// counter from this creature onto that creature." reads the counter from the
// Graft source (CounterSourceSelf) and places it on the permanent involved in
// the triggering enters event (EventPermanentReference), not on the source.
func TestGraftMovesCounterOntoEnteringCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Graft Source",
		Types: []types.Card{types.Creature}},
	})
	entering := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Entering Creature",
		Types: []types.Card{types.Creature}},
	})
	source.Counters.Add(counter.PlusOnePlusOne, 2)

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		SourceID:        source.ObjectID,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventPermanentEnteredBattlefield,
			PermanentID: entering.ObjectID,
		},
	}

	resolveInstruction(engine, g, obj, game.MoveCounters{
		Amount:      game.Fixed(1),
		Object:      game.EventPermanentReference(),
		CounterKind: counter.PlusOnePlusOne,
		Source:      game.CounterSourceSpec{Kind: game.CounterSourceSelf},
	}, &TurnLog{})

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("source +1/+1 counters = %d, want 1 (one moved off)", got)
	}
	if got := entering.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("entering creature +1/+1 counters = %d, want 1", got)
	}
}
