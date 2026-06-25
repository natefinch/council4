package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
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
	rebound, ok := g.ReboundCards[cardID]
	if !ok || rebound.Controller != game.Player1 {
		t.Fatalf("rebound tracking = %+v ok=%v, want controller Player1", rebound, ok)
	}

	engine.processReboundUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if _, stillTracked := g.ReboundCards[cardID]; stillTracked {
		t.Fatal("rebound permission was not consumed at upkeep")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.SourceZone != zone.Exile {
		t.Fatalf("recast stack object = %+v ok=%v, want spell cast from exile", obj, ok)
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("recast spell is still in exile")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, stillTracked := g.ReboundCards[cardID]; stillTracked {
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

	if _, tracked := g.ReboundCards[cardID]; tracked {
		t.Fatal("spell cast from a zone other than hand must not rebound")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("non-hand-cast rebound spell did not go to the graveyard")
	}
}
