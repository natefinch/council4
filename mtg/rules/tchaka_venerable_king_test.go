package rules

import (
	"testing"

	cardt "github.com/natefinch/council4/mtg/cards/t"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// tchakaGraveyardActivateAction scans the real legal-action enumeration for
// T'Chaka's graveyard become-monarch activated ability, identified by its source
// card in the graveyard. Driving legality through engine.legalActions ensures the
// commander-control activation condition is exercised by the real
// activation-legality path rather than a hand-constructed activation.
func tchakaGraveyardActivateAction(engine *Engine, g *game.Game, cardID id.ID) (action.Action, bool) {
	for _, act := range engine.legalActions(g, game.Player1) {
		payload, ok := act.ActivateAbilityPayload()
		if ok && payload.SourceID == cardID {
			return act, true
		}
	}
	return action.Action{}, false
}

// stageTchakaGraveyardActivation puts the real T'Chaka, Venerable King card into
// Player1's graveyard with Player1 holding priority in their precombat main phase
// and three colorless mana available to pay the ability's {3} cost. The returned
// id is T'Chaka's card in the graveyard.
func stageTchakaGraveyardActivation(g *game.Game) id.ID {
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	return addCardToGraveyard(g, game.Player1, cardt.TChakaVenerableKing())
}

// TestTchakaGraveyardBecomeMonarchRequiresCommander proves T'Chaka, Venerable
// King's "{3}, Exile this card from your graveyard: You become the monarch.
// Activate only if you control your commander." ability is legal through the real
// activation-legality enumeration only while its controller controls their
// commander, and that activating it makes them the monarch.
func TestTchakaGraveyardBecomeMonarchRequiresCommander(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := stageTchakaGraveyardActivation(g)

	if _, ok := tchakaGraveyardActivateAction(engine, g, cardID); ok {
		t.Fatal("graveyard become-monarch ability was legal while not controlling a commander")
	}

	commander := addCommanderPermanent(g, game.Player1)

	act, ok := tchakaGraveyardActivateAction(engine, g, cardID)
	if !ok {
		t.Fatal("graveyard become-monarch ability was not legal while controlling a commander")
	}

	// Losing control of the commander again removes the ability from the legal
	// set, confirming the gate tracks live commander control rather than a
	// one-time check.
	commander.PhasedOut = true
	if _, ok := tchakaGraveyardActivateAction(engine, g, cardID); ok {
		t.Fatal("graveyard become-monarch ability remained legal after the commander phased out")
	}
	commander.PhasedOut = false

	if g.Players[game.Player1].IsMonarch {
		t.Fatal("Player1 was already the monarch before activating the ability")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(graveyard become-monarch) = false, want true")
	}
	engine.resolveTopOfStack(g, nil)
	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("Player1 did not become the monarch after activating the ability")
	}
}
