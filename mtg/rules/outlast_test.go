package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func outlastCreature() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:               "Outlast Bear",
		Types:              []types.Card{types.Creature},
		Power:              opt.Val(game.PT{Value: 2}),
		Toughness:          opt.Val(game.PT{Value: 2}),
		ActivatedAbilities: []game.ActivatedAbility{game.OutlastActivatedAbility(cost.Mana{cost.W})},
	}}
}

// TestOutlastTapsAndAddsCounterAtSorcerySpeed verifies that the Outlast keyword
// ability taps the creature, pays its mana cost, resolves a +1/+1 counter onto
// the creature, and is only legal at sorcery speed (CR 702.105).
func TestOutlastTapsAndAddsCounterAtSorcerySpeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, outlastCreature())
	plains := addBasicLandPermanent(g, game.Player1, types.Plains)
	g.Turn.PriorityPlayer = game.Player1

	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("outlast activation was legal outside sorcery speed")
	}

	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("outlast activation was not legal at sorcery speed")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(outlast) = false, want true")
	}
	if !source.Tapped {
		t.Fatal("outlast creature was not tapped to pay the {T} cost")
	}
	if !plains.Tapped {
		t.Fatal("plains was not tapped to pay the outlast mana cost")
	}
	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("outlast counter before resolution = %d, want 0", got)
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, nil)
	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("outlast counter after resolution = %d, want 1", got)
	}
}
