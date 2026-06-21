package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestMossbornHydraDoublesPlusOneCounters proves the counter-doubling effect
// ("Double the number of +1/+1 counters on this creature.", Mossborn Hydra): a
// creature with K +1/+1 counters ends with 2K after the effect resolves, because
// the placement adds counters equal to the source's current count read through
// DynamicAmountObjectCounters.
func TestMossbornHydraDoublesPlusOneCounters(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Mossborn Hydra",
		Types: []types.Card{types.Creature},
	}})
	source.Counters.Add(counter.PlusOnePlusOne, 3)
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}
	instruction := game.Instruction{Primitive: game.AddCounter{
		Object:      game.SourcePermanentReference(),
		CounterKind: counter.PlusOnePlusOne,
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:        game.DynamicAmountObjectCounters,
			Object:      game.SourcePermanentReference(),
			CounterKind: counter.PlusOnePlusOne,
		}),
	}}
	log := TurnLog{}
	engine.resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &log)

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 6 {
		t.Fatalf("+1/+1 counters after doubling = %d, want 6", got)
	}
}
