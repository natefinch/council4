package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func suspendCostForCard(card *game.CardDef) (cost.Mana, int, bool) {
	for i := range card.ActivatedAbilities {
		if manaCost, counters, ok := game.ActivatedBodySuspendInfo(&card.ActivatedAbilities[i]); ok {
			return manaCost, counters, true
		}
	}
	for i := range card.StaticAbilities {
		if manaCost, counters, ok := game.StaticBodySuspendInfo(&card.StaticAbilities[i]); ok {
			return manaCost, counters, true
		}
	}
	return nil, 0, false
}

func (*Engine) canSuspendCard(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
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
	manaCost, _, ok := suspendCostForCard(spellDef)
	if !ok || !canCastAtCurrentTiming(g, playerID, spellDef) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &manaCost})
}

func (e *Engine) applySuspendCard(g *game.Game, playerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !e.canSuspendCard(g, playerID, cardID) {
		return false
	}
	card, _ := g.GetCardInstance(cardID)
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	manaCost, counters, _ := suspendCostForCard(spellDef)
	prefs := e.paymentPreferencesForCost(g, playerID, &manaCost, nil, 0, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &manaCost, Prefs: prefs}) {
		return false
	}
	destination, moved := moveCardBetweenZonesWithEvent(g, game.Event{
		Controller: playerID, Player: card.Owner, CardID: cardID,
		FromZone: zone.Hand, ToZone: zone.Exile,
	}, zoneMoveOptions{})
	if !moved {
		return false
	}
	if destination == zone.Exile {
		if g.SuspendedCards == nil {
			g.SuspendedCards = make(map[id.ID]game.SuspendedCard)
		}
		g.SuspendedCards[cardID] = game.SuspendedCard{
			Owner: card.Owner, Controller: playerID, TimeCounters: counters,
		}
	}
	return true
}

func (e *Engine) legalSuspendActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		if e.canSuspendCard(g, playerID, cardID) {
			actions = append(actions, actionBuild.suspendCard(cardID))
		}
	}
	return actions
}

func (e *Engine) processSuspendUpkeep(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	for _, cardID := range suspendedCardIDsInOrder(g) {
		suspended := g.SuspendedCards[cardID]
		if suspended.Controller != playerID || suspended.TimeCounters <= 0 {
			continue
		}
		suspended.TimeCounters--
		if suspended.TimeCounters > 0 {
			g.SuspendedCards[cardID] = suspended
			continue
		}
		g.SuspendedCards[cardID] = suspended
		e.castSuspendedCard(g, playerID, cardID, agents, log)
	}
}

func suspendedCardIDsInOrder(g *game.Game) []id.ID {
	ids := make([]id.ID, 0, len(g.SuspendedCards))
	for cardID := range g.SuspendedCards {
		ids = append(ids, cardID)
	}
	slices.Sort(ids)
	return ids
}

// castSuspendedCard lets the controller cast a card whose last time counter
// was removed from exile without paying its mana cost (CR 702.62a). The cast
// follows the alternative-cost rules (CR 702.62d, 601.2b, 601.2f-h):
// additional costs, cost increases and reductions still apply, and a creature
// spell cast this way gains haste. If the card isn't cast it remains exiled.
func (e *Engine) castSuspendedCard(g *game.Game, playerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Exile.Contains(cardID) {
		return false
	}
	return e.chooseAndCastFreeSpell(g, player, playerID, cardID, zone.Exile, game.Selection{}, opt.V[int]{}, freeCastSpell{suspend: true}, true, agents, log)
}
