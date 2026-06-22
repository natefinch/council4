package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestVorelDoublesEachKindOfCounterOnTarget proves the all-kinds counter-doubling
// effect ("Double the number of each kind of counter on target ...", Vorel of the
// Hull Clade): every kind of counter on the chosen target is doubled, while a
// kind absent from the target stays at zero. The AddCounter{AllKinds} runtime
// snapshots the counts before placing any counters so doubling one kind never
// feeds another.
func TestVorelDoublesEachKindOfCounterOnTarget(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Vorel of the Hull Clade",
		Types: []types.Card{types.Creature},
	}})
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Target",
		Types: []types.Card{types.Creature},
	}})
	target.Counters.Add(counter.PlusOnePlusOne, 3)
	target.Counters.Add(counter.Charge, 2)

	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
		Targets:      []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	instruction := game.Instruction{Primitive: game.AddCounter{
		Object:   game.TargetPermanentReference(0),
		AllKinds: true,
	}}
	log := TurnLog{}
	engine.resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &log)

	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 6 {
		t.Fatalf("+1/+1 counters after doubling = %d, want 6", got)
	}
	if got := target.Counters.Get(counter.Charge); got != 4 {
		t.Fatalf("charge counters after doubling = %d, want 4", got)
	}
	if got := target.Counters.Get(counter.Loyalty); got != 0 {
		t.Fatalf("loyalty counters after doubling = %d, want 0", got)
	}
	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("source +1/+1 counters = %d, want 0 (only the target is doubled)", got)
	}
}

// TestVorelDoublesSingleKindOnTarget proves the single-kind target form ("Double
// the number of +1/+1 counters on target creature", Gilder Bairn-adjacent): only
// the named counter kind on the target is doubled, leaving other kinds untouched.
func TestVorelDoublesSingleKindOnTarget(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Doubler",
		Types: []types.Card{types.Creature},
	}})
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Target",
		Types: []types.Card{types.Creature},
	}})
	target.Counters.Add(counter.PlusOnePlusOne, 4)
	target.Counters.Add(counter.Charge, 2)

	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
		Targets:      []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	instruction := game.Instruction{Primitive: game.AddCounter{
		Object:      game.TargetPermanentReference(0),
		CounterKind: counter.PlusOnePlusOne,
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:        game.DynamicAmountObjectCounters,
			Object:      game.TargetPermanentReference(0),
			CounterKind: counter.PlusOnePlusOne,
		}),
	}}
	log := TurnLog{}
	engine.resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &log)

	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 8 {
		t.Fatalf("+1/+1 counters after doubling = %d, want 8", got)
	}
	if got := target.Counters.Get(counter.Charge); got != 2 {
		t.Fatalf("charge counters = %d, want 2 (only +1/+1 is doubled)", got)
	}
}
