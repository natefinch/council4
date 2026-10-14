package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestAbundanceReplacedTurnBasedDrawLeavesLaterFirstDrawExempt(t *testing.T) {
	t.Parallel()
	g, e, _ := abundanceGame(t)
	abundanceLibrary(g, "Land1", "A", "B", "C", "D", "E")
	drawDoublingPermanent(g, game.Player1, 2, true)
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepDraw
	emitBeginningOfStepEvent(g, game.StepDraw)
	e.drawCardWithReplacements(g, game.Player1, [game.NumPlayers]PlayerAgent{&abundanceChoices{}}, &TurnLog{}, true)
	if len(abundanceEvents(g, game.EventCardDrawn)) != 0 {
		t.Fatal("turn-based draw was not replaced")
	}
	// A later spell is a separate pending event, after the invocation and its
	// lineage expired. This still is the first actual draw of this draw step.
	addEffectSpellToStack(g, game.Player1, game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}, nil)
	e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{&abundanceChoices{accepts: []bool{false}}}, &TurnLog{})
	draws := abundanceEvents(g, game.EventCardDrawn)
	if len(draws) != 1 || g.Players[game.Player1].Hand.Size() != 2 {
		t.Fatal("draw-step exemption expired with the replaced turn-based draw")
	}
	e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(abundanceEvents(g, game.EventCardDrawn)) != 3 {
		t.Fatal("a subsequent real draw in the same step was not doubled")
	}
	// An extra draw step resets the first-draw identity; history from the
	// preceding step must not disable its exemption.
	emitBeginningOfStepEvent(g, game.StepDraw)
	e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(abundanceEvents(g, game.EventCardDrawn)) != 4 {
		t.Fatal("extra draw step did not reset its own first actual draw")
	}
}
