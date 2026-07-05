package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// plotCostForCard returns the Plot mana cost printed on the card's front face
// (CR 718), or (nil, false) when the card has no Plot keyword.
func plotCostForCard(card *game.CardDef) (cost.Mana, bool) {
	return card.PlotCost()
}

// canPlotCard reports whether the player may plot the hand card cardID: plotting
// is a special action usable any time the player could cast a sorcery (CR 718.2),
// so it requires priority, sorcery timing, the card in the player's hand, a Plot
// keyword, and the ability to pay the plot cost.
func (*Engine) canPlotCard(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Contains(cardID) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	manaCost, ok := plotCostForCard(spellDef)
	if !ok || !isSorcerySpeed(g, playerID) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &manaCost})
}

// applyPlotCard resolves a plot special action: the player pays the plot cost and
// exiles the card from hand, recording the plot turn so the card may be cast from
// exile without paying its mana cost on a later turn (CR 718.2).
func (e *Engine) applyPlotCard(g *game.Game, playerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !e.canPlotCard(g, playerID, cardID) {
		return false
	}
	player := g.Players[playerID]
	card, _ := g.GetCardInstance(cardID)
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	manaCost, _ := plotCostForCard(spellDef)
	prefs := e.paymentPreferencesForCost(g, playerID, &manaCost, nil, 0, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &manaCost, Prefs: prefs}) {
		return false
	}
	if !player.Hand.Remove(cardID) {
		return false
	}
	player.Exile.Add(cardID)
	if g.PlottedCards == nil {
		g.PlottedCards = make(map[id.ID]int)
	}
	g.PlottedCards[cardID] = g.Turn.TurnNumber
	emitZoneChangeEvent(g, game.Event{
		Controller: playerID,
		Player:     card.Owner,
		CardID:     cardID,
		FromZone:   zone.Hand,
		ToZone:     zone.Exile,
	})
	return true
}

// legalPlotActions returns the plot special actions available to the player: one
// for each hand card whose Plot keyword may currently be activated.
func (e *Engine) legalPlotActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		if e.canPlotCard(g, playerID, cardID) {
			actions = append(actions, actionBuild.plotCard(cardID))
		}
	}
	return actions
}

// cardIsPlottedInExile reports whether cardID is a card plotted into exile that
// may now be cast from exile: it must be recorded as plotted and the current turn
// must be later than the turn it was plotted (CR 718.2 "on a later turn").
func cardIsPlottedInExile(g *game.Game, cardID id.ID) bool {
	plotTurn, ok := g.PlottedCards[cardID]
	return ok && g.Turn.TurnNumber > plotTurn
}
