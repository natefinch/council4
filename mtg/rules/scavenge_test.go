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

func TestScavengeExilesSourceAndAddsPowerCountersToTarget(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Scavenge Source",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Insect},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		ActivatedAbilities: []game.ActivatedAbility{
			game.ScavengeActivatedAbility(cost.Mana{cost.O(0)}),
		},
	}})
	g.Players[game.Player1].Hand.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)

	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Target Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	act := action.ActivateAbility(cardID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("scavenge ability was not legal from the graveyard")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("scavenge activation failed")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("scavenge source remained in graveyard after exile cost")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("scavenge source was not exiled to pay cost")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("target +1/+1 counters = %d; want 4 (source power)", got)
	}
}
