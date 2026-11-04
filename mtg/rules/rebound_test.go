package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func reboundSorcery() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:            "Rebound Sorcery",
		Types:           []types.Card{types.Sorcery},
		Colors:          []color.Color{color.Green},
		ManaCost:        opt.Val(cost.Mana{cost.G}),
		StaticAbilities: []game.StaticAbility{game.ReboundStaticBody},
	}}
}

// reboundTriggerPending reports whether a rebound delayed trigger for cardID is
// waiting for its controller's next upkeep.
func reboundTriggerPending(g *game.Game, cardID id.ID) bool {
	for i := range g.DelayedTriggers {
		if g.DelayedTriggers[i].Rebound && g.DelayedTriggers[i].CapturedCardID == cardID {
			return true
		}
	}
	return false
}

// beginUpkeep starts player's upkeep step and puts the abilities that trigger
// at its beginning on the stack. Earlier events were handled by earlier
// priority windows.
func beginUpkeep(e *Engine, g *game.Game, player game.PlayerID, agents [game.NumPlayers]PlayerAgent) {
	g.Turn.TurnNumber++
	beginAdditionalUpkeep(e, g, player, agents)
}

// beginAdditionalUpkeep starts an additional upkeep step for player in the
// current turn (an additional beginning phase, as Sphinx of the Second Sun
// creates), which is still that player's next upkeep.
func beginAdditionalUpkeep(e *Engine, g *game.Game, player game.PlayerID, agents [game.NumPlayers]PlayerAgent) {
	g.Turn.ActivePlayer = player
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep
	g.Turn.PriorityPlayer = player
	g.TriggerEventCursor = len(g.Events)
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	e.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
}

// runReboundUpkeep begins player's upkeep and resolves the rebound triggers on
// top of the stack with every player passing, leaving any recast spell on the
// stack.
func runReboundUpkeep(e *Engine, g *game.Game, player game.PlayerID, agents [game.NumPlayers]PlayerAgent) {
	beginUpkeep(e, g, player, agents)
	resolveReboundTriggers(e, g, agents)
}

// resolveReboundTriggers resolves the rebound triggers on top of the stack.
func resolveReboundTriggers(e *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) {
	for {
		top, ok := g.Stack.Peek()
		if !ok || top.Kind != game.StackTriggeredAbility || !top.Rebound {
			return
		}
		e.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	}
}

func TestReboundExilesSpellFromHandAndRecastsAtUpkeep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, reboundSorcery())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyActionWithChoices(
		g,
		game.Player1,
		action.CastSpell(cardID, nil, 0, nil),
		[game.NumPlayers]PlayerAgent{},
		&TurnLog{},
	) {
		t.Fatal("casting rebound sorcery from hand failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("rebounding spell went to graveyard instead of being exiled")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("rebounding spell was not exiled on resolution")
	}
	if len(g.DelayedTriggers) != 1 || !g.DelayedTriggers[0].Rebound ||
		g.DelayedTriggers[0].Controller != game.Player1 || g.DelayedTriggers[0].CapturedCardID != cardID {
		t.Fatalf("delayed triggers = %+v, want one Player1 rebound trigger for the card", g.DelayedTriggers)
	}

	runReboundUpkeep(engine, g, game.Player1, [game.NumPlayers]PlayerAgent{})

	if reboundTriggerPending(g, cardID) {
		t.Fatal("rebound trigger was not consumed at upkeep")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.SourceZone != zone.Exile {
		t.Fatalf("recast stack object = %+v ok=%v, want spell cast from exile", obj, ok)
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("recast spell is still in exile")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if reboundTriggerPending(g, cardID) {
		t.Fatal("recast spell re-rebounded; rebound must only fire when cast from hand")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("recast spell (cast from exile) did not go to the graveyard")
	}
}

func TestReboundDoesNotExileWhenNotCastFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: reboundSorcery(), Owner: game.Player1}

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Face:       game.FaceFront,
		Controller: game.Player1,
		SourceZone: zone.Graveyard,
	}
	g.Stack.Push(obj)

	engine.resolveTopOfStack(g, &TurnLog{})

	if reboundTriggerPending(g, cardID) {
		t.Fatal("spell cast from a zone other than hand must not rebound")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("non-hand-cast rebound spell did not go to the graveyard")
	}
}
