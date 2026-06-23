package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestRemoveCounterChooseKindRemovesLonePresentKind covers the kind-unspecified
// "remove a counter from target permanent" form (Ferropede): with a single
// counter kind present, ChooseKind removes one counter of that kind without a
// prompt.
func TestRemoveCounterChooseKindRemovesLonePresentKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Charged Relic",
		Types: []types.Card{types.Artifact}},
	})
	target.Counters.Add(counter.Charge, 2)
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		SourceID:   target.ObjectID,
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	resolveInstruction(engine, g, obj, game.RemoveCounter{
		Amount:     game.Fixed(1),
		Object:     game.TargetPermanentReference(0),
		ChooseKind: true,
	}, &TurnLog{})

	if got := target.Counters.Get(counter.Charge); got != 1 {
		t.Fatalf("charge counters = %d, want 1", got)
	}
}

// TestRemoveCounterNamedKindRemovesThatKind covers the named-kind activated form
// ("Remove a -1/-1 counter from target creature.", Chainbreaker): the named kind
// is removed while other kinds are left untouched.
func TestRemoveCounterNamedKindRemovesThatKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Burdened Beast",
		Types: []types.Card{types.Creature}},
	})
	target.Counters.Add(counter.MinusOneMinusOne, 2)
	target.Counters.Add(counter.Charge, 1)
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		SourceID:   target.ObjectID,
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	resolveInstruction(engine, g, obj, game.RemoveCounter{
		Amount:      game.Fixed(1),
		Object:      game.TargetPermanentReference(0),
		CounterKind: counter.MinusOneMinusOne,
	}, &TurnLog{})

	if got := target.Counters.Get(counter.MinusOneMinusOne); got != 1 {
		t.Fatalf("-1/-1 counters = %d, want 1", got)
	}
	if got := target.Counters.Get(counter.Charge); got != 1 {
		t.Fatalf("charge counters = %d, want 1 (named-kind removal leaves other kinds)", got)
	}
}
