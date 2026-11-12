package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
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
		change := newExileCardCounterChange(g, playerID)
		change.add(cardID, counter.Time, counters)
		change.publish()
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

// processSuspendUpkeep removes a time counter from each card the active player
// owns that is suspended (CR 702.62a, 702.62b); a suspended card's abilities
// are controlled by its owner (CR 108.4a). Each removal is published as an
// EventCountersRemoved for the exiled card, so removing the last one triggers
// suspend's cast ability.
func processSuspendUpkeep(g *game.Game, playerID game.PlayerID) {
	for _, cardID := range suspendedCardIDsInOrder(g, playerID) {
		if !cardIsSuspended(g, cardID) {
			continue
		}
		change := newExileCardCounterChange(g, playerID)
		change.remove(cardID, counter.Time, 1)
		change.publish()
	}
}

// suspendedCardIDsInOrder lists the suspended cards owned by playerID in card
// ID order.
func suspendedCardIDsInOrder(g *game.Game, playerID game.PlayerID) []id.ID {
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var ids []id.ID
	for _, cardID := range player.Exile.All() {
		if card, ok := g.GetCardInstance(cardID); ok && card.Owner == playerID && cardIsSuspended(g, cardID) {
			ids = append(ids, cardID)
		}
	}
	slices.Sort(ids)
	return ids
}

// suspendCastAbility is suspend's intrinsic triggered ability that functions
// in exile (CR 702.62a). It is shared immutable data; its behavior is
// resolveSuspendCastTrigger, selected by StackObject.SuspendCast.
var suspendCastAbility = &game.TriggeredAbility{
	Text: "When the last time counter is removed from this card, if it's exiled, you may play it without paying its mana cost if able. If you don't, it remains exiled. If you cast a creature spell this way, it gains haste until you lose control of the spell or the permanent it becomes.",
}

// suspendCastTriggers detects suspend's cast ability (CR 702.62a) from an
// event that removed the last time counter from an exiled card with suspend,
// however the counter was removed. The intervening "if it's exiled" condition
// requires the card to still be in exile as the incarnation the event names
// (CR 603.4, 400.7). The ability is controlled by the card's owner (CR 108.4a)
// and is bound to that exiled card's zone version.
func suspendCastTriggers(g *game.Game, event game.Event) []pendingTriggeredAbility {
	if event.Kind != game.EventCountersRemoved || !exileCardCounterEvent(event) ||
		event.CounterKind != counter.Time || event.PreviousCounterAmount <= 0 ||
		event.Amount < event.PreviousCounterAmount {
		return nil
	}
	card, ok := g.GetCardInstance(event.CardID)
	if !ok || card.ZoneVersion != event.CardZoneVersion || !exiledCardHasSuspend(g, card) {
		return nil
	}
	return []pendingTriggeredAbility{{
		controller: card.Owner, sourceID: card.ID, sourceCardID: card.ID,
		sourceZone: zone.Exile, sourceZoneVersion: card.ZoneVersion, face: game.FaceFront,
		inline: suspendCastAbility, suspendCast: true, event: event, hasEvent: true,
		triggerMultiplierCaptured: true,
	}}
}

// resolveSuspendCastTrigger resolves suspend's cast ability (CR 702.62a). The
// intervening "if it's exiled" condition is rechecked (CR 603.4): a card that
// left exile, even if it has since returned, is a new object the ability
// doesn't affect (CR 400.7). Otherwise the ability's controller may cast the
// card from its owner's exile without paying its mana cost. The cast follows
// the alternative-cost rules (CR 702.62d, 601.2b, 601.2f-h): additional
// costs, cost increases and reductions still apply, and a creature spell cast
// this way gains haste. If the card isn't cast it remains exiled.
func (e *Engine) resolveSuspendCastTrigger(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	card, ok := g.GetCardInstance(obj.SourceCardID)
	if !ok || card.ZoneVersion != obj.SourceZoneVersion {
		return "resolved"
	}
	owner, ok := playerByID(g, card.Owner)
	if !ok || !owner.Exile.Contains(card.ID) {
		return "resolved"
	}
	e.chooseAndCastFreeSpell(g, owner, obj.Controller, card.ID, zone.Exile, game.Selection{}, opt.V[int]{}, freeCastSpell{suspend: true}, true, agents, log)
	return "resolved"
}
