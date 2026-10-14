package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestMindslaverRoutesLivePriorityThroughControllingSearchAgent(t *testing.T) {
	g, e, _ := controlledMindslaverGame(t)
	land := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Controlled land", Types: []types.Card{types.Land},
	}})
	searcher := &stubSearchAgent{}
	victim := &controlScript{}
	log := &TurnLog{}
	e.runPriorityLoop(g, [game.NumPlayers]PlayerAgent{searcher, victim}, log)
	if !searcher.searched || len(victim.actions) != 0 || permanentByCardID(g, land) == nil ||
		permanentByCardID(g, land).Controller != game.Player2 {
		t.Fatal("live search bypassed controlling agent or changed the object's controller")
	}
	found := false
	for _, act := range log.Actions {
		if act.Action.Kind == action.ActionPlayLand {
			found = true
			if act.Player != game.Player2 || act.DecisionMaker != game.Player1 {
				t.Fatal("searched action log conflated actor and decision maker")
			}
		}
	}
	if !found {
		t.Fatal("controlling search agent's action did not reach the live Engine")
	}
}

func TestMindslaverOverwrittenPendingEffectNeverRevivesAfterDeparture(t *testing.T) {
	g, e, def := mindslaverGame(t)
	activateMindslaver(t, g, e, def, game.Player1, game.Player2)
	activateMindslaver(t, g, e, def, game.Player3, game.Player2)
	if len(g.PlayerControlEffects) != 1 || !e.Concede(g, game.Player3) {
		t.Fatal("competing pending control did not overwrite")
	}
	e.advanceToNextTurn(g)
	if g.Turn.ActivePlayer != game.Player2 || g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player2 ||
		len(g.PlayerControlEffects) != 0 {
		t.Fatal("controller departure revived overwritten pending control")
	}
}

func TestMindslaverControlledActivatorRemainsEffectController(t *testing.T) {
	g, e, def := controlledMindslaverGame(t)
	activateMindslaver(t, g, e, def, game.Player2, game.Player3)
	if g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player1 {
		t.Fatal("scheduling a later turn replaced current-turn control")
	}
	e.advanceToNextTurn(g)
	if g.Turn.ActivePlayer != game.Player3 || g.DecisionMaker(game.Player3, game.DecisionRules) != game.Player2 ||
		g.DecisionMaker(game.Player2, game.DecisionRules) != game.Player2 {
		t.Fatal("decision maker, not the actual activator, became the later effect's controller")
	}
}

func TestMindslaverAuthoredContextCannotBypassDecisionControl(t *testing.T) {
	g, e, _ := controlledMindslaverGame(t)
	controller, victim := &controlScript{}, &controlScript{}
	ctx := &programContext{active: true, engine: e, game: g, obj: &game.StackObject{Controller: game.Player2},
		log: &TurnLog{}, agents: [game.NumPlayers]PlayerAgent{controller, victim}}
	request := game.ChoiceRequest{Kind: game.ChoiceMay, Player: game.Player2,
		DecisionKind: game.DecisionPersonal, Options: []game.ChoiceOption{{Index: 0}, {Index: 1}},
		MinChoices: 1, MaxChoices: 1}
	if _, err := ctx.Ask(request); err == nil || len(controller.choices) != 0 || len(victim.choices) != 0 {
		t.Fatal("card code could reclassify a rules choice as personal")
	}
	request.DecisionKind = game.DecisionRules
	if _, err := ctx.Ask(request); err != nil || len(controller.choices) != 1 || len(victim.choices) != 0 {
		t.Fatal("authored rule choice did not use common control dispatch")
	}
	ctx.active = false
	if _, err := ctx.ControlNextTurn(game.ControllerReference()); err == nil || len(g.PlayerControlEffects) != 1 {
		t.Fatal("expired context could schedule mutable decision state")
	}
}
