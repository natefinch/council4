package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
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

// canSuspendCard reports whether the player may suspend the hand card cardID
// (CR 702.62a, 116.2f): they could begin to cast it, and its suspend cost is
// payable now or with payment-time mana abilities (CR 605.3a).
func (e *Engine) canSuspendCard(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	c, _, _ := suspendCost(g, playerID, cardID)
	return e.activationFeasible(g, playerID, c)
}

// suspendCost is the cost of suspending the hand card cardID with its suspend
// mana cost and time counters, or the zero activationCost when that isn't
// legal.
func suspendCost(g *game.Game, playerID game.PlayerID, cardID id.ID) (activationCost, cost.Mana, int) {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return activationCost{}, nil, 0
	}
	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Contains(cardID) {
		return activationCost{}, nil, 0
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return activationCost{}, nil, 0
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	manaCost, counters, ok := suspendCostForCard(spellDef)
	if !ok || !canCastAtCurrentTiming(g, playerID, spellDef) {
		return activationCost{}, nil, 0
	}
	return specialActionCost(playerID, &manaCost, handCardPresent(playerID, cardID)), manaCost, counters
}

// applySuspendCard resolves a suspend special action: the player pays the
// suspend cost, activating mana abilities while paying if they choose
// (CR 605.3a), and exiles the card with its time counters (CR 702.62a).
func (e *Engine) applySuspendCard(g *game.Game, playerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	defer g.OpenLibraryTopWindow()()
	c, manaCost, counters := suspendCost(g, playerID, cardID)
	if !e.activationFeasible(g, playerID, c) {
		return actionFailed
	}
	card, _ := g.GetCardInstance(cardID)
	if outcome := e.paySpecialActionCost(g, playerID, c, &manaCost, agents, log); outcome != actionApplied {
		return outcome
	}
	destination, moved := moveCardBetweenZonesWithEvent(g, game.Event{
		Controller: playerID, Player: card.Owner, CardID: cardID,
		FromZone: zone.Hand, ToZone: zone.Exile,
	}, zoneMoveOptions{})
	if !moved {
		return actionFailed
	}
	if destination == zone.Exile {
		if g.SuspendedCards == nil {
			g.SuspendedCards = make(map[id.ID]game.SuspendedCard)
		}
		g.SuspendedCards[cardID] = game.SuspendedCard{
			Owner: card.Owner, Controller: playerID, TimeCounters: counters,
		}
	}
	return actionApplied
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
