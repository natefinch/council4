package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
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
	player := g.Players[playerID]
	card, _ := g.GetCardInstance(cardID)
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	manaCost, counters, _ := suspendCostForCard(spellDef)
	prefs := e.paymentPreferencesForCost(g, playerID, &manaCost, nil, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &manaCost, Prefs: prefs}) {
		return false
	}
	if !player.Hand.Remove(cardID) {
		return false
	}
	player.Exile.Add(cardID)
	if g.SuspendedCards == nil {
		g.SuspendedCards = make(map[id.ID]game.SuspendedCard)
	}
	g.SuspendedCards[cardID] = game.SuspendedCard{
		Owner:        card.Owner,
		Controller:   playerID,
		TimeCounters: counters,
	}
	emitZoneChangeEvent(g, game.Event{
		Controller: playerID,
		Player:     card.Owner,
		CardID:     cardID,
		FromZone:   zone.Hand,
		ToZone:     zone.Exile,
	})
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

func (e *Engine) processSuspendUpkeep(g *game.Game, playerID game.PlayerID) {
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
		e.castSuspendedCard(g, playerID, cardID)
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

func (*Engine) castSuspendedCard(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Exile.Contains(cardID) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	modes, targets, ok := firstLegalSpellCastChoice(g, playerID, spellDef)
	if !ok {
		return false
	}
	targetCounts, ok := spellTargetCounts(g, playerID, spellDef, modes, targets)
	if !ok {
		panic("validated suspended spell targets could not be segmented")
	}
	if !player.Exile.Remove(cardID) {
		return false
	}
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     cardID,
		Face:         game.FaceFront,
		Controller:   playerID,
		Targets:      append([]game.Target(nil), targets...),
		TargetCounts: targetCounts,
		ChosenModes:  append([]int(nil), modes...),
		Suspend:      true,
	}
	pushSpellToStack(g, obj, game.Event{
		SourceID:      cardID,
		StackObjectID: obj.ID,
		Controller:    playerID,
		CardID:        cardID,
		CardTypes:     cardTypes(spellDef),
		FromZone:      zone.Exile,
		ToZone:        zone.Stack,
	})
	return true
}
