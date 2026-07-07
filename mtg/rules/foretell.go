package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// foretellExileCost is the fixed generic cost paid to foretell a card
// (CR 702.144a): {2}, regardless of the card's foretell cost.
func foretellExileCost() cost.Mana {
	return cost.Mana{cost.O(2)}
}

// foretellCostForCard returns the Foretell cast cost printed on the card's front
// face (CR 702.144), or (nil, false) when the card has no Foretell keyword. This
// is the cost paid to cast the card from exile after foretelling it, not the
// fixed {2} paid to foretell it.
func foretellCostForCard(card *game.CardDef) (cost.Mana, bool) {
	return card.ForetellCost()
}

// canForetellCard reports whether the player may foretell the hand card cardID:
// foretelling is a special action usable any time the player has priority during
// their own turn (CR 702.144a), so it requires the player's turn, priority, the
// card in the player's hand, a Foretell keyword, and the ability to pay the fixed
// {2} foretell cost. Unlike Plot, foretelling is not restricted to sorcery speed.
func (*Engine) canForetellCard(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	if !canAct(g, playerID) || playerID != g.Turn.ActivePlayer || playerID != g.Turn.PriorityPlayer {
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
	if _, ok := foretellCostForCard(spellDef); !ok {
		return false
	}
	exileCost := foretellExileCost()
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &exileCost})
}

// applyForetellCard resolves a foretell special action: the player pays the fixed
// {2} and exiles the card from hand, recording the foretell turn so the card may
// be cast from exile for its foretell cost on a later turn (CR 702.144a).
func (e *Engine) applyForetellCard(g *game.Game, playerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !e.canForetellCard(g, playerID, cardID) {
		return false
	}
	player := g.Players[playerID]
	card, _ := g.GetCardInstance(cardID)
	exileCost := foretellExileCost()
	prefs := e.paymentPreferencesForCost(g, playerID, &exileCost, nil, 0, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &exileCost, Prefs: prefs}) {
		return false
	}
	if !player.Hand.Remove(cardID) {
		return false
	}
	player.Exile.Add(cardID)
	if g.ForetoldCards == nil {
		g.ForetoldCards = make(map[id.ID]int)
	}
	g.ForetoldCards[cardID] = g.Turn.TurnNumber
	emitZoneChangeEvent(g, game.Event{
		Controller: playerID,
		Player:     card.Owner,
		CardID:     cardID,
		FromZone:   zone.Hand,
		ToZone:     zone.Exile,
	})
	return true
}

// legalForetellActions returns the foretell special actions available to the
// player: one for each hand card whose Foretell keyword may currently be
// activated.
func (e *Engine) legalForetellActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.ActivePlayer || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		if e.canForetellCard(g, playerID, cardID) {
			actions = append(actions, actionBuild.foretellCard(cardID))
		}
	}
	return actions
}

// cardIsForetoldInExile reports whether cardID is a card foretold into exile that
// may now be cast from exile: it must be recorded as foretold and the current
// turn must be later than the turn it was foretold (CR 702.144b "on a later
// turn").
func cardIsForetoldInExile(g *game.Game, cardID id.ID) bool {
	foretoldTurn, ok := g.ForetoldCards[cardID]
	return ok && g.Turn.TurnNumber > foretoldTurn
}
