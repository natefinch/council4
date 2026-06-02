package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/rules/payment"
)

func madnessCostForCard(card *game.CardDef) (mana.Cost, bool) {
	for i := range card.Abilities {
		ability := &card.Abilities[i]
		if abilityHasKeyword(ability, game.Madness) && ability.MadnessCost.Exists {
			return ability.MadnessCost.Val, true
		}
	}
	return nil, false
}

func (e *Engine) resolveMadnessTriggeredAbilityWithChoices(g *game.Game, obj *game.StackObject, ability *game.AbilityDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	cardID := obj.SourceID
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return "missing source"
	}
	player, ok := playerByID(g, obj.Controller)
	if !ok || !player.Exile.Contains(cardID) {
		return "resolved"
	}
	if !e.chooseMay(g, agents, obj.Controller, "Cast card for its madness cost?", log) {
		moveExiledCardToGraveyard(g, obj.Controller, cardID)
		return "declined"
	}
	if !e.castMadnessSpellWithChoices(g, obj.Controller, card, ability.MadnessCost.Val, agents, log) {
		moveExiledCardToGraveyard(g, obj.Controller, cardID)
		return "resolved"
	}
	return "resolved"
}

func (e *Engine) castMadnessSpellWithChoices(g *game.Game, playerID game.PlayerID, card *game.CardInstance, cost mana.Cost, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Exile.Contains(card.ID) {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	modes, targets, ok := firstLegalSpellCastChoice(g, playerID, spellDef)
	if !ok {
		return false
	}
	prefs := e.paymentPreferencesForCost(g, playerID, &cost, nil, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &cost, Prefs: prefs}) {
		return false
	}
	if !player.Exile.Remove(card.ID) {
		return false
	}
	stackObj := &game.StackObject{
		ID:          g.IDGen.Next(),
		Kind:        game.StackSpell,
		SourceID:    card.ID,
		Face:        game.FaceFront,
		Controller:  playerID,
		Targets:     append([]game.Target(nil), targets...),
		ChosenModes: append([]int(nil), modes...),
	}
	pushSpellToStack(g, stackObj, game.GameEvent{
		SourceID:      card.ID,
		StackObjectID: stackObj.ID,
		Controller:    playerID,
		CardID:        card.ID,
		CardTypes:     cardTypes(spellDef),
		FromZone:      game.ZoneExile,
		ToZone:        game.ZoneStack,
	})
	return true
}

func firstLegalSpellCastChoice(g *game.Game, playerID game.PlayerID, spellDef *game.CardDef) ([]int, []game.Target, bool) {
	if !isSupportedSpell(spellDef) {
		return nil, nil, false
	}
	for _, modes := range modeChoicesForSpell(spellDef) {
		targetResult := targetChoicesForSpell(g, playerID, spellDef, modes)
		if targetResult.kind == targetInvalidSpec {
			continue
		}
		for _, targets := range targetResult.choices {
			if modesValidForSpell(spellDef, modes) && targetsValidForSpell(g, playerID, spellDef, modes, targets) {
				return append([]int(nil), modes...), append([]game.Target(nil), targets...), true
			}
		}
	}
	return nil, nil, false
}

func moveExiledCardToGraveyard(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Exile.Remove(cardID) {
		return false
	}
	player.Graveyard.Add(cardID)
	emitZoneChangeEvent(g, game.GameEvent{
		Player:   playerID,
		CardID:   cardID,
		FromZone: game.ZoneExile,
		ToZone:   game.ZoneGraveyard,
	})
	return true
}
