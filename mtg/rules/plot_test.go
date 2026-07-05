package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// plotSorcery is a minimal castable sorcery carrying the Plot keyword with the
// given plot cost. Its printed mana cost is high so the free plotted cast is
// distinguishable from a normal cast in tests.
func plotSorcery(plotCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:         "Plot Sorcery",
		Types:        []types.Card{types.Sorcery},
		ManaCost:     opt.Val(cost.Mana{cost.O(9)}),
		SpellAbility: opt.Val(game.AbilityContent{}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.PlotKeyword{Cost: plotCost}},
		}},
	}}
}

func setSorceryTiming(g *game.Game, playerID game.PlayerID) {
	g.Turn.ActivePlayer = playerID
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = playerID
}

// TestLegalActionsIncludePlotFromHand verifies a hand card with the Plot keyword
// offers a plot special action at sorcery speed when its cost can be paid.
func TestLegalActionsIncludePlotFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, plotSorcery(cost.Mana{cost.G}))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setSorceryTiming(g, game.Player1)

	legal := engine.legalActions(g, game.Player1)

	if !actionsContain(legal, action.PlotCard(cardID)) {
		t.Fatalf("legal actions = %+v, want plot action", legal)
	}
}

// TestPlotActionPaysCostAndExiles verifies plotting pays the plot cost, moves the
// card from hand to exile, and records the plot turn.
func TestPlotActionPaysCostAndExiles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, plotSorcery(cost.Mana{cost.G}))
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	setSorceryTiming(g, game.Player1)
	g.Turn.TurnNumber = 3

	if !engine.applyAction(g, game.Player1, action.PlotCard(cardID)) {
		t.Fatal("plot action failed")
	}
	if !forest.Tapped {
		t.Fatal("plot cost did not tap the mana source")
	}
	if g.Players[game.Player1].Hand.Contains(cardID) || !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("plotted card did not move from hand to exile")
	}
	if plotTurn, ok := g.PlottedCards[cardID]; !ok || plotTurn != 3 {
		t.Fatalf("PlottedCards[card] = (%d,%v), want (3,true)", plotTurn, ok)
	}
}

// TestPlottedCardNotCastableSameTurn verifies a card plotted this turn cannot be
// cast from exile until a later turn (CR 718.2).
func TestPlottedCardNotCastableSameTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, plotSorcery(cost.Mana{cost.G}))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setSorceryTiming(g, game.Player1)
	g.Turn.TurnNumber = 3

	if !engine.applyAction(g, game.Player1, action.PlotCard(cardID)) {
		t.Fatal("plot action failed")
	}
	if cardIsPlottedInExile(g, cardID) {
		t.Fatal("card is castable from exile the same turn it was plotted")
	}
	legal := engine.legalActions(g, game.Player1)
	if actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("cast-from-exile offered the same turn the card was plotted")
	}
}

// TestPlottedCardCastableFreeLaterTurn verifies a plotted card is castable from
// exile without paying its mana cost on a later turn, at sorcery speed.
func TestPlottedCardCastableFreeLaterTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, plotSorcery(cost.Mana{cost.G}))
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	setSorceryTiming(g, game.Player1)
	g.Turn.TurnNumber = 3

	if !engine.applyAction(g, game.Player1, action.PlotCard(cardID)) {
		t.Fatal("plot action failed")
	}

	// Advance to a later turn with no available mana (the forest is tapped): the
	// plotted card must still be castable because the cast is free.
	g.Turn.TurnNumber = 4
	setSorceryTiming(g, game.Player1)
	if !forest.Tapped {
		t.Fatal("precondition: forest should still be tapped from the plot cost")
	}

	if !cardIsPlottedInExile(g, cardID) {
		t.Fatal("card is not castable from exile on a later turn")
	}
	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, want free cast-from-exile of the plotted card", legal)
	}

	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("casting the plotted card from exile failed")
	}
	if _, onStack := g.Stack.Peek(); !onStack {
		t.Fatal("plotted spell was not put on the stack")
	}
	if _, ok := g.PlottedCards[cardID]; ok {
		t.Fatal("PlottedCards entry not cleared after casting the plotted card")
	}
}
