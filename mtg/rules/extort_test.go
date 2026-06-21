package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// extortDrainInstructions models the resolved Extort ability: pay {W/B}, and if
// the payment is made, each opponent loses 1 life and the controller gains that
// much. The drain effects are gated on the optional payment having succeeded.
func extortDrainInstructions() []game.Instruction {
	paid := opt.Val(game.InstructionResultGate{Key: "controller-paid", Succeeded: game.TriTrue})
	manaCost := cost.Mana{cost.HybridMana(mana.W, mana.B)}
	return []game.Instruction{
		{
			Primitive:     game.Pay{Payment: game.ResolutionPayment{Prompt: "Pay {W/B}?", ManaCost: opt.Val(manaCost)}},
			PublishResult: "controller-paid",
		},
		{
			Primitive:     game.LoseLife{PlayerGroup: game.OpponentsReference(), Amount: game.Fixed(1)},
			PublishResult: "life-change",
			ResultGate:    paid,
		},
		{
			Primitive: game.GainLife{
				Player: game.ControllerReference(),
				Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "life-change"}),
			},
			ResultGate: paid,
		},
	}
}

// TestExtortPaymentDrainsEachOpponentAndGainsLife proves that paying the
// optional {W/B} Extort cost makes each opponent lose 1 life and the controller
// gain the total drained.
func TestExtortPaymentDrainsEachOpponentAndGainsLife(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBasicLandPermanent(g, game.Player1, types.Plains)
	addInstructionSpellToStack(g, extortDrainInstructions())

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Life; got != 39 {
			t.Fatalf("opponent %d life = %d, want 39 (paid Extort drains 1)", playerID, got)
		}
	}
	if got := g.Players[game.Player1].Life; got != 43 {
		t.Fatalf("controller life = %d, want 43 (gained total 3)", got)
	}
}

// TestExtortDeclinedPaymentSkipsDrain proves that declining the optional {W/B}
// payment leaves every life total unchanged.
func TestExtortDeclinedPaymentSkipsDrain(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBasicLandPermanent(g, game.Player1, types.Plains)
	addInstructionSpellToStack(g, extortDrainInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Life; got != 40 {
			t.Fatalf("opponent %d life = %d, want 40 (declined Extort)", playerID, got)
		}
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("controller life = %d, want 40 (declined Extort)", got)
	}
}
